package httpapi

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"radishmind.local/services/platform/internal/config"
	control "radishmind.local/services/platform/internal/developmentcontrol"
	probe "radishmind.local/services/platform/internal/developmentprobe"
	"radishmind.local/services/platform/internal/developmentsandbox"
)

// Pinned to the audited launcher under services/runtime/development_sdk.
const fixedTaskLauncher = "sha256:7445ea8facecb23ed2ed7c11f8be7776d5788b2b5797e5f7abc6b5d1a12e0c2f"

type workbenchTaskEnvironment struct {
	workers             developmentsandbox.TaskWorkers
	settings            config.WorkbenchTaskExecution
	configPath, keyPath string
}

func newWorkbenchTaskEnvironment(cfg config.Config, t config.WorkbenchTaskExecution) *workbenchTaskEnvironment {
	w := developmentsandbox.TaskWorkers{
		Directory: filepath.Join(filepath.Dir(cfg.SQLiteDevDatabasePath), "development-tasks"), DockerPath: t.DockerPath,
		GoImageDigest: fixedTaskGoImage, SDKImageDigest: fixedTaskSDKImage,
		RunnerPath: filepath.Join(t.InputsDirectory, "code-runner"), Source: filepath.Join(t.InputsDirectory, "list-filter"), SDKPath: t.LauncherPath,
	}
	return &workbenchTaskEnvironment{w, t, cfg.ConfigFile, cfg.APIKeyFile}
}

func taskInputFile(path string, limit int64) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() <= 0 || info.Size() > limit {
		return nil, control.ErrUnavailable
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, control.ErrUnavailable
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil || !os.SameFile(opened, info) {
		return nil, control.ErrUnavailable
	}
	raw, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil || int64(len(raw)) != info.Size() {
		return nil, control.ErrUnavailable
	}
	return raw, nil
}

func (h *workbenchTaskEnvironment) controllerDockerAvailable() bool {
	return runtime.GOOS == "linux" && runtime.GOARCH == "arm64" && os.Geteuid() == 0 && developmentsandbox.CheckControllerExecutable(h.settings.DockerPath) == nil
}

func (h *workbenchTaskEnvironment) launcherArgs(mode string) []string {
	return []string{mode, fixedTaskSDKImage, fixedTaskGoImage, h.settings.DockerPath, filepath.Dir(h.workers.Directory), h.configPath, h.keyPath}
}

// The pinned launcher verifies the actual SDK isolation (including its image)
// and the code image on its local daemon. Its verify mode launches no workloads.
func (h *workbenchTaskEnvironment) verified(ctx context.Context) (developmentsandbox.TaskWorkers, error) {
	w := h.workers
	if !h.controllerDockerAvailable() {
		return w, control.ErrUnavailable
	}
	private := filepath.Dir(w.Directory)
	inputs, err := filepath.Rel(private, h.settings.InputsDirectory)
	if err != nil || !filepath.IsLocal(inputs) {
		return w, control.ErrUnavailable
	}
	if err := developmentsandbox.CheckControllerExecutable(w.SDKPath); err != nil {
		return w, control.ErrUnavailable
	}
	raw, err := taskInputFile(w.SDKPath, 64<<10)
	if err != nil || control.PayloadDigest(raw) != fixedTaskLauncher {
		return w, control.ErrUnavailable
	}
	raw, err = taskInputFile(w.RunnerPath, 16<<20)
	if err != nil || control.PayloadDigest(raw) != fixedTaskExecutor {
		return w, control.ErrUnavailable
	}
	raw, err = taskInputFile(filepath.Join(h.settings.InputsDirectory, "list-filter-plan.json"), 64<<10)
	if err != nil {
		return w, err
	}
	w.Probe, err = probe.ParsePlan(raw)
	if err != nil || w.Probe.ProbeID != fixedTaskSource || w.Probe.TaskID != fixedTaskVersion || probe.TreeDigest(w.Probe.SourceFiles) != fixedTaskBaseline {
		return w, control.ErrUnavailable
	}
	snapshot, err := probe.Inspect(w.Probe, w.Source)
	if err != nil || len(snapshot.ChangedPaths) != 0 {
		return w, control.ErrUnavailable
	}
	check, stop := context.WithTimeout(ctx, 4*time.Second)
	defer stop()
	cmd := exec.CommandContext(check, w.SDKPath, h.launcherArgs("verify")...)
	cmd.Env = []string{"PATH=/usr/local/bin:/usr/bin:/bin", "LANG=C"}
	cmd.WaitDelay = 250 * time.Millisecond
	var daemon bytes.Buffer
	cmd.Stdout = &daemon // The pinned verifier emits only the daemon identity.
	if err := cmd.Run(); err != nil {
		return w, control.ErrUnavailable
	}
	if err := w.CheckDaemon(check, bytes.TrimSpace(daemon.Bytes())); err != nil {
		return w, control.ErrUnavailable
	}
	w.SDKArgs = h.launcherArgs("run")
	return w, nil
}
func (h *workbenchTaskEnvironment) Check(ctx context.Context) error {
	_, err := h.verified(ctx)
	return err
}
func (h *workbenchTaskEnvironment) Prepare(ctx context.Context, o control.TaskResourceOwner, p control.TaskPlan, d control.Delegation) (control.WorkerLaunch, func() error, error) {
	w, err := h.verified(ctx)
	if err != nil {
		return control.WorkerLaunch{}, nil, err
	}
	if err := os.Mkdir(w.Directory, 0700); err != nil && !errors.Is(err, os.ErrExist) {
		return control.WorkerLaunch{}, nil, control.ErrUnavailable
	}
	return w.Prepare(ctx, o, p, d)
}
