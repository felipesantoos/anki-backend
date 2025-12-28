package ownership

import (
	"errors"
	"fmt"
)

var (
	// ErrResourceNotFound is returned when a resource is not found
	// This error is used to avoid revealing whether a resource exists or belongs to another user
	ErrResourceNotFound = errors.New("resource not found")

	// ErrAccessDenied is returned when a user tries to access a resource they don't own
	// This should only be used internally - handlers should return 404 to avoid information leakage
	ErrAccessDenied = errors.New("access denied: resource does not belong to user")
)

// ValidateOwnership checks if a user owns a resource
// Returns true if userID matches resourceUserID, false otherwise
func ValidateOwnership(userID, resourceUserID int64) bool {
	return userID == resourceUserID
}

// EnsureOwnership validates ownership and returns an error if the user doesn't own the resource
// Returns ErrAccessDenied if ownership validation fails
func EnsureOwnership(userID, resourceUserID int64) error {
	if !ValidateOwnership(userID, resourceUserID) {
		return fmt.Errorf("%w: user %d does not own resource (owner: %d)", ErrAccessDenied, userID, resourceUserID)
	}
	return nil
}

// ValidateOwnershipSlice validates ownership for multiple resources
// Returns an error if any resource doesn't belong to the user
// The error message includes which resources failed validation
func ValidateOwnershipSlice(userID int64, resourceUserIDs []int64) error {
	var invalidResources []int64
	for _, resourceUserID := range resourceUserIDs {
		if !ValidateOwnership(userID, resourceUserID) {
			invalidResources = append(invalidResources, resourceUserID)
		}
	}

	if len(invalidResources) > 0 {
		return fmt.Errorf("%w: user %d does not own %d resource(s)", ErrAccessDenied, userID, len(invalidResources))
	}

	return nil
}

