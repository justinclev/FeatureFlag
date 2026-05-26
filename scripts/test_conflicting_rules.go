package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

const (
	baseURL = "http://localhost:8081/api/flags"
	apiKey  = "test-api-key"
)

type Rule struct {
	Type   string         `json:"type"`
	Config map[string]any `json:"config"`
	Value  bool           `json:"value"`
}

type CreateFlagRequest struct {
	Key               string `json:"key"`
	Name              string `json:"name"`
	Enabled           bool   `json:"enabled"`
	OffValue          bool   `json:"offValue"`
	FallthroughValue  bool   `json:"fallthroughValue"`
	RuleMatchStrategy string `json:"ruleMatchStrategy"`
	Rules             []Rule `json:"rules"`
}

type EvalContext struct {
	UserID     string         `json:"userId,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

type EvalResponse struct {
	Enabled bool   `json:"enabled"`
	Reason  string `json:"reason"`
}

func main() {
	rand.Seed(time.Now().UnixNano())
	suffix := fmt.Sprintf("%d", rand.Intn(100000))

	// Scenario: Rule 1 is TRUE, Rule 2 is FALSE.
	// Context matches BOTH.
	// ANY strategy should return FALSE (Deny Wins) even though Rule 1 is TRUE.
	rules := []Rule{
		{
			Type: "attribute",
			Config: map[string]any{"attributeKey": "Product", "attributeOp": "eq", "attributeValue": "Pro"},
			Value: true,
		},
		{
			Type: "attribute",
			Config: map[string]any{"attributeKey": "Market", "attributeOp": "eq", "attributeValue": "US"},
			Value: false,
		},
	}

	// Context that matches BOTH rules
	ctx := EvalContext{
		Attributes: map[string]any{"Market": "US", "Product": "Pro"},
	}

	fmt.Println("--- CONFLICTING RULES TEST (DENY WINS) ---")
	fmt.Println("Rule 1 (First): Product=Pro -> true")
	fmt.Println("Rule 2 (Last):  Market=US -> false")
	fmt.Println("Context Sent:   Market=US, Product=Pro (Matches BOTH)")
	fmt.Println(strings.Repeat("-", 60))

	// 1. Test ANY strategy (Should return FALSE now)
	runScenario("any", rules, ctx, suffix)

	// 2. Test ALL strategy (Should fail CREATION now)
	fmt.Println("\n--- STRICT VALIDATION TEST (ALL STRATEGY - CREATE) ---")
	testAllStrategyValidation(rules, suffix)

	// 3. Test ALL strategy (Should fail UPDATE now)
	fmt.Println("\n--- STRICT VALIDATION TEST (ALL STRATEGY - UPDATE) ---")
	testAllStrategyUpdateValidation(rules, suffix)
}

func testAllStrategyUpdateValidation(rules []Rule, suffix string) {
	key := "conflict-test-update-validation-" + suffix
	
	// Create a valid 'all' flag first
	validRules := []Rule{rules[0]} // Only one rule is always valid
	f := CreateFlagRequest{
		Key:               key,
		Name:              "Update Validation Test",
		Enabled:           true,
		OffValue:          true,
		FallthroughValue:  true,
		RuleMatchStrategy: "all",
		Rules:             validRules,
	}

	// Delete old
	delReq, _ := http.NewRequest("DELETE", baseURL+"/"+key, nil)
	delReq.Header.Set("X-API-KEY", apiKey)
	http.DefaultClient.Do(delReq)

	// Create new
	body, _ := json.Marshal(f)
	creReq, _ := http.NewRequest("POST", baseURL, bytes.NewBuffer(body))
	creReq.Header.Set("X-API-KEY", apiKey)
	creReq.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(creReq)
	
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		fmt.Printf("FAILURE: Failed to create initial flag: %d %s\n", resp.StatusCode, respBody)
		return
	}

	var created struct { ID string `json:"id"` }
	json.NewDecoder(resp.Body).Decode(&created)
	id := created.ID
	if id == "" {
		fmt.Println("FAILURE: Created flag ID is empty")
		return
	}

	// Now try to PATCH it with conflicting rules
	update := map[string]any{
		"rules": rules, // multiple conflicting rules
	}
	upBody, _ := json.Marshal(update)
	upReq, _ := http.NewRequest("PATCH", baseURL+"/"+id, bytes.NewBuffer(upBody))
	upReq.Header.Set("X-API-KEY", apiKey)
	upReq.Header.Set("Content-Type", "application/json")
	
	upResp, _ := http.DefaultClient.Do(upReq)
	defer upResp.Body.Close()

	if upResp.StatusCode == http.StatusBadRequest {
		respBody, _ := io.ReadAll(upResp.Body)
		fmt.Printf("SUCCESS: Update blocked as expected (HTTP 400). Error: %s\n", strings.TrimSpace(string(respBody)))
	} else {
		fmt.Printf("FAILURE: Update should have been blocked (400), but got HTTP %d\n", upResp.StatusCode)
		respBody, _ := io.ReadAll(upResp.Body)
		fmt.Println(string(respBody))
	}
}

func testAllStrategyValidation(rules []Rule, suffix string) {
	key := "conflict-test-all-validation-" + suffix
	f := CreateFlagRequest{
		Key:               key,
		Name:              "Conflict Test All Validation",
		Enabled:           true,
		OffValue:          true,
		FallthroughValue:  true,
		RuleMatchStrategy: "all",
		Rules:             rules,
	}

	body, _ := json.Marshal(f)
	req, _ := http.NewRequest("POST", baseURL, bytes.NewBuffer(body))
	req.Header.Set("X-API-KEY", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		respBody, _ := io.ReadAll(resp.Body)
		fmt.Printf("SUCCESS: Creation blocked as expected (HTTP 400). Error: %s\n", strings.TrimSpace(string(respBody)))
	} else {
		fmt.Printf("FAILURE: Creation should have been blocked, but got HTTP %d\n", resp.StatusCode)
	}
}

func runScenario(strategy string, rules []Rule, ctx EvalContext, suffix string) {
	key := "conflict-test-" + strategy + "-" + suffix
	
	// 1. Create flag
	f := CreateFlagRequest{
		Key:               key,
		Name:              "Conflict Test " + strategy,
		Enabled:           true,
		OffValue:          true,
		FallthroughValue:  true, // Different from both rules to see if it falls through
		RuleMatchStrategy: strategy,
		Rules:             rules,
	}

	// Delete old
	delReq, _ := http.NewRequest("DELETE", baseURL+"/"+key, nil)
	delReq.Header.Set("X-API-KEY", apiKey)
	http.DefaultClient.Do(delReq)

	// Create new
	body, _ := json.Marshal(f)
	creReq, _ := http.NewRequest("POST", baseURL, bytes.NewBuffer(body))
	creReq.Header.Set("X-API-KEY", apiKey)
	creReq.Header.Set("Content-Type", "application/json")
	http.DefaultClient.Do(creReq)

	// 2. Evaluate
	evalBody, _ := json.Marshal(ctx)
	evalReq, _ := http.NewRequest("POST", baseURL+"/"+key+"/evaluate", bytes.NewBuffer(evalBody))
	evalReq.Header.Set("X-API-KEY", apiKey)
	evalReq.Header.Set("Content-Type", "application/json")
	
	resp, _ := http.DefaultClient.Do(evalReq)
	defer resp.Body.Close()

	var eval EvalResponse
	json.NewDecoder(resp.Body).Decode(&eval)

	fmt.Printf("Strategy: %-5s | Result: %-5t | Reason: %s\n", strategy, eval.Enabled, eval.Reason)
}
