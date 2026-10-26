package workbench

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"radishmind.local/services/platform/internal/config"
	"radishmind.local/services/platform/internal/sqlitedev"
	localmigrations "radishmind.local/services/platform/migrations/sqlite"
)

type BackupManifest struct {
	FormatVersion int               `json:"format_version"`
	CreatedAt     time.Time         `json:"created_at"`
	BuildVersion  string            `json:"build_version"`
	TenantRef     string            `json:"tenant_ref"`
	WorkspaceID   string            `json:"workspace_id"`
	Migrations    []MigrationRecord `json:"migrations"`
	Files         map[string]string `json:"files"`
}

type backupConfiguration struct {
	Temperature            float64                `json:"temperature"`
	PythonBinary           string                 `json:"python_binary"`
	BridgeScript           string                 `json:"bridge_script"`
	ReadHeaderTimeout      string                 `json:"read_header_timeout"`
	WriteTimeout           string                 `json:"write_timeout"`
	BridgeTimeout          string                 `json:"bridge_timeout"`
	BridgeMode             string                 `json:"bridge_mode"`
	BridgeWorkerCount      int                    `json:"bridge_worker_count"`
	BridgeQueueCapacity    int                    `json:"bridge_queue_capacity"`
	BridgeHandshakeTimeout string                 `json:"bridge_handshake_timeout"`
	RuntimeProfile         string                 `json:"runtime_profile"`
	ListenAddr             string                 `json:"listen_addr"`
	LocalPersistenceMode   string                 `json:"local_persistence_mode"`
	DatabasePath           string                 `json:"sqlite_dev_database_path"`
	Provider               string                 `json:"provider"`
	Model                  string                 `json:"model"`
	BaseURL                string                 `json:"base_url,omitempty"`
	APIKeyFile             string                 `json:"api_key_file,omitempty"`
	Workbench              config.WorkbenchConfig `json:"workbench"`
}

func Backup(ctx context.Context, cfg config.Config, destination, buildVersion string) error {
	lock, err := Acquire(filepath.Dir(cfg.SQLiteDevDatabasePath))
	if err != nil {
		return err
	}
	defer lock.Close()
	return backupLocked(ctx, cfg, destination, buildVersion)
}

func backupLocked(ctx context.Context, cfg config.Config, destination, buildVersion string) error {
	records, err := Inspect(ctx, cfg.SQLiteDevDatabasePath, true)
	if err != nil {
		return err
	}
	if destination == "" {
		return errors.New("backup requires an explicit new destination directory")
	}
	if err := os.Mkdir(destination, 0700); err != nil {
		return fmt.Errorf("create new backup directory (existing targets are never overwritten): %w", err)
	}
	db, err := openReadOnly(cfg.SQLiteDevDatabasePath)
	if err != nil {
		return err
	}
	_, snapshotErr := db.ExecContext(ctx, "VACUUM INTO ?", filepath.Join(destination, "radishmind.db"))
	closeErr := db.Close()
	if snapshotErr != nil {
		return errors.New("create consistent SQLite backup snapshot; incomplete directory preserved")
	}
	if closeErr != nil {
		return closeErr
	}
	if err := os.Chmod(filepath.Join(destination, "radishmind.db"), 0600); err != nil {
		return err
	}
	if _, err := Inspect(ctx, filepath.Join(destination, "radishmind.db"), true); err != nil {
		return err
	}
	snapshot := backupConfiguration{
		Temperature: cfg.Temperature, PythonBinary: cfg.PythonBinary, BridgeScript: cfg.BridgeScript,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout.String(), WriteTimeout: cfg.WriteTimeout.String(), BridgeTimeout: cfg.BridgeTimeout.String(),
		BridgeMode: cfg.BridgeMode, BridgeWorkerCount: cfg.BridgeWorkerCount, BridgeQueueCapacity: cfg.BridgeQueueCapacity, BridgeHandshakeTimeout: cfg.BridgeHandshakeTimeout.String(),
		RuntimeProfile: config.LocalWorkbenchProfile, ListenAddr: cfg.ListenAddr, LocalPersistenceMode: "sqlite_dev",
		DatabasePath: "data/radishmind.db", Provider: cfg.Provider, Model: cfg.Model, BaseURL: cfg.BaseURL, Workbench: cfg.Workbench,
	}
	if cfg.Provider != "mock" {
		snapshot.APIKeyFile = "secrets/provider-key"
	}
	if err := writeNewJSON(filepath.Join(destination, "platform.json"), snapshot); err != nil {
		return err
	}
	manifest := BackupManifest{FormatVersion: 1, CreatedAt: time.Now().UTC(), BuildVersion: buildVersion,
		TenantRef: cfg.Workbench.TenantRef, WorkspaceID: cfg.Workbench.WorkspaceID, Migrations: records, Files: map[string]string{}}
	for _, name := range []string{"radishmind.db", "platform.json"} {
		checksum, err := fileChecksum(filepath.Join(destination, name))
		if err != nil {
			return err
		}
		manifest.Files[name] = checksum
	}
	// Publishing the manifest last distinguishes a complete backup from a failed one.
	return writeNewJSON(filepath.Join(destination, "manifest.json"), manifest)
}

func Upgrade(ctx context.Context, cfg config.Config, backupDestination, buildVersion string) error {
	lock, err := Acquire(filepath.Dir(cfg.SQLiteDevDatabasePath))
	if err != nil {
		return err
	}
	defer lock.Close()
	if err := backupLocked(ctx, cfg, backupDestination, buildVersion); err != nil {
		return err
	}
	runtime, err := sqlitedev.Open(ctx, sqlitedev.Options{DatabasePath: cfg.SQLiteDevDatabasePath, Migrations: localmigrations.Migrations()})
	if err != nil {
		return fmt.Errorf("upgrade failed; backup and database preserved, service remains stopped: %w", err)
	}
	if err := runtime.Close(); err != nil {
		return err
	}
	_, err = Inspect(ctx, cfg.SQLiteDevDatabasePath, false)
	return err
}

func fileChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}

func writeNewJSON(path string, value any) (resultErr error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, file.Close()) }()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return err
	}
	return file.Sync()
}
