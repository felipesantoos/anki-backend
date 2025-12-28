package entities

import (
	"testing"

	"github.com/felipesantos/anki-backend/core/domain/entities"
)

func TestBackup_IsAutomatic(t *testing.T) {
	backup := &entities.Backup{
		BackupType: entities.BackupTypeAutomatic,
	}

	if !backup.IsAutomatic() {
		t.Errorf("Backup.IsAutomatic() = false, want true")
	}

	manualBackup := &entities.Backup{
		BackupType: entities.BackupTypeManual,
	}

	if manualBackup.IsAutomatic() {
		t.Errorf("Backup.IsAutomatic() = true, want false for manual backup")
	}
}

func TestBackup_IsManual(t *testing.T) {
	backup := &entities.Backup{
		BackupType: entities.BackupTypeManual,
	}

	if !backup.IsManual() {
		t.Errorf("Backup.IsManual() = false, want true")
	}

	automaticBackup := &entities.Backup{
		BackupType: entities.BackupTypeAutomatic,
	}

	if automaticBackup.IsManual() {
		t.Errorf("Backup.IsManual() = true, want false for automatic backup")
	}
}

func TestBackup_IsPreOperation(t *testing.T) {
	backup := &entities.Backup{
		BackupType: entities.BackupTypePreOperation,
	}

	if !backup.IsPreOperation() {
		t.Errorf("Backup.IsPreOperation() = false, want true")
	}

	automaticBackup := &entities.Backup{
		BackupType: entities.BackupTypeAutomatic,
	}

	if automaticBackup.IsPreOperation() {
		t.Errorf("Backup.IsPreOperation() = true, want false for automatic backup")
	}
}

