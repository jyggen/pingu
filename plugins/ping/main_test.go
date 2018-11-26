package main

import (
	"github.com/jyggen/pingu/slack"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	plugin := newPlugin()

	if n := plugin.Name(); n != name {
		t.Errorf("Name() was incorrect, got: %v, want %v.", n, name)
	}
}

func TestOnLatencyReport(t *testing.T) {
	plugin := newPlugin()
	latency, _ := time.ParseDuration("300ms")
	event := &slack.LatencyReportEvent{}

	plugin.OnLatencyReport(event)

	if l := plugin.Latency(); l != latency {
		t.Errorf("Latency() was incorrect, got: %v, want %v.", l, latency)
	}
}
