package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"radishmind.local/services/platform/internal/httpapi"
)

const workflowRAGEvaluationMaximumFileBytes = 4 * 1024 * 1024

type workflowRAGEvaluationConfig struct {
	snapshotPath string
	datasetPath  string
	outputPath   string
	check        bool
}

type workflowRAGEvaluationSummary struct {
	DatasetID      string `json:"dataset_id"`
	DatasetVersion int    `json:"dataset_version"`
	DatasetDigest  string `json:"dataset_digest"`
	SnapshotDigest string `json:"snapshot_digest"`
	ProfileDigest  string `json:"profile_digest"`
	SampleCount    int    `json:"sample_count"`
	Status         string `json:"status"`
	ReportPath     string `json:"report_path"`
}

func main() {
	if err := runWorkflowRAGEvaluation(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "workflow RAG evaluation failed: %v\n", err)
		os.Exit(1)
	}
}

func runWorkflowRAGEvaluation(arguments []string, stdout, stderr io.Writer) error {
	config, err := parseWorkflowRAGEvaluationConfig(arguments, stderr)
	if err != nil {
		return err
	}
	snapshotPayload, err := readWorkflowRAGEvaluationFile(config.snapshotPath)
	if err != nil {
		return errors.New("read snapshot failed")
	}
	datasetPayload, err := readWorkflowRAGEvaluationFile(config.datasetPath)
	if err != nil {
		return errors.New("read dataset failed")
	}
	review, err := httpapi.BuildWorkflowRAGQualityReview(snapshotPayload, datasetPayload)
	if err != nil {
		return err
	}
	reportPayload, err := json.MarshalIndent(review, "", "  ")
	if err != nil {
		return errors.New("encode quality review failed")
	}
	reportPayload = append(reportPayload, '\n')
	if config.check {
		committed, readErr := readWorkflowRAGEvaluationFile(config.outputPath)
		if readErr != nil {
			return errors.New("read committed quality review failed")
		}
		if _, decodeErr := httpapi.DecodeWorkflowRAGQualityReview(committed); decodeErr != nil {
			return errors.New("committed quality review contract is invalid")
		}
		if !bytes.Equal(committed, reportPayload) {
			return errors.New("committed quality review drifted")
		}
	} else if err = writeWorkflowRAGEvaluationFile(config.outputPath, reportPayload); err != nil {
		return errors.New("write quality review failed")
	}
	summary := workflowRAGEvaluationSummary{
		DatasetID: review.Dataset.DatasetID, DatasetVersion: review.Dataset.DatasetVersion, DatasetDigest: review.Dataset.DatasetDigest,
		SnapshotDigest: review.Snapshot.SnapshotDigest, ProfileDigest: review.Profile.ProfileDigest,
		SampleCount: review.SampleCount, Status: review.Status, ReportPath: filepath.Clean(config.outputPath),
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetEscapeHTML(false)
	if err = encoder.Encode(summary); err != nil {
		return errors.New("write evaluation summary failed")
	}
	if review.Status != "passed" {
		return fmt.Errorf("quality review status is %s", review.Status)
	}
	return nil
}

func parseWorkflowRAGEvaluationConfig(arguments []string, stderr io.Writer) (workflowRAGEvaluationConfig, error) {
	config := workflowRAGEvaluationConfig{}
	flags := flag.NewFlagSet("radishmind-workflow-rag-eval", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&config.snapshotPath, "snapshot", "", "path to one immutable synthetic-public workflow RAG snapshot")
	flags.StringVar(&config.datasetPath, "dataset", "", "path to one versioned workflow RAG evaluation dataset")
	flags.StringVar(&config.outputPath, "output", "", "path to the metadata-only quality review")
	flags.BoolVar(&config.check, "check", false, "rebuild and compare the committed report without writing")
	if err := flags.Parse(arguments); err != nil {
		return workflowRAGEvaluationConfig{}, errors.New("parse evaluation arguments failed")
	}
	if flags.NArg() != 0 || config.snapshotPath == "" || config.datasetPath == "" || config.outputPath == "" {
		return workflowRAGEvaluationConfig{}, errors.New("snapshot, dataset and output paths are required; positional arguments are not accepted")
	}
	if filepath.Clean(config.snapshotPath) == filepath.Clean(config.outputPath) || filepath.Clean(config.datasetPath) == filepath.Clean(config.outputPath) {
		return workflowRAGEvaluationConfig{}, errors.New("output path must differ from input paths")
	}
	return config, nil
}

func readWorkflowRAGEvaluationFile(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() < 1 || info.Size() > workflowRAGEvaluationMaximumFileBytes {
		return nil, errors.New("evaluation file is missing, not regular, empty or oversized")
	}
	payload, err := os.ReadFile(path)
	if err != nil || len(payload) < 1 || len(payload) > workflowRAGEvaluationMaximumFileBytes {
		return nil, errors.New("evaluation file changed while reading or exceeded its budget")
	}
	return payload, nil
}

func writeWorkflowRAGEvaluationFile(path string, payload []byte) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(directory, ".workflow-rag-review-*.tmp")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err = temporary.Chmod(0o644); err == nil {
		_, err = temporary.Write(payload)
	}
	if err == nil {
		err = temporary.Sync()
	}
	if closeErr := temporary.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}
