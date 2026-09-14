// Copyright Honeycomb.io
// SPDX-License-Identifier: Apache-2.0

package redissamplingstateextension // import "github.com/honeycombio/opentelemetry-collector-samplingstate/redissamplingstateextension"

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
)

// Type is the component type of the extension.
var Type = component.MustNewType("redis_sampling_state")

// NewFactory creates the redis_sampling_state extension factory.
func NewFactory() extension.Factory {
	return extension.NewFactory(
		Type,
		createDefaultConfig,
		create,
		component.StabilityLevelDevelopment,
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		KeyPrefix: "samplingstate",
		BucketTTL: 10 * time.Minute,
	}
}

func create(_ context.Context, set extension.Settings, cfg component.Config) (extension.Extension, error) {
	return newExtension(cfg.(*Config), set.Logger), nil
}
