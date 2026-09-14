// Copyright Honeycomb.io
// SPDX-License-Identifier: Apache-2.0

package redissamplingstateextension

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/extension/extensiontest"
	"go.uber.org/zap"
)

func testExtension(t *testing.T) *samplingStateExtension {
	t.Helper()
	srv := miniredis.RunT(t)
	cfg := createDefaultConfig().(*Config)
	cfg.Endpoint = srv.Addr()
	require.NoError(t, cfg.Validate())
	e := newExtension(cfg, zap.NewNop())
	require.NoError(t, e.Start(t.Context(), componenttest.NewNopHost()))
	t.Cleanup(func() { require.NoError(t, e.Shutdown(t.Context())) })
	return e
}

func TestAddCounts_MergesAdditively(t *testing.T) {
	e := testExtension(t)
	ctx := t.Context()

	// Two instances write into the same sampler and bucket.
	require.NoError(t, e.AddCounts(ctx, "rule-a", 100, map[string]float64{"svc-1": 10.5, "svc-2": 5}))
	require.NoError(t, e.AddCounts(ctx, "rule-a", 100, map[string]float64{"svc-1": 7, "svc-3": 2}))

	counts, err := e.ReadCounts(ctx, "rule-a", 100)
	require.NoError(t, err)
	assert.Equal(t, map[string]float64{"svc-1": 17.5, "svc-2": 5, "svc-3": 2}, counts)
}

func TestBucketsAndSamplersAreIsolated(t *testing.T) {
	e := testExtension(t)
	ctx := t.Context()

	require.NoError(t, e.AddCounts(ctx, "rule-a", 100, map[string]float64{"svc-1": 10}))
	require.NoError(t, e.AddCounts(ctx, "rule-a", 101, map[string]float64{"svc-1": 3}))
	require.NoError(t, e.AddCounts(ctx, "rule-b", 100, map[string]float64{"svc-1": 99}))

	counts, err := e.ReadCounts(ctx, "rule-a", 100)
	require.NoError(t, err)
	assert.Equal(t, map[string]float64{"svc-1": 10}, counts)
	counts, err = e.ReadCounts(ctx, "rule-a", 101)
	require.NoError(t, err)
	assert.Equal(t, map[string]float64{"svc-1": 3}, counts)
}

func TestReadCounts_UnwrittenBucketIsEmpty(t *testing.T) {
	e := testExtension(t)
	counts, err := e.ReadCounts(t.Context(), "rule-a", 42)
	require.NoError(t, err)
	assert.Empty(t, counts)
	assert.NotNil(t, counts)
}

func TestAddCounts_SetsBucketTTL(t *testing.T) {
	srv := miniredis.RunT(t)
	cfg := createDefaultConfig().(*Config)
	cfg.Endpoint = srv.Addr()
	cfg.BucketTTL = time.Minute
	e := newExtension(cfg, zap.NewNop())
	require.NoError(t, e.Start(t.Context(), componenttest.NewNopHost()))
	t.Cleanup(func() { require.NoError(t, e.Shutdown(t.Context())) })

	require.NoError(t, e.AddCounts(t.Context(), "rule-a", 100, map[string]float64{"svc-1": 1}))
	assert.Positive(t, srv.TTL(e.bucketKey("rule-a", 100)), "bucket keys must expire")

	srv.FastForward(2 * time.Minute)
	counts, err := e.ReadCounts(t.Context(), "rule-a", 100)
	require.NoError(t, err)
	assert.Empty(t, counts, "expired buckets read as empty")
}

func TestStart_UnreachableEndpoint(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	cfg.Endpoint = "127.0.0.1:1"
	e := newExtension(cfg, zap.NewNop())
	assert.Error(t, e.Start(t.Context(), componenttest.NewNopHost()))
	require.NoError(t, e.Shutdown(t.Context()))
}

func TestFactory(t *testing.T) {
	f := NewFactory()
	assert.Equal(t, Type, f.Type())
	cfg := f.CreateDefaultConfig().(*Config)
	assert.Equal(t, "samplingstate", cfg.KeyPrefix)
	assert.Equal(t, 10*time.Minute, cfg.BucketTTL)
	assert.ErrorContains(t, cfg.Validate(), "endpoint", "default config must not validate without an endpoint")

	cfg.Endpoint = "localhost:6379"
	ext, err := f.Create(t.Context(), extensiontest.NewNopSettings(Type), cfg)
	require.NoError(t, err)
	assert.NotNil(t, ext)
}

func TestConfig_Validate(t *testing.T) {
	cfg := &Config{Endpoint: "localhost:6379", BucketTTL: 0}
	assert.ErrorContains(t, cfg.Validate(), "bucket_ttl")
}
