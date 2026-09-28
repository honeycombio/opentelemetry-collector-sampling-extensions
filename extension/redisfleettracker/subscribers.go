// Copyright Honeycomb.io
// SPDX-License-Identifier: Apache-2.0

package redisfleettracker // import "github.com/honeycombio/opentelemetry-collector-sampling-extensions/extension/redisfleettracker"

// fleetTracker mirrors the consumer-side contract the adaptive_tail_sampling
// processor asserts structurally via host.GetExtensions (contrib #50577).
type fleetTracker interface {
	SubscribeMemberCount(callback func(count int)) (cancel func(), err error)
}

var _ fleetTracker = (*fleetTrackerExtension)(nil)
