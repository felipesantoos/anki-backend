package valueobjects

import (
	"testing"

	"github.com/felipesantos/anki-backend/core/domain/valueobjects"
)

func TestSchedulerType_IsValid(t *testing.T) {
	tests := []struct {
		name       string
		scheduler  valueobjects.SchedulerType
		want       bool
	}{
		{
			name:       "valid sm2",
			scheduler:  valueobjects.SchedulerTypeSM2,
			want:       true,
		},
		{
			name:       "valid fsrs",
			scheduler:  valueobjects.SchedulerTypeFSRS,
			want:       true,
		},
		{
			name:       "invalid scheduler",
			scheduler:  valueobjects.SchedulerType("invalid"),
			want:       false,
		},
		{
			name:       "empty scheduler",
			scheduler:  valueobjects.SchedulerType(""),
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.scheduler.IsValid()
			if got != tt.want {
				t.Errorf("SchedulerType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSchedulerType_String(t *testing.T) {
	if valueobjects.SchedulerTypeSM2.String() != "sm2" {
		t.Errorf("SchedulerTypeSM2.String() = %v, want 'sm2'", valueobjects.SchedulerTypeSM2.String())
	}
	if valueobjects.SchedulerTypeFSRS.String() != "fsrs" {
		t.Errorf("SchedulerTypeFSRS.String() = %v, want 'fsrs'", valueobjects.SchedulerTypeFSRS.String())
	}
}

