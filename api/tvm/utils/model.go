package tvmUtils

// RawResponse represents the raw JSON structure of the API response.
type RawResponse struct {
	Success  bool        `json:"success"`
	ExitCode int         `json:"exit_code"`
	Stack    []StackItem `json:"stack"`
}

// StackItem represents individual items in the stack array with the custom format.
type StackItem struct {
	Type string `json:"type"`
	Num  string `json:"num,omitempty"`
	Cell string `json:"cell,omitempty"`
}
