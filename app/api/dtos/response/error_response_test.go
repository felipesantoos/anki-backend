package response

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewErrorResponse(t *testing.T) {
	errorCode := "NOT_FOUND"
	message := "Resource not found"
	requestID := "abc123def456"
	path := "/api/users/123"

	resp := NewErrorResponse(errorCode, message, requestID, path)

	if resp.Error != errorCode {
		t.Errorf("Expected Error '%s', got '%s'", errorCode, resp.Error)
	}

	if resp.Message != message {
		t.Errorf("Expected Message '%s', got '%s'", message, resp.Message)
	}

	if resp.Code != errorCode {
		t.Errorf("Expected Code '%s', got '%s'", errorCode, resp.Code)
	}

	if resp.RequestID != requestID {
		t.Errorf("Expected RequestID '%s', got '%s'", requestID, resp.RequestID)
	}

	if resp.Path != path {
		t.Errorf("Expected Path '%s', got '%s'", path, resp.Path)
	}

	// Verify timestamp is set and valid
	if resp.Timestamp == "" {
		t.Error("Expected Timestamp to be set")
	}

	// Verify timestamp is valid RFC3339 format
	if _, err := time.Parse(time.RFC3339, resp.Timestamp); err != nil {
		t.Errorf("Expected Timestamp to be valid RFC3339 format, got error: %v", err)
	}
}

func TestErrorResponse_JSON(t *testing.T) {
	resp := NewErrorResponse("VALIDATION_ERROR", "Invalid input", "req123", "/api/test")

	jsonData, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal ErrorResponse to JSON: %v", err)
	}

	// Verify JSON contains expected fields
	jsonStr := string(jsonData)
	if !contains(jsonStr, "error") {
		t.Error("Expected JSON to contain 'error' field")
	}
	if !contains(jsonStr, "message") {
		t.Error("Expected JSON to contain 'message' field")
	}
	if !contains(jsonStr, "code") {
		t.Error("Expected JSON to contain 'code' field")
	}
	if !contains(jsonStr, "request_id") {
		t.Error("Expected JSON to contain 'request_id' field")
	}
	if !contains(jsonStr, "timestamp") {
		t.Error("Expected JSON to contain 'timestamp' field")
	}
	if !contains(jsonStr, "path") {
		t.Error("Expected JSON to contain 'path' field")
	}

	// Verify JSON can be unmarshaled back
	var unmarshaled ErrorResponse
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal JSON back to ErrorResponse: %v", err)
	}

	if unmarshaled.Error != resp.Error {
		t.Errorf("Expected Error '%s', got '%s'", resp.Error, unmarshaled.Error)
	}

	if unmarshaled.Message != resp.Message {
		t.Errorf("Expected Message '%s', got '%s'", resp.Message, unmarshaled.Message)
	}

	if unmarshaled.Code != resp.Code {
		t.Errorf("Expected Code '%s', got '%s'", resp.Code, unmarshaled.Code)
	}

	if unmarshaled.RequestID != resp.RequestID {
		t.Errorf("Expected RequestID '%s', got '%s'", resp.RequestID, unmarshaled.RequestID)
	}

	if unmarshaled.Path != resp.Path {
		t.Errorf("Expected Path '%s', got '%s'", resp.Path, unmarshaled.Path)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
