//go:build !linux && !darwin

package developmentcontrol

import "os/exec"

func prepareWorkerProcess(*exec.Cmd) error { return ErrWorker }
func signalWorkerGroup(int, bool) error    { return ErrWorker }
func workerGroupEmpty(int) bool            { return false }

func reapWorkerGroup(int)                  {}
func validateWorkerDirectory(string) error { return ErrWorker }
