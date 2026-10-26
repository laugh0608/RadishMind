package workbench

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
)

// Restore only creates a new directory. Failed output is intentionally retained
// for inspection; neither the original database nor the backup is modified.
func Restore(ctx context.Context, source, destination string) (BackupManifest, error) {
	manifest, err := verifyBackup(ctx, source)
	if err != nil {
		return BackupManifest{}, err
	}
	if destination == "" {
		return BackupManifest{}, errors.New("restore requires a new destination directory")
	}
	if err := os.Mkdir(destination, 0700); err != nil {
		return BackupManifest{}, fmt.Errorf("restore destination must not exist: %w", err)
	}
	dataDir := filepath.Join(destination, "data")
	if err := os.Mkdir(dataDir, 0700); err != nil {
		return BackupManifest{}, err
	}
	lock, err := acquire(dataDir, true)
	if err != nil {
		return BackupManifest{}, err
	}
	defer lock.Close()
	for sourceName, target := range map[string]string{"radishmind.db": filepath.Join(dataDir, "radishmind.db"), "platform.json": filepath.Join(destination, "platform.json")} {
		if err := copyNewFile(filepath.Join(source, sourceName), target); err != nil {
			return BackupManifest{}, err
		}
		checksum, err := fileChecksum(target)
		if err != nil || checksum != manifest.Files[sourceName] {
			return BackupManifest{}, errors.New("backup changed during restore; incomplete target preserved")
		}
	}
	if _, err := Inspect(ctx, filepath.Join(dataDir, "radishmind.db"), true); err != nil {
		return BackupManifest{}, err
	}
	return manifest, nil
}

func verifyBackup(ctx context.Context, source string) (BackupManifest, error) {
	entries, err := os.ReadDir(source)
	if err != nil {
		return BackupManifest{}, err
	}
	if len(entries) != 3 {
		return BackupManifest{}, errors.New("backup must contain exactly database, configuration and manifest")
	}
	for _, entry := range entries {
		if entry.Name() != "manifest.json" && entry.Name() != "platform.json" && entry.Name() != "radishmind.db" {
			return BackupManifest{}, errors.New("unexpected backup file")
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() {
			return BackupManifest{}, errors.New("backup entries must be regular files")
		}
	}
	var manifest BackupManifest
	if err := readSmallJSON(filepath.Join(source, "manifest.json"), &manifest); err != nil {
		return BackupManifest{}, err
	}
	if manifest.FormatVersion != 1 || manifest.CreatedAt.IsZero() || manifest.BuildVersion == "" || len(manifest.Files) != 2 {
		return BackupManifest{}, errors.New("unsupported or incomplete backup manifest")
	}
	for _, name := range []string{"radishmind.db", "platform.json"} {
		checksum, err := fileChecksum(filepath.Join(source, name))
		if err != nil || checksum != manifest.Files[name] {
			return BackupManifest{}, errors.New("backup checksum mismatch")
		}
	}
	var snapshot backupConfiguration
	if err := readSmallJSON(filepath.Join(source, "platform.json"), &snapshot); err != nil {
		return BackupManifest{}, err
	}
	if snapshot.RuntimeProfile != "local-workbench" || snapshot.DatabasePath != "data/radishmind.db" || snapshot.LocalPersistenceMode != "sqlite_dev" ||
		snapshot.Workbench.TenantRef != manifest.TenantRef || snapshot.Workbench.WorkspaceID != manifest.WorkspaceID {
		return BackupManifest{}, errors.New("backup configuration or scope mismatch")
	}
	records, err := Inspect(ctx, filepath.Join(source, "radishmind.db"), true)
	if err != nil {
		return BackupManifest{}, err
	}
	if !reflect.DeepEqual(records, manifest.Migrations) {
		return BackupManifest{}, errors.New("backup migration inventory mismatch")
	}
	return manifest, nil
}

func readSmallJSON(path string, value any) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(io.LimitReader(file, 65537))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return errors.New("invalid backup JSON document")
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return errors.New("unexpected data after backup JSON document")
	}
	return nil
}

func copyNewFile(source, destination string) (resultErr error) {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, output.Close()) }()
	if _, err := io.Copy(output, input); err != nil {
		return err
	}
	return output.Sync()
}
