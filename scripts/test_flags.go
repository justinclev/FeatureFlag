package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	baseURL = "http://localhost:8081/api/flags"
	apiKey  = "test-api-key"
)

type EvalContext struct {
	UserID     string         `json:"userId,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

type EvalResponse struct {
	Enabled bool   `json:"enabled"`
	Reason  string `json:"reason"`
}

type TestCase struct {
	Name          string
	FlagKey       string
	Context       EvalContext
	ExpectedValue bool
	InputDesc     string // Custom description of input values for report
}

func main() {
	testCases := []TestCase{
		// 1. Default Flag (No Rules, Enabled=true, DefaultValue=true)
		{
			Name:          "Default Flag - Positive",
			FlagKey:       "defaultFeatureFlag",
			Context:       EvalContext{},
			ExpectedValue: true,
			InputDesc:     "SENT: {} | RULE: No rules, Default=true",
		},

		// 2. Attributes Flag (Market=US OR Product=Pro)
		{
			Name:    "Attributes Flag - Positive (Market)",
			FlagKey: "attributesFeatureFlag",
			Context: EvalContext{
				Attributes: map[string]any{"Market": "US"},
			},
			ExpectedValue: true,
			InputDesc:     "SENT: Market=US | RULE: Market=US",
		},
		{
			Name:    "Attributes Flag - Positive (Product)",
			FlagKey: "attributesFeatureFlag",
			Context: EvalContext{
				Attributes: map[string]any{"Product": "Pro"},
			},
			ExpectedValue: true,
			InputDesc:     "SENT: Product=Pro | RULE: Product=Pro",
		},
		{
			Name:    "Attributes Flag - Negative (Mismatch)",
			FlagKey: "attributesFeatureFlag",
			Context: EvalContext{
				Attributes: map[string]any{"Market": "UK", "Product": "Basic"},
			},
			ExpectedValue: false,
			InputDesc:     "SENT: Market=UK, Product=Basic | RULE: Market=US, Product=Pro",
		},

		// 3. User ID Flag (IDs: user-123, admin-99)
		{
			Name:    "User ID Flag - Positive",
			FlagKey: "userFeatureFlag",
			Context: EvalContext{
				UserID: "user-123",
			},
			ExpectedValue: true,
			InputDesc:     "SENT: UserID=user-123 | RULE: user-123, admin-99",
		},
		{
			Name:    "User ID Flag - Negative",
			FlagKey: "userFeatureFlag",
			Context: EvalContext{
				UserID: "guest-user",
			},
			ExpectedValue: false,
			InputDesc:     "SENT: UserID=guest-user | RULE: user-123, admin-99",
		},

		// 4. Missing Flag
		{
			Name:          "Missing Flag - Negative",
			FlagKey:       "NoFlagFlag",
			Context:       EvalContext{},
			ExpectedValue: false,
			InputDesc:     "SENT: Key=NoFlagFlag | RULE: Non-existent",
		},
	}

	fmt.Printf("%-35s | %-20s | %-10s | %-10s | %-s\n", "TEST NAME", "FLAG KEY", "EXPECTED", "ACTUAL", "INPUT/RULES COMPARISON")
	fmt.Println(strings.Repeat("-", 140))

	for _, tc := range testCases {
		runTest(tc)
	}
}

func runTest(tc TestCase) {
	url := fmt.Sprintf("%s/%s/evaluate", baseURL, tc.FlagKey)
	
	body, _ := json.Marshal(tc.Context)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("X-API-KEY", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("%-35s | %-20s | %-10t | %-10s | %-s\n", tc.Name, tc.FlagKey, tc.ExpectedValue, "ERROR", err.Error())
		return
	}
	defer resp.Body.Close()

	// Special case for missing flags (404)
	if resp.StatusCode == http.StatusNotFound && !tc.ExpectedValue {
		fmt.Printf("%-35s | %-20s | %-10t | %-10s | %-s [%s]\n", tc.Name, tc.FlagKey, tc.ExpectedValue, "FALSE(404)", tc.InputDesc, "PASS")
		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("%-35s | %-20s | %-10t | %-10s | %-s\n", tc.Name, tc.FlagKey, tc.ExpectedValue, fmt.Sprintf("HTTP %d", resp.StatusCode), tc.InputDesc)
		return
	}

	var eval EvalResponse
	json.NewDecoder(resp.Body).Decode(&eval)

	status := "PASS"
	if eval.Enabled != tc.ExpectedValue {
		status = "FAIL"
	}

	fmt.Printf("%-35s | %-20s | %-10t | %-10t | %-s [%s]\n", tc.Name, tc.FlagKey, tc.ExpectedValue, eval.Enabled, tc.InputDesc, status)
}
