package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSummarizeDurationsUsesNearestRank(t *testing.T) {
	stats := summarizeDurations([]float64{9, 1, 5, 3, 7, 2, 4, 6, 8, 10})
	if stats.MinimumMS != 1 || stats.MaximumMS != 10 || stats.MeanMS != 5.5 {
		t.Fatalf("unexpected summary bounds: %#v", stats)
	}
	if stats.P50MS != 5 || stats.P95MS != 10 {
		t.Fatalf("unexpected nearest-rank percentiles: %#v", stats)
	}
}

func TestSummarizePhaseSeparatesTimingSegments(t *testing.T) {
	measurements := []bridgeMeasurement{
		{totalMS: 100, processAndIPCMS: 70, pythonGatewayMS: 20, providerMS: 10},
		{totalMS: 120, processAndIPCMS: 80, pythonGatewayMS: 25, providerMS: 15},
	}
	summary := summarizePhase(measurements, 2, 50*time.Millisecond)
	if summary.RequestCount != 2 || summary.Concurrency != 2 {
		t.Fatalf("unexpected phase shape: %#v", summary)
	}
	if summary.Total.P50MS != 100 || summary.ProcessAndIPC.P95MS != 80 || summary.Provider.P95MS != 15 {
		t.Fatalf("timing segments drifted: %#v", summary)
	}
	if summary.ThroughputRequestsPerSecond != 40 {
		t.Fatalf("unexpected throughput: %#v", summary)
	}
}

func TestMetadataMillisecondsAcceptsDecodedJSONNumbers(t *testing.T) {
	metadata := map[string]any{"duration_ms": float64(12), "provider_duration_ms": json.Number("3")}
	if value, ok := metadataMilliseconds(metadata, "duration_ms"); !ok || value != 12 {
		t.Fatalf("float metadata was not accepted: value=%v ok=%v", value, ok)
	}
	if value, ok := metadataMilliseconds(metadata, "provider_duration_ms"); !ok || value != 3 {
		t.Fatalf("json number metadata was not accepted: value=%v ok=%v", value, ok)
	}
}

func TestParseBenchmarkConfigRejectsUnsafeCounts(t *testing.T) {
	if _, err := parseBenchmarkConfig([]string{"--concurrency", "0"}); err == nil {
		t.Fatal("zero concurrency must be rejected")
	}
	if _, err := parseBenchmarkConfig([]string{"--sequential-requests", "1001"}); err == nil {
		t.Fatal("unbounded request count must be rejected")
	}
}
