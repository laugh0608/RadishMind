//go:build linux || darwin

package developmentcontrol

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

func prepareWorkerProcess(cmd *exec.Cmd) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return nil
}
func signalWorkerGroup(pid int, hard bool) error {
	signal := syscall.SIGTERM
	if hard {
		signal = syscall.SIGKILL
	}
	err := syscall.Kill(-pid, signal)
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}
	return err
}
func workerGroupEmpty(pid int) bool { return errors.Is(syscall.Kill(-pid, 0), syscall.ESRCH) }

func reapWorkerGroup(pid int) {
	for {
		var status syscall.WaitStatus
		n, err := syscall.Wait4(-pid, &status, syscall.WNOHANG, nil)
		if err != nil || n <= 0 {
			return
		}
	}
}
func validateWorkerDirectory(path string) error {
	info, err := os.Lstat(path)
	if err != nil || !info.IsDir() || info.Mode().Perm() != 0700 {
		return ErrWorker
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || int(stat.Uid) != os.Geteuid() {
		return ErrWorker
	}
	entries, err := os.ReadDir(path)
	if err != nil || len(entries) != 0 {
		return ErrWorker
	}
	return nil
}
