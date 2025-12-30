package ownership

import (
	"errors"
	"testing"

)

func TestValidateOwnership(t *testing.T) {
	tests := []struct {
		name           string
		userID         int64
		resourceUserID int64
		want           bool
	}{
		{
			name:           "same user ID",
			userID:         1,
			resourceUserID: 1,
			want:           true,
		},
		{
			name:           "different user IDs",
			userID:         1,
			resourceUserID: 2,
			want:           false,
		},
		{
			name:           "zero user ID",
			userID:         0,
			resourceUserID: 0,
			want:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateOwnership(tt.userID, tt.resourceUserID)
			if got != tt.want {
				t.Errorf("ValidateOwnership(%d, %d) = %v, want %v", tt.userID, tt.resourceUserID, got, tt.want)
			}
		})
	}
}

func TestEnsureOwnership(t *testing.T) {
	tests := []struct {
		name           string
		userID         int64
		resourceUserID int64
		wantErr        bool
		wantErrType    error
	}{
		{
			name:           "same user ID - no error",
			userID:         1,
			resourceUserID: 1,
			wantErr:        false,
		},
		{
			name:           "different user IDs - error",
			userID:         1,
			resourceUserID: 2,
			wantErr:        true,
			wantErrType:    ErrAccessDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := EnsureOwnership(tt.userID, tt.resourceUserID)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnsureOwnership(%d, %d) error = %v, wantErr %v", tt.userID, tt.resourceUserID, err, tt.wantErr)
				return
			}
			if tt.wantErr && !errors.Is(err, tt.wantErrType) {
				t.Errorf("EnsureOwnership(%d, %d) error = %v, want error type %v", tt.userID, tt.resourceUserID, err, tt.wantErrType)
			}
		})
	}
}

func TestValidateOwnershipSlice(t *testing.T) {
	tests := []struct {
		name            string
		userID          int64
		resourceUserIDs []int64
		wantErr         bool
	}{
		{
			name:            "all resources belong to user",
			userID:          1,
			resourceUserIDs: []int64{1, 1, 1},
			wantErr:         false,
		},
		{
			name:            "one resource doesn't belong",
			userID:          1,
			resourceUserIDs: []int64{1, 2, 1},
			wantErr:         true,
		},
		{
			name:            "all resources don't belong",
			userID:          1,
			resourceUserIDs: []int64{2, 3, 4},
			wantErr:         true,
		},
		{
			name:            "empty slice",
			userID:          1,
			resourceUserIDs: []int64{},
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOwnershipSlice(tt.userID, tt.resourceUserIDs)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOwnershipSlice(%d, %v) error = %v, wantErr %v", tt.userID, tt.resourceUserIDs, err, tt.wantErr)
			}
		})
	}
}

func TestErrResourceNotFound(t *testing.T) {
	err := ErrResourceNotFound
	if err.Error() != "resource not found" {
		t.Errorf("ErrResourceNotFound.Error() = %v, want 'resource not found'", err.Error())
	}
}

func TestErrAccessDenied(t *testing.T) {
	err := ErrAccessDenied
	if err.Error() != "access denied: resource does not belong to user" {
		t.Errorf("ErrAccessDenied.Error() = %v, want 'access denied: resource does not belong to user'", err.Error())
	}
}

