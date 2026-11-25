package httpapi

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	control "radishmind.local/services/platform/internal/developmentcontrol"
	"radishmind.local/services/platform/internal/developmentsandbox"
)

func (s *Server) recoverWorkbenchTaskWorkers() error {
	if s.workbenchLock == nil {
		return control.ErrUnavailable
	}
	s.developmentMu.Lock()
	store, err := s.developmentStoreLocked()
	s.developmentMu.Unlock()
	if err != nil {
		return err
	}
	var check func(context.Context, control.TaskResourceOwner) (bool, error)
	if docker := s.options.DevelopmentTaskDockerPath; docker != "" {
		if !filepath.IsAbs(docker) {
			return control.ErrTaskInvalid
		}
		workers := developmentsandbox.TaskWorkers{DockerPath: docker, Directory: filepath.Join(filepath.Dir(s.config.SQLiteDevDatabasePath), "development-tasks")}
		check = func(ctx context.Context, owner control.TaskResourceOwner) (bool, error) {
			inspection, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			return workers.RecoverCode(inspection, owner)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err = control.RecoverWorkbenchTasks(ctx, store, check)
	if errors.Is(err, control.ErrTaskTerminationUnconfirmed) {
		return nil
	} // History remains available with the occupied slot.
	return err
}
