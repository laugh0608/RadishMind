package developmentprobe

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const maxSourceBytes = 32768

type Inspection struct {
	ProbeID         string       `json:"probe_id"`
	BaseDigest      string       `json:"base_digest"`
	CandidateDigest string       `json:"candidate_digest"`
	Files           []SourceFile `json:"files"`
	ChangedPaths    []string     `json:"changed_paths"`
}

func TreeDigest(files []SourceFile) string {
	var content strings.Builder
	for _, file := range files {
		content.WriteString(file.Path + "\x00" + file.SHA256 + "\n")
	}
	return Digest([]byte(content.String()))
}

func ReadSourceFile(root, name string) ([]byte, error) {
	if name != "go.mod" && name != "list.go" && name != "list_test.go" {
		return nil, errors.New("file is outside the probe task")
	}
	path := filepath.Join(root, name)
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() < 1 || info.Size() > maxSourceBytes {
		return nil, errors.New("probe source must be a bounded regular file")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open probe source: %w", err)
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil || !os.SameFile(info, opened) {
		return nil, errors.New("probe source changed during open")
	}
	// Bound the read even if a trusted preparation input changes concurrently.
	data := make([]byte, maxSourceBytes+1)
	count, err := file.ReadAt(data, 0)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("read probe source: %w", err)
	}
	if count != int(info.Size()) || count > maxSourceBytes {
		return nil, errors.New("probe source size changed during read")
	}
	after, err := file.Stat()
	if err != nil || after.Size() != info.Size() || !after.ModTime().Equal(info.ModTime()) {
		return nil, errors.New("probe source changed during read")
	}
	return data[:count], nil
}

func Inspect(plan Plan, candidate string) (Inspection, error) {
	if err := plan.Validate(); err != nil {
		return Inspection{}, err
	}
	entries, err := os.ReadDir(candidate)
	if err != nil || len(entries) != len(plan.SourceFiles) {
		return Inspection{}, errors.New("candidate must contain exactly the task files")
	}
	result := Inspection{ProbeID: plan.ProbeID, BaseDigest: TreeDigest(plan.SourceFiles), Files: []SourceFile{}, ChangedPaths: []string{}}
	for index, source := range plan.SourceFiles {
		if entries[index].Name() != source.Path {
			return Inspection{}, errors.New("candidate contains an unexpected path")
		}
		data, err := ReadSourceFile(candidate, source.Path)
		if err != nil {
			return Inspection{}, err
		}
		digest := Digest(data)
		if digest != source.SHA256 {
			if source.Path != "list.go" {
				return Inspection{}, errors.New("candidate modified a protected task file")
			}
			result.ChangedPaths = append(result.ChangedPaths, source.Path)
		}
		result.Files = append(result.Files, SourceFile{Path: source.Path, SHA256: digest})
	}
	result.CandidateDigest = TreeDigest(result.Files)
	return result, nil
}

// Prepare creates inspectable copies only. It starts no SDK, model or process.
// Output must be new; incomplete output is retained on IO failure for diagnosis.
func Prepare(plan Plan, source, output string) (Inspection, error) {
	if err := plan.Validate(); err != nil {
		return Inspection{}, err
	}
	baseline, err := Inspect(plan, source)
	if err != nil || len(baseline.ChangedPaths) != 0 {
		return Inspection{}, errors.New("probe source does not match the frozen baseline")
	}
	if err := os.Mkdir(output, 0o700); err != nil {
		return Inspection{}, fmt.Errorf("create new probe output without overwrite: %w", err)
	}
	for _, name := range []string{"base", "candidate"} {
		directory := filepath.Join(output, name)
		if err := os.Mkdir(directory, 0o700); err != nil {
			return Inspection{}, err
		}
		for _, sourceFile := range plan.SourceFiles {
			data, err := ReadSourceFile(source, sourceFile.Path)
			if err != nil || Digest(data) != sourceFile.SHA256 {
				return Inspection{}, errors.New("probe source changed during preparation")
			}
			mode := os.FileMode(0o400)
			if name == "candidate" && sourceFile.Path == "list.go" {
				mode = 0o600
			}
			if err := os.WriteFile(filepath.Join(directory, sourceFile.Path), data, mode); err != nil {
				return Inspection{}, err
			}
		}
	}
	inspection, err := Inspect(plan, filepath.Join(output, "candidate"))
	if err != nil {
		return Inspection{}, err
	}
	metadata := struct {
		SchemaVersion string     `json:"schema_version"`
		Mode          string     `json:"mode"`
		PlanDigest    string     `json:"plan_digest"`
		Inspection    Inspection `json:"inspection"`
		Execution     string     `json:"execution"`
	}{"development_task_probe_workspace.v1", plan.Mode, PlanDigest(plan), inspection, "not_started"}
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return Inspection{}, err
	}
	if err := os.WriteFile(filepath.Join(output, "workspace.json"), append(data, '\n'), 0o600); err != nil {
		return Inspection{}, err
	}
	return inspection, nil
}
