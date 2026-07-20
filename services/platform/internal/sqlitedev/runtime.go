package sqlitedev

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

const (
	defaultBusyTimeout  = 5 * time.Second
	defaultMaxOpenConns = 8
)

type Options struct {
	DatabasePath string
	BusyTimeout  time.Duration
	Migrations   []Migration
}

type Runtime struct {
	database     *sql.DB
	databasePath string
	timeout      time.Duration
	closeOnce    sync.Once
	closeErr     error
}

func Open(ctx context.Context, options Options) (*Runtime, error) {
	databasePath := strings.TrimSpace(options.DatabasePath)
	if databasePath == "" {
		return nil, errors.New("SQLite development database path is missing")
	}
	busyTimeout := options.BusyTimeout
	if busyTimeout == 0 {
		busyTimeout = defaultBusyTimeout
	}
	if busyTimeout < time.Millisecond {
		return nil, errors.New("SQLite development busy timeout must be at least one millisecond")
	}
	if err := validateMigrations(options.Migrations); err != nil {
		return nil, err
	}

	absolutePath, err := filepath.Abs(filepath.Clean(databasePath))
	if err != nil {
		return nil, errors.New("resolve SQLite development database path")
	}
	if err := prepareDatabaseDirectory(filepath.Dir(absolutePath)); err != nil {
		return nil, err
	}
	if err := secureExistingDatabaseFile(absolutePath); err != nil {
		return nil, err
	}

	database, err := sql.Open("sqlite", databaseDSN(absolutePath, busyTimeout))
	if err != nil {
		return nil, errors.New("open SQLite development database")
	}
	database.SetMaxOpenConns(defaultMaxOpenConns)
	database.SetMaxIdleConns(defaultMaxOpenConns)
	runtime := &Runtime{database: database, databasePath: absolutePath, timeout: busyTimeout}
	if err := runtime.initialize(ctx, absolutePath, options.Migrations); err != nil {
		_ = database.Close()
		return nil, err
	}
	return runtime, nil
}

func (runtime *Runtime) DB() *sql.DB {
	if runtime == nil {
		return nil
	}
	return runtime.database
}

func (runtime *Runtime) Close() error {
	if runtime == nil || runtime.database == nil {
		return nil
	}
	runtime.closeOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), runtime.timeout)
		defer cancel()
		if err := secureDatabaseSidecars(runtime.databasePath); err != nil {
			runtime.closeErr = err
		}
		if _, err := runtime.database.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
			if runtime.closeErr == nil {
				runtime.closeErr = errors.New("checkpoint SQLite development database")
			}
		}
		if err := runtime.database.Close(); err != nil && runtime.closeErr == nil {
			runtime.closeErr = errors.New("close SQLite development database")
		}
	})
	return runtime.closeErr
}

func (runtime *Runtime) initialize(ctx context.Context, databasePath string, migrations []Migration) error {
	if err := runtime.database.PingContext(ctx); err != nil {
		return errors.New("connect SQLite development database")
	}
	if err := os.Chmod(databasePath, 0o600); err != nil {
		return errors.New("secure SQLite development database file")
	}
	if err := verifyDatabaseSettings(ctx, runtime.database, runtime.timeout); err != nil {
		return err
	}
	if err := verifyDatabaseIntegrity(ctx, runtime.database); err != nil {
		return err
	}
	if err := applyMigrations(ctx, runtime.database, migrations); err != nil {
		return err
	}
	if err := secureDatabaseSidecars(databasePath); err != nil {
		return err
	}
	return nil
}

func prepareDatabaseDirectory(directoryPath string) error {
	info, err := os.Stat(directoryPath)
	switch {
	case err == nil && !info.IsDir():
		return errors.New("SQLite development database parent is not a directory")
	case err == nil:
		return nil
	case !errors.Is(err, os.ErrNotExist):
		return errors.New("inspect SQLite development database directory")
	}
	if err := os.MkdirAll(directoryPath, 0o700); err != nil {
		return errors.New("create SQLite development database directory")
	}
	if err := os.Chmod(directoryPath, 0o700); err != nil {
		return errors.New("secure SQLite development database directory")
	}
	return nil
}

func secureExistingDatabaseFile(databasePath string) error {
	info, err := os.Lstat(databasePath)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return nil
	case err != nil:
		return errors.New("inspect SQLite development database file")
	case !info.Mode().IsRegular():
		return errors.New("SQLite development database path is not a regular file")
	default:
		if err := os.Chmod(databasePath, 0o600); err != nil {
			return errors.New("secure SQLite development database file")
		}
		return nil
	}
}

func secureDatabaseSidecars(databasePath string) error {
	for _, sidecarPath := range []string{databasePath + "-wal", databasePath + "-shm"} {
		info, err := os.Lstat(sidecarPath)
		switch {
		case errors.Is(err, os.ErrNotExist):
			continue
		case err != nil:
			return errors.New("inspect SQLite development sidecar file")
		case !info.Mode().IsRegular():
			return errors.New("SQLite development sidecar path is not a regular file")
		default:
			if err := os.Chmod(sidecarPath, 0o600); err != nil {
				return errors.New("secure SQLite development sidecar file")
			}
		}
	}
	return nil
}

func databaseDSN(databasePath string, busyTimeout time.Duration) string {
	databaseURL := url.URL{Scheme: "file", Path: filepath.ToSlash(databasePath)}
	query := databaseURL.Query()
	query.Add("_pragma", "busy_timeout("+strconv.FormatInt(busyTimeout.Milliseconds(), 10)+")")
	query.Add("_pragma", "foreign_keys(ON)")
	query.Add("_pragma", "journal_mode(WAL)")
	databaseURL.RawQuery = query.Encode()
	return databaseURL.String()
}

func verifyDatabaseSettings(ctx context.Context, database *sql.DB, busyTimeout time.Duration) error {
	var foreignKeys int
	if err := database.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&foreignKeys); err != nil || foreignKeys != 1 {
		return errors.New("verify SQLite development foreign key enforcement")
	}
	var journalMode string
	if err := database.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&journalMode); err != nil || !strings.EqualFold(journalMode, "wal") {
		return errors.New("verify SQLite development WAL mode")
	}
	var busyTimeoutMilliseconds int64
	if err := database.QueryRowContext(ctx, "PRAGMA busy_timeout").Scan(&busyTimeoutMilliseconds); err != nil || busyTimeoutMilliseconds != busyTimeout.Milliseconds() {
		return errors.New("verify SQLite development busy timeout")
	}
	return nil
}

func verifyDatabaseIntegrity(ctx context.Context, database *sql.DB) error {
	rows, err := database.QueryContext(ctx, "PRAGMA quick_check")
	if err != nil {
		return errors.New("check SQLite development database integrity")
	}
	defer rows.Close()
	checked := false
	for rows.Next() {
		checked = true
		var result string
		if err := rows.Scan(&result); err != nil {
			return errors.New("read SQLite development integrity result")
		}
		if result != "ok" {
			return errors.New("SQLite development database integrity check failed")
		}
	}
	if err := rows.Err(); err != nil {
		return errors.New("complete SQLite development database integrity check")
	}
	if !checked {
		return errors.New("SQLite development database integrity result is missing")
	}
	return nil
}

func sanitizedSQLiteError(operation string) error {
	return errors.New(operation + ": SQLite operation failed")
}
