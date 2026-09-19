package commspliant

import "fmt"

// APIError is returned when the CommsPliant API responds with a non-success status.
type APIError struct {
	StatusCode  int
	Message     string
	Code        string
	Details     map[string]any
	RequestID   string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("commspliant api error: %s (status %d)", e.Message, e.StatusCode)
	}
	return fmt.Sprintf("commspliant api error: status %d", e.StatusCode)
}
