// Copyright Honeycomb.io
// SPDX-License-Identifier: Apache-2.0

package redissamplingstateextension // import "github.com/honeycombio/opentelemetry-collector-samplingstate/redissamplingstateextension"

import (
	"errors"
	"time"

	"go.opentelemetry.io/collector/config/configopaque"
)

// Config configures the redis_sampling_state extension.
type Config struct {
	// Endpoint is the Redis server address as host:port.
	Endpoint string `mapstructure:"endpoint"`
	// Password authenticates to Redis. Optional.
	Password configopaque.String `mapstructure:"password"`
	// DB is the Redis logical database number.
	DB int `mapstructure:"db"`
	// KeyPrefix namespaces every key this extension writes. All collector
	// instances sharing state must use the same prefix.
	KeyPrefix string `mapstructure:"key_prefix"`
	// BucketTTL is how long an interval bucket survives in Redis after its
	// last write. It only needs to outlive the sampler's adjustment interval
	// by a small margin; the default of 10m covers any sensible interval.
	BucketTTL time.Duration `mapstructure:"bucket_ttl"`
}

func (cfg *Config) Validate() error {
	if cfg.Endpoint == "" {
		return errors.New("endpoint is required")
	}
	if cfg.BucketTTL <= 0 {
		return errors.New("bucket_ttl must be greater than zero")
	}
	return nil
}
