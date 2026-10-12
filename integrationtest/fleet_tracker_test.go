// Copyright Honeycomb.io
// SPDX-License-Identifier: Apache-2.0

//go:build integration

// Package integrationtest wires the redis_fleet_tracker extension to the
// adaptive_tail_sampling processor that consumes it. The two are deliberately
// decoupled, the processor resolves the extension by method shape rather than
// by importing it, so nothing in either module's build catches a change that
// breaks the pairing. This module exists to catch exactly that, and it lives
// outside the extension module so the extension never depends on contrib.
package integrationtest

import (
	"context"
	"os"
	"testing"
	"time"

	ats "github.com/open-telemetry/opentelemetry-collector-contrib/processor/adaptivetailsamplingprocessor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	rft "github.com/honeycombio/opentelemetry-collector-sampling-extensions/extension/redisfleettracker"
)

func redisEndpoint() string {
	if e := os.Getenv("REDIS_ENDPOINT"); e != "" {
		return e
	}
	return "localhost:6379"
}

// hostWithExtensions is the slice of component.Host the processor uses to
// resolve its fleet_tracker.
type hostWithExtensions struct {
	exts map[component.ID]component.Component
}

func (h hostWithExtensions) GetExtensions() map[component.ID]component.Component {
	return h.exts
}

// startTracker builds and starts a real extension with its own
// service.instance.id, so each one registers as a distinct fleet member.
func startTracker(t *testing.T, id component.ID, instanceID, fleet string) extension.Extension {
	t.Helper()

	f := rft.NewFactory()
	cfg := f.CreateDefaultConfig().(*rft.Config)
	cfg.Endpoint = redisEndpoint()
	cfg.Fleet = fleet
	cfg.HeartbeatInterval = 200 * time.Millisecond
	cfg.TTL = 800 * time.Millisecond

	set := extension.Settings{
		ID:                id,
		TelemetrySettings: componenttest.NewNopTelemetrySettings(),
	}
	set.Resource = pcommon.NewResource()
	set.Resource.Attributes().PutStr("service.instance.id", instanceID)

	ext, err := f.Create(context.Background(), set, cfg)
	require.NoError(t, err)
	require.NoError(t, ext.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() { _ = ext.Shutdown(context.Background()) })
	return ext
}

// TestFleetTracker_ProcessorDividesGoalByLiveMembers is the cross-component
// check: the processor resolves the extension structurally, then divides its
// configured goal_throughput by the member count the extension reports as
// instances join and leave.
func TestFleetTracker_ProcessorDividesGoalByLiveMembers(t *testing.T) {
	fleet := "integration-" + time.Now().Format("20060102150405.000000")
	trackerID := component.NewID(rft.NewFactory().Type())

	// The processor logs the goal it applies, which is the only place the
	// divided value is observable from outside the package.
	core, logs := observer.New(zapcore.DebugLevel)

	trackerA := startTracker(t, trackerID, "instance-a", fleet)

	pf := ats.NewFactory()
	cfg := pf.CreateDefaultConfig().(*ats.Config)
	cfg.DecisionDelay = 10 * time.Millisecond
	cfg.TraceTimeout = time.Hour
	cfg.FleetTrackerID = &trackerID
	cfg.Rules = []ats.RuleConfig{{
		Name: "throughput-cap",
		Sampler: ats.SamplerConfig{
			Type:           ats.AdaptiveThroughput,
			GoalThroughput: 1000,
		},
	}}

	pset := processor.Settings{
		ID:                component.NewID(pf.Type()),
		TelemetrySettings: componenttest.NewNopTelemetrySettings(),
		BuildInfo:         component.NewDefaultBuildInfo(),
	}
	pset.TelemetrySettings.Logger = zap.New(core)

	p, err := pf.CreateTraces(context.Background(), pset, cfg, consumertest.NewNop())
	require.NoError(t, err)

	host := hostWithExtensions{exts: map[component.ID]component.Component{
		trackerID: trackerA,
	}}
	// A Start error here means the processor could not resolve the extension
	// against its fleet_tracker contract.
	require.NoError(t, p.Start(context.Background(), host))
	t.Cleanup(func() { require.NoError(t, p.Shutdown(context.Background())) })

	// The processor starts at N=1, so reporting 1 is a no-change and no goal
	// is applied. Nothing to assert yet beyond the absence of a change.
	assert.Empty(t, appliedGoals(logs), "no goal should be applied while N is unchanged at 1")

	// A second instance joins: 1000 / 2.
	trackerB := startTracker(t, trackerID, "instance-b", fleet)
	assert.Eventually(t, func() bool {
		return lastAppliedGoal(logs) == 500
	}, 15*time.Second, 100*time.Millisecond,
		"expected the goal to halve when a second member joined, saw %v", appliedGoals(logs))

	// It leaves again, deregistering on shutdown: back to 1000 / 1.
	require.NoError(t, trackerB.Shutdown(context.Background()))
	assert.Eventually(t, func() bool {
		return lastAppliedGoal(logs) == 1000
	}, 15*time.Second, 100*time.Millisecond,
		"expected the goal to return to 1000 when the member left, saw %v", appliedGoals(logs))

	t.Logf("applied goal sequence: %v", appliedGoals(logs))
}

// appliedGoals returns the goal from every "applied fleet-divided goal" log
// line the processor has emitted, in order.
func appliedGoals(logs *observer.ObservedLogs) []int64 {
	var out []int64
	for _, e := range logs.FilterMessage("applied fleet-divided goal").All() {
		for _, f := range e.Context {
			if f.Key == "goal_per_sec" {
				out = append(out, f.Integer)
			}
		}
	}
	return out
}

func lastAppliedGoal(logs *observer.ObservedLogs) int64 {
	g := appliedGoals(logs)
	if len(g) == 0 {
		return -1
	}
	return g[len(g)-1]
}
