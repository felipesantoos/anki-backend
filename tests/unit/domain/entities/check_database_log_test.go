package entities

import (
	"testing"

	"github.com/felipesantos/anki-backend/core/domain/entities"
)

func TestCheckDatabaseLog_IsCompleted(t *testing.T) {
	tests := []struct {
		name  string
		log   *entities.CheckDatabaseLog
		want  bool
	}{
		{
			name: "completed status",
			log: &entities.CheckDatabaseLog{
				Status: entities.CheckStatusCompleted,
			},
			want: true,
		},
		{
			name: "running status",
			log: &entities.CheckDatabaseLog{
				Status: entities.CheckStatusRunning,
			},
			want: false,
		},
		{
			name: "failed status",
			log: &entities.CheckDatabaseLog{
				Status: entities.CheckStatusFailed,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.log.IsCompleted()
			if got != tt.want {
				t.Errorf("CheckDatabaseLog.IsCompleted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckDatabaseLog_IsFailed(t *testing.T) {
	tests := []struct {
		name  string
		log   *entities.CheckDatabaseLog
		want  bool
	}{
		{
			name: "failed status",
			log: &entities.CheckDatabaseLog{
				Status: entities.CheckStatusFailed,
			},
			want: true,
		},
		{
			name: "completed status",
			log: &entities.CheckDatabaseLog{
				Status: entities.CheckStatusCompleted,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.log.IsFailed()
			if got != tt.want {
				t.Errorf("CheckDatabaseLog.IsFailed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckDatabaseLog_IsCorrupted(t *testing.T) {
	tests := []struct {
		name  string
		log   *entities.CheckDatabaseLog
		want  bool
	}{
		{
			name: "corrupted status",
			log: &entities.CheckDatabaseLog{
				Status: entities.CheckStatusCorrupted,
			},
			want: true,
		},
		{
			name: "completed status",
			log: &entities.CheckDatabaseLog{
				Status: entities.CheckStatusCompleted,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.log.IsCorrupted()
			if got != tt.want {
				t.Errorf("CheckDatabaseLog.IsCorrupted() = %v, want %v", got, tt.want)
			}
		})
	}
}
