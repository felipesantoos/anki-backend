package entities

import (
	"testing"

	"github.com/felipesantos/anki-backend/core/domain/entities/backup"
)

func TestBackup_IsAutomatic(t *testing.T) {
	b := &backup.Backup{}
	b.SetBackupType(backup.BackupTypeAutomatic)

	if !b.IsAutomatic() {
		t.Errorf("Backup.IsAutomatic() = false, want true")
	}

	manualBackup := &backup.Backup{}
	manualBackup.SetBackupType(backup.BackupTypeManual)

	if manualBackup.IsAutomatic() {
		t.Errorf("Backup.IsAutomatic() = true, want false for manual backup")
	}
}

func TestBackup_IsManual(t *testing.T) {
	b := &backup.Backup{}
	b.SetBackupType(backup.BackupTypeManual)

	if !b.IsManual() {
		t.Errorf("Backup.IsManual() = false, want true")
	}

	automaticBackup := &backup.Backup{}
	automaticBackup.SetBackupType(backup.BackupTypeAutomatic)

	if automaticBackup.IsManual() {
		t.Errorf("Backup.IsManual() = true, want false for automatic backup")
	}
}

func TestBackup_IsPreOperation(t *testing.T) {
	b := &backup.Backup{}
	b.SetBackupType(backup.BackupTypePreOperation)

	if !b.IsPreOperation() {
		t.Errorf("Backup.IsPreOperation() = false, want true")
	}

	automaticBackup := &backup.Backup{}
	automaticBackup.SetBackupType(backup.BackupTypeAutomatic)

	if automaticBackup.IsPreOperation() {
		t.Errorf("Backup.IsPreOperation() = true, want false for automatic backup")
	}
}

