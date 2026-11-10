package workbench

import (
	"errors"
	"os"
	"path/filepath"
)

// Lock is an OS-owned advisory lock on the data directory, shared by all
// workbench commands and containers. Keep the file after release: unlinking a
// locked inode would let another process create a second independent lock.
type Lock struct{ file *os.File }

func Acquire(directory string) (*Lock, error) {
	return acquire(directory, false)
}

// Only explicit initialization and restore may establish workbench ownership.
// Otherwise an already running legacy process could predate the marker.
func acquire(directory string, create bool) (*Lock, error) {
	info, err := os.Stat(directory)
	if err != nil || !info.IsDir() {
		return nil, errors.New("workbench data directory is missing or invalid; initialize it first")
	}
	path := filepath.Join(directory, ".workbench.lock")
	if info, err := os.Lstat(path); err == nil && !info.Mode().IsRegular() {
		return nil, errors.New("workbench lock must be a regular file")
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	flags := os.O_RDWR
	if create {
		flags |= os.O_CREATE
	}
	file, err := os.OpenFile(path, flags, 0600)
	if errors.Is(err, os.ErrNotExist) {
		return nil, errors.New("data directory was not initialized by workbench; use init for a new database or restore a workbench backup")
	}
	if err != nil {
		return nil, errors.New("workbench data directory is not writable")
	}
	if err := lockFile(file); err != nil {
		file.Close()
		return nil, errors.New("workbench data directory is in use; stop the other instance before continuing")
	}
	return &Lock{file: file}, nil
}

func (lock *Lock) Close() error {
	if lock == nil || lock.file == nil {
		return nil
	}
	file := lock.file
	lock.file = nil
	return file.Close()
}
