// Copyright Honeycomb.io
// SPDX-License-Identifier: Apache-2.0

// Package redisfleettracker backs a fleet membership count with Redis, so
// collector instances that need to know how many peers are alive (for example
// the adaptive_tail_sampling processor's per-key rendezvous routing) can
// subscribe to a live member count without running their own coordination
// protocol.
package redisfleettracker // import "github.com/honeycombio/opentelemetry-collector-sampling-extensions/extension/redisfleettracker"

//go:generate go tool mdatagen metadata.yaml
