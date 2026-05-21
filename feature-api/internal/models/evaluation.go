package models

// EvaluationContext contains user information for flag evaluation.
type EvaluationContext struct {
	UserID     string            `bson:"userId"     json:"userId"`
	Country    string            `bson:"country"    json:"country"`
	State      string            `bson:"state"      json:"state"`
	City       string            `bson:"city"       json:"city"`
	ZipCode    string            `bson:"zipCode"    json:"zipCode"`
	Attributes map[string]string `bson:"attributes" json:"attributes"`
}

// EvaluationResult contains the result of flag evaluation.
type EvaluationResult struct {
	Enabled bool   `json:"enabled"`
	Reason  string `json:"reason"`
}
