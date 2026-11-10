package developmentsandbox

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"unicode/utf8"

	control "radishmind.local/services/platform/internal/developmentcontrol"
	probe "radishmind.local/services/platform/internal/developmentprobe"
)

// Sources are flat, frozen probe files. No model path is used before exact
// membership validation. Neither source nor candidate is mounted writable.
func readRegular(directory, name string) ([]byte, error) {
	if name != "go.mod" && name != "list.go" && name != "list_test.go" {
		return nil, ErrSandbox
	}
	info, err := os.Lstat(filepath.Join(directory, name))
	if err != nil || !info.Mode().IsRegular() || info.Size() < 1 || info.Size() > 32768 {
		return nil, ErrSandbox
	}
	if !singleLink(info) {
		return nil, ErrSandbox
	}
	file, err := os.Open(filepath.Join(directory, name))
	if err != nil {
		return nil, ErrSandbox
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil || !os.SameFile(info, opened) {
		return nil, ErrSandbox
	}
	data, err := io.ReadAll(io.LimitReader(file, 32769))
	if err != nil || len(data) != int(info.Size()) || !utf8.Valid(data) || bytes.IndexByte(data, 0) >= 0 {
		return nil, ErrSandbox
	}
	after, err := file.Stat()
	if err != nil || !after.ModTime().Equal(info.ModTime()) || after.Size() != info.Size() {
		return nil, ErrSandbox
	}
	return data, nil
}
func (s *Session) inspect() (probe.Inspection, error) {
	info, err := os.Lstat(filepath.Join(s.config.Directory, "candidate"))
	if err != nil || !info.IsDir() || info.Mode().Perm() != 0700 {
		return probe.Inspection{}, ErrSandbox
	}
	result, err := probe.Inspect(s.config.Probe, filepath.Join(s.config.Directory, "candidate"))
	if err != nil {
		return result, ErrSandbox
	}
	for _, file := range result.Files {
		if _, err := readRegular(filepath.Join(s.config.Directory, "candidate"), file.Path); err != nil {
			return result, err
		}
	}
	return result, nil
}
func atomicJSON(path string, value any) error {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return ErrSandbox
	}
	return atomicFile(path, append(body, '\n'), 0600)
}
func atomicFile(path string, body []byte, mode os.FileMode) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".pending-")
	if err != nil {
		return ErrSandbox
	}
	temporary := file.Name()
	defer os.Remove(temporary)
	if err = file.Chmod(mode); err == nil {
		_, err = file.Write(body)
	}
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(temporary, path)
	}
	if err != nil {
		return ErrSandbox
	}
	directory, err := os.Open(filepath.Dir(path))
	if err != nil {
		return ErrSandbox
	}
	defer directory.Close()
	if directory.Sync() != nil {
		return ErrSandbox
	}
	return nil
}
func (s *Session) snapshot(request string) (string, error) {
	directory := filepath.Join(s.config.Directory, request)
	if err := os.Mkdir(directory, 0700); err != nil {
		return "", ErrSandbox
	}
	for _, file := range s.config.Probe.SourceFiles {
		data, err := readRegular(filepath.Join(s.config.Directory, "candidate"), file.Path)
		if err != nil {
			return "", err
		}
		if err = os.WriteFile(filepath.Join(directory, file.Path), data, 0444); err != nil {
			return "", ErrSandbox
		}
	}
	if err := os.Chmod(directory, 0555); err != nil {
		return "", ErrSandbox
	}
	return directory, nil
}
func marshalObservation(result Observation) (string, error) {
	body, err := json.Marshal(result)
	if err != nil || len(body) > 65536 {
		return "", ErrSandbox
	}
	return string(body), nil
}
func observation(action control.ToolAction, digest string) Observation {
	return Observation{SchemaVersion: "development_sandbox_observation.v1", Action: action.Action, Path: action.Path, CandidateDigest: digest}
}
