// Copyright Honeycomb.io
// SPDX-License-Identifier: Apache-2.0

// Package redissamplingstateextension backs the adaptive_tail_sampling
// processor's shared_counters option with Redis, so collector instances share
// one throughput budget. It implements the processor's counter store contract
// (AddCounts/ReadCounts) structurally.
package redissamplingstateextension // import "github.com/honeycombio/opentelemetry-collector-samplingstate/redissamplingstateextension"

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

type samplingStateExtension struct {
	cfg    *Config
	logger *zap.Logger
	client *redis.Client
}

func newExtension(cfg *Config, logger *zap.Logger) *samplingStateExtension {
	return &samplingStateExtension{cfg: cfg, logger: logger}
}

// Start connects to Redis. The connection is verified with a ping so
// misconfiguration surfaces at startup rather than on the first sync tick.
func (e *samplingStateExtension) Start(ctx context.Context, _ component.Host) error {
	e.client = redis.NewClient(&redis.Options{
		Addr:     e.cfg.Endpoint,
		Password: string(e.cfg.Password),
		DB:       e.cfg.DB,
	})
	if err := e.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis_sampling_state: ping %s: %w", e.cfg.Endpoint, err)
	}
	return nil
}

// Shutdown closes the Redis client.
func (e *samplingStateExtension) Shutdown(context.Context) error {
	if e.client == nil {
		return nil
	}
	return e.client.Close()
}

// bucketKey builds the Redis key for one sampler's interval bucket. The
// samplerID is wrapped in braces as a Redis Cluster hash tag, so all buckets
// of one sampler land in the same slot and stay pipelineable under cluster
// deployments.
func (e *samplingStateExtension) bucketKey(samplerID string, bucket int64) string {
	return fmt.Sprintf("%s:{%s}:%d", e.cfg.KeyPrefix, samplerID, bucket)
}

// AddCounts implements the counter store contract. Each key's count is folded
// in with HINCRBYFLOAT, so concurrent writers merge additively without
// read-modify-write races. The bucket's TTL is refreshed on every write.
func (e *samplingStateExtension) AddCounts(ctx context.Context, samplerID string, bucket int64, counts map[string]float64) error {
	key := e.bucketKey(samplerID, bucket)
	pipe := e.client.Pipeline()
	for k, v := range counts {
		pipe.HIncrByFloat(ctx, key, k, v)
	}
	pipe.Expire(ctx, key, e.cfg.BucketTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("redis_sampling_state: add counts: %w", err)
	}
	return nil
}

// ReadCounts implements the counter store contract. A bucket nobody has
// written reads as an empty map, not an error.
func (e *samplingStateExtension) ReadCounts(ctx context.Context, samplerID string, bucket int64) (map[string]float64, error) {
	vals, err := e.client.HGetAll(ctx, e.bucketKey(samplerID, bucket)).Result()
	if err != nil {
		return nil, fmt.Errorf("redis_sampling_state: read counts: %w", err)
	}
	counts := make(map[string]float64, len(vals))
	for k, v := range vals {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("redis_sampling_state: non-numeric count for key %q: %w", k, err)
		}
		counts[k] = f
	}
	return counts, nil
}
