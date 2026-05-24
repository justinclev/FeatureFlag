package models

// EvaluationContext contains user-specific data used to match rules.
type EvaluationContext struct {
	UserID     string         `json:"userId"`
	Country    string         `json:"country"`
	State      string         `json:"state"`
	City       string         `json:"city"`
	ZipCode    string         `json:"zipCode"`
	Attributes map[string]any `json:"attributes"`
}

// EvaluationResult is the response returned to the client.
type EvaluationResult struct {
	Enabled  bool           `json:"enabled"`
	Reason   string         `json:"reason"`
	Metadata map[string]any `json:"metadata,omitempty"`
}
