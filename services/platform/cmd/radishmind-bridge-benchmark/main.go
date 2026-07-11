package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"radishmind.local/services/platform/internal/bridge"
)

const (
	defaultRequestFile       = "datasets/examples/radishflow-copilot-request-ghost-valve-ambiguous-001.json"
	defaultSequentialCount   = 20
	defaultConcurrentCount   = 20
	defaultConcurrency       = 4
	defaultWarmupCount       = 2
	defaultPerRequestTimeout = 30 * time.Second
)

type benchmarkConfig struct {
	requestFile       string
	pythonBinary      string
	scriptPath        string
	sequentialCount   int
	concurrentCount   int
	concurrency       int
	warmupCount       int
	perRequestTimeout time.Duration
}

type bridgeMeasurement struct {
	totalMS         float64
	processAndIPCMS float64
	pythonGatewayMS float64
	providerMS      float64
}

type durationStats struct {
	MinimumMS float64 `json:"minimum_ms"`
	MeanMS    float64 `json:"mean_ms"`
	P50MS     float64 `json:"p50_ms"`
	P95MS     float64 `json:"p95_ms"`
	MaximumMS float64 `json:"maximum_ms"`
}

type phaseSummary struct {
	RequestCount                int           `json:"request_count"`
	Concurrency                 int           `json:"concurrency"`
	WallTimeMS                  float64       `json:"wall_time_ms"`
	ThroughputRequestsPerSecond float64       `json:"throughput_requests_per_second"`
	Total                       durationStats `json:"total"`
	ProcessAndIPC               durationStats `json:"process_and_ipc"`
	PythonGateway               durationStats `json:"python_gateway"`
	Provider                    durationStats `json:"provider"`
}

type benchmarkOutput struct {
	SchemaVersion         int          `json:"schema_version"`
	Status                string       `json:"status"`
	MeasuredAt            string       `json:"measured_at"`
	Mode                  string       `json:"mode"`
	Provider              string       `json:"provider"`
	RequestFixture        string       `json:"request_fixture"`
	PercentileMethod      string       `json:"percentile_method"`
	WarmupRequests        int          `json:"warmup_requests"`
	ExpectedProcessStarts int          `json:"expected_process_starts"`
	Runtime               runtimeInfo  `json:"runtime"`
	Sequential            phaseSummary `json:"sequential"`
	Concurrent            phaseSummary `json:"concurrent"`
	Sanitized             bool         `json:"sanitized"`
}

type runtimeInfo struct {
	GoVersion     string `json:"go_version"`
	Platform      string `json:"platform"`
	PythonRuntime string `json:"python_runtime"`
}

type phaseResult struct {
	measurement bridgeMeasurement
	err         error
}

func main() {
	config, err := parseBenchmarkConfig(os.Args[1:])
	if err != nil {
		fail(err)
	}
	repoRoot, err := findRepoRoot()
	if err != nil {
		fail(err)
	}
	requestDocument, requestLabel, err := loadBenchmarkRequest(repoRoot, config.requestFile)
	if err != nil {
		fail(err)
	}
	if config.pythonBinary == "" {
		config.pythonBinary = filepath.Join(repoRoot, ".venv", "bin", "python")
	}
	if config.scriptPath == "" {
		config.scriptPath = filepath.Join(repoRoot, "scripts", "run-platform-bridge.py")
	}
	config.pythonBinary, err = resolvePythonBinary(config.pythonBinary)
	if err != nil {
		fail(err)
	}

	client := bridge.NewClient(config.pythonBinary, config.scriptPath)
	options := bridge.EnvelopeOptions{
		Provider:       "mock",
		RequestTimeout: config.perRequestTimeout,
	}
	for index := 0; index < config.warmupCount; index++ {
		payload, err := benchmarkRequestPayload(requestDocument, fmt.Sprintf("warmup-%03d", index+1))
		if err != nil {
			fail(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), config.perRequestTimeout)
		_, err = measureBridgeRequest(ctx, client, payload, options)
		cancel()
		if err != nil {
			fail(err)
		}
	}

	sequentialPayloads, err := benchmarkRequestPayloads(requestDocument, "sequential", config.sequentialCount)
	if err != nil {
		fail(err)
	}
	concurrentPayloads, err := benchmarkRequestPayloads(requestDocument, "concurrent", config.concurrentCount)
	if err != nil {
		fail(err)
	}
	sequential, err := runBenchmarkPhase(client, sequentialPayloads, options, config.perRequestTimeout, 1)
	if err != nil {
		fail(err)
	}
	concurrent, err := runBenchmarkPhase(
		client,
		concurrentPayloads,
		options,
		config.perRequestTimeout,
		config.concurrency,
	)
	if err != nil {
		fail(err)
	}

	output := benchmarkOutput{
		SchemaVersion:         1,
		Status:                "ok",
		MeasuredAt:            time.Now().UTC().Format(time.RFC3339),
		Mode:                  "process_per_request",
		Provider:              "mock",
		RequestFixture:        requestLabel,
		PercentileMethod:      "nearest_rank",
		WarmupRequests:        config.warmupCount,
		ExpectedProcessStarts: config.warmupCount + config.sequentialCount + config.concurrentCount,
		Runtime: runtimeInfo{
			GoVersion:     runtime.Version(),
			Platform:      runtime.GOOS + "/" + runtime.GOARCH,
			PythonRuntime: sanitizedPythonRuntime(config.pythonBinary, repoRoot),
		},
		Sequential: sequential,
		Concurrent: concurrent,
		Sanitized:  true,
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		fail(errors.New("encode bridge benchmark output failed"))
	}
}

func parseBenchmarkConfig(arguments []string) (benchmarkConfig, error) {
	config := benchmarkConfig{}
	flags := flag.NewFlagSet("radishmind-bridge-benchmark", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.StringVar(&config.requestFile, "request-file", defaultRequestFile, "repository-relative canonical request fixture")
	flags.StringVar(&config.pythonBinary, "python", "", "Python executable; defaults to the repository .venv")
	flags.StringVar(&config.scriptPath, "script", "", "platform bridge script; defaults to the repository bridge")
	flags.IntVar(&config.sequentialCount, "sequential-requests", defaultSequentialCount, "measured sequential request count")
	flags.IntVar(&config.concurrentCount, "concurrent-requests", defaultConcurrentCount, "measured concurrent request count")
	flags.IntVar(&config.concurrency, "concurrency", defaultConcurrency, "concurrent worker count")
	flags.IntVar(&config.warmupCount, "warmup-requests", defaultWarmupCount, "warmup request count")
	flags.DurationVar(&config.perRequestTimeout, "request-timeout", defaultPerRequestTimeout, "per-request timeout")
	if err := flags.Parse(arguments); err != nil {
		return benchmarkConfig{}, errors.New("parse bridge benchmark arguments failed")
	}
	if flags.NArg() != 0 {
		return benchmarkConfig{}, errors.New("bridge benchmark does not accept positional arguments")
	}
	if config.sequentialCount <= 0 || config.concurrentCount <= 0 {
		return benchmarkConfig{}, errors.New("sequential and concurrent request counts must be positive")
	}
	if config.sequentialCount > 1000 || config.concurrentCount > 1000 {
		return benchmarkConfig{}, errors.New("request counts must not exceed 1000")
	}
	if config.concurrency <= 0 || config.concurrency > 64 {
		return benchmarkConfig{}, errors.New("concurrency must be between 1 and 64")
	}
	if config.warmupCount < 0 || config.warmupCount > 100 {
		return benchmarkConfig{}, errors.New("warmup request count must be between 0 and 100")
	}
	if config.perRequestTimeout <= 0 {
		return benchmarkConfig{}, errors.New("request timeout must be positive")
	}
	return config, nil
}

func runBenchmarkPhase(
	client *bridge.Client,
	payloads [][]byte,
	options bridge.EnvelopeOptions,
	timeout time.Duration,
	concurrency int,
) (phaseSummary, error) {
	if len(payloads) == 0 {
		return phaseSummary{}, errors.New("benchmark phase requires at least one request")
	}
	if concurrency > len(payloads) {
		concurrency = len(payloads)
	}
	jobs := make(chan []byte, len(payloads))
	results := make(chan phaseResult, len(payloads))
	for _, payload := range payloads {
		jobs <- payload
	}
	close(jobs)

	startedAt := time.Now()
	var workers sync.WaitGroup
	for workerIndex := 0; workerIndex < concurrency; workerIndex++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for payload := range jobs {
				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				measurement, err := measureBridgeRequest(ctx, client, payload, options)
				cancel()
				results <- phaseResult{measurement: measurement, err: err}
			}
		}()
	}
	workers.Wait()
	close(results)
	wallTime := time.Since(startedAt)

	measurements := make([]bridgeMeasurement, 0, len(payloads))
	for result := range results {
		if result.err != nil {
			return phaseSummary{}, result.err
		}
		measurements = append(measurements, result.measurement)
	}
	return summarizePhase(measurements, concurrency, wallTime), nil
}

func measureBridgeRequest(
	ctx context.Context,
	client *bridge.Client,
	payload []byte,
	options bridge.EnvelopeOptions,
) (bridgeMeasurement, error) {
	startedAt := time.Now()
	envelope, err := client.HandleEnvelope(ctx, payload, options)
	totalMS := float64(time.Since(startedAt).Nanoseconds()) / float64(time.Millisecond)
	if err != nil {
		return bridgeMeasurement{}, errors.New("bridge benchmark request failed")
	}
	if strings.EqualFold(envelope.Status, "failed") {
		return bridgeMeasurement{}, errors.New("bridge benchmark received a failed Gateway envelope")
	}
	pythonDurationMS, ok := metadataMilliseconds(envelope.Metadata, "duration_ms")
	if !ok {
		return bridgeMeasurement{}, errors.New("Gateway envelope is missing duration_ms")
	}
	providerDurationMS, ok := metadataMilliseconds(envelope.Metadata, "provider_duration_ms")
	if !ok {
		return bridgeMeasurement{}, errors.New("Gateway envelope is missing provider_duration_ms")
	}
	return bridgeMeasurement{
		totalMS:         totalMS,
		processAndIPCMS: nonNegative(totalMS - pythonDurationMS),
		pythonGatewayMS: nonNegative(pythonDurationMS - providerDurationMS),
		providerMS:      providerDurationMS,
	}, nil
}

func summarizePhase(measurements []bridgeMeasurement, concurrency int, wallTime time.Duration) phaseSummary {
	totalValues := make([]float64, 0, len(measurements))
	processValues := make([]float64, 0, len(measurements))
	pythonValues := make([]float64, 0, len(measurements))
	providerValues := make([]float64, 0, len(measurements))
	for _, measurement := range measurements {
		totalValues = append(totalValues, measurement.totalMS)
		processValues = append(processValues, measurement.processAndIPCMS)
		pythonValues = append(pythonValues, measurement.pythonGatewayMS)
		providerValues = append(providerValues, measurement.providerMS)
	}
	throughput := float64(len(measurements)) / wallTime.Seconds()
	return phaseSummary{
		RequestCount:                len(measurements),
		Concurrency:                 concurrency,
		WallTimeMS:                  roundMilliseconds(float64(wallTime.Nanoseconds()) / float64(time.Millisecond)),
		ThroughputRequestsPerSecond: roundMilliseconds(throughput),
		Total:                       summarizeDurations(totalValues),
		ProcessAndIPC:               summarizeDurations(processValues),
		PythonGateway:               summarizeDurations(pythonValues),
		Provider:                    summarizeDurations(providerValues),
	}
}

func summarizeDurations(values []float64) durationStats {
	if len(values) == 0 {
		return durationStats{}
	}
	sortedValues := append([]float64(nil), values...)
	sort.Float64s(sortedValues)
	total := 0.0
	for _, value := range sortedValues {
		total += value
	}
	return durationStats{
		MinimumMS: roundMilliseconds(sortedValues[0]),
		MeanMS:    roundMilliseconds(total / float64(len(sortedValues))),
		P50MS:     roundMilliseconds(nearestRank(sortedValues, 0.50)),
		P95MS:     roundMilliseconds(nearestRank(sortedValues, 0.95)),
		MaximumMS: roundMilliseconds(sortedValues[len(sortedValues)-1]),
	}
}

func nearestRank(sortedValues []float64, percentile float64) float64 {
	if len(sortedValues) == 0 {
		return 0
	}
	rank := int(math.Ceil(percentile*float64(len(sortedValues)))) - 1
	if rank < 0 {
		rank = 0
	}
	if rank >= len(sortedValues) {
		rank = len(sortedValues) - 1
	}
	return sortedValues[rank]
}

func metadataMilliseconds(metadata map[string]any, key string) (float64, bool) {
	value, ok := metadata[key]
	if !ok {
		return 0, false
	}
	switch typed := value.(type) {
	case float64:
		return nonNegative(typed), true
	case float32:
		return nonNegative(float64(typed)), true
	case int:
		return nonNegative(float64(typed)), true
	case int64:
		return nonNegative(float64(typed)), true
	case json.Number:
		parsed, err := typed.Float64()
		return nonNegative(parsed), err == nil
	default:
		return 0, false
	}
}

func benchmarkRequestPayloads(document map[string]any, prefix string, count int) ([][]byte, error) {
	payloads := make([][]byte, 0, count)
	for index := 0; index < count; index++ {
		payload, err := benchmarkRequestPayload(document, fmt.Sprintf("%s-%03d", prefix, index+1))
		if err != nil {
			return nil, err
		}
		payloads = append(payloads, payload)
	}
	return payloads, nil
}

func benchmarkRequestPayload(document map[string]any, suffix string) ([]byte, error) {
	copyDocument := make(map[string]any, len(document))
	for key, value := range document {
		copyDocument[key] = value
	}
	copyDocument["request_id"] = "gateway-bridge-benchmark-" + suffix
	payload, err := json.Marshal(copyDocument)
	if err != nil {
		return nil, errors.New("encode bridge benchmark request failed")
	}
	return payload, nil
}

func loadBenchmarkRequest(repoRoot string, relativePath string) (map[string]any, string, error) {
	normalizedPath := filepath.Clean(strings.TrimSpace(relativePath))
	if normalizedPath == "." || filepath.IsAbs(normalizedPath) || strings.HasPrefix(normalizedPath, ".."+string(filepath.Separator)) {
		return nil, "", errors.New("request file must be a repository-relative path")
	}
	content, err := os.ReadFile(filepath.Join(repoRoot, normalizedPath))
	if err != nil {
		return nil, "", errors.New("read bridge benchmark request fixture failed")
	}
	document := map[string]any{}
	if err := json.Unmarshal(content, &document); err != nil {
		return nil, "", errors.New("decode bridge benchmark request fixture failed")
	}
	return document, filepath.ToSlash(normalizedPath), nil
}

func findRepoRoot() (string, error) {
	currentDirectory, err := os.Getwd()
	if err != nil {
		return "", errors.New("resolve current directory failed")
	}
	for {
		goModule := filepath.Join(currentDirectory, "services", "platform", "go.mod")
		bridgeScript := filepath.Join(currentDirectory, "scripts", "run-platform-bridge.py")
		if _, goModuleErr := os.Stat(goModule); goModuleErr == nil {
			if _, bridgeErr := os.Stat(bridgeScript); bridgeErr == nil {
				return currentDirectory, nil
			}
		}
		parentDirectory := filepath.Dir(currentDirectory)
		if parentDirectory == currentDirectory {
			break
		}
		currentDirectory = parentDirectory
	}
	return "", errors.New("RadishMind repository root was not found")
}

func resolvePythonBinary(pythonBinary string) (string, error) {
	normalizedBinary := strings.TrimSpace(pythonBinary)
	if normalizedBinary == "" {
		return "", errors.New("Python executable is unavailable")
	}
	if filepath.IsAbs(normalizedBinary) || strings.ContainsAny(normalizedBinary, `/\`) {
		fileInfo, err := os.Stat(normalizedBinary)
		if err != nil || fileInfo.IsDir() {
			return "", errors.New("Python executable is unavailable; run ./scripts/bootstrap-dev.sh")
		}
		return normalizedBinary, nil
	}
	resolvedBinary, err := exec.LookPath(normalizedBinary)
	if err != nil {
		return "", errors.New("Python executable is unavailable")
	}
	return resolvedBinary, nil
}

func sanitizedPythonRuntime(pythonBinary string, repoRoot string) string {
	repositoryVenv := filepath.Join(repoRoot, ".venv", "bin", "python")
	if filepath.Clean(pythonBinary) == filepath.Clean(repositoryVenv) {
		return "repository_venv"
	}
	return filepath.Base(pythonBinary)
}

func nonNegative(value float64) float64 {
	if value < 0 {
		return 0
	}
	return value
}

func roundMilliseconds(value float64) float64 {
	return math.Round(value*1000) / 1000
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "bridge benchmark failed: %s\n", strings.TrimSpace(err.Error()))
	os.Exit(1)
}
