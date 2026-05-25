package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
)

const (
	baseURL     = "http://localhost:8081/api/flags"
	apiKey      = "test-api-key"
	concurrency = 20
	totalReqs   = 100
)

type EvalContext struct {
	UserID     string         `json:"userId,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

type EvalResponse struct {
	Enabled bool   `json:"enabled"`
	Reason  string `json:"reason"`
}

type Scenario struct {
	Name     string
	FlagKey  string
	Context  EvalContext
	Expected bool
}

func main() {
	scenarios := []Scenario{
		// positive
		{"Attr-US-Pos", "attributesFeatureFlag", EvalContext{Attributes: map[string]any{"Market": "US"}}, true},
		{"Attr-Pro-Pos", "attributesFeatureFlag", EvalContext{Attributes: map[string]any{"Product": "Pro"}}, true},
		{"User-123-Pos", "userFeatureFlag", EvalContext{UserID: "user-123"}, true},
		{"Default-Pos", "defaultFeatureFlag", EvalContext{}, true},
		
		// negative
		{"Attr-Neg", "attributesFeatureFlag", EvalContext{Attributes: map[string]any{"Market": "UK", "Product": "Free"}}, false},
		{"User-Neg", "userFeatureFlag", EvalContext{UserID: "random-user"}, false},
		{"Missing-Neg", "NoFlagFlag", EvalContext{}, false},
	}

	fmt.Printf("Starting stress test with %d scenarios and %d concurrency...\n\n", len(scenarios), concurrency)

	var wg sync.WaitGroup
	results := make(chan string, totalReqs)

	for i := 0; i < totalReqs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s := scenarios[rand.Intn(len(scenarios))]
			
			url := fmt.Sprintf("%s/%s/evaluate", baseURL, s.FlagKey)
			body, _ := json.Marshal(s.Context)
			
			req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
			req.Header.Set("X-API-KEY", apiKey)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				results <- fmt.Sprintf("[ERROR] %s: %v", s.Name, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNotFound {
				if !s.Expected {
					results <- fmt.Sprintf("[PASS] %s (404 as expected)", s.Name)
				} else {
					results <- fmt.Sprintf("[FAIL] %s (Unexpected 404)", s.Name)
				}
				return
			}

			if resp.StatusCode != http.StatusOK {
				results <- fmt.Sprintf("[FAIL] %s (HTTP %d)", s.Name, resp.StatusCode)
				return
			}

			var eval EvalResponse
			json.NewDecoder(resp.Body).Decode(&eval)

			if eval.Enabled == s.Expected {
				results <- fmt.Sprintf("[PASS] %s", s.Name)
			} else {
				results <- fmt.Sprintf("[FAIL] %s (Expected %t, got %t)", s.Name, s.Expected, eval.Enabled)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	passCount := 0
	failCount := 0
	for res := range results {
		if len(res) >= 6 && res[:6] == "[PASS]" {
			passCount++
		} else {
			failCount++
			fmt.Println(res)
		}
	}

	fmt.Printf("\nFinal Results: %d Passed, %d Failed\n", passCount, failCount)
}
