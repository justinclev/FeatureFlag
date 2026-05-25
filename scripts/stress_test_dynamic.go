package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
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

type Flag struct {
	ID      string `json:"id"`
	Key     string `json:"key"`
	Enabled bool   `json:"enabled"`
	Rules   []Rule `json:"rules"`
}

func main() {
	fmt.Println("--- DYNAMIC BOMBARDIER STRESS TEST ---")

	// 1. Fetch flags to understand rules
	req, _ := http.NewRequest("GET", baseURL, nil)
	req.Header.Set("X-API-KEY", apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Failed to fetch flags: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var flags []Flag
	json.NewDecoder(resp.Body).Decode(&flags)

	if len(flags) == 0 {
		fmt.Println("No flags found. Run 'make seed-flags' first.")
		os.Exit(1)
	}

	// 2. Generate varied payloads based on rules
	var payloads []string
	for _, f := range flags {
		if !f.Enabled {
			continue
		}

		// Add a positive payload for each rule
		for _, r := range f.Rules {
			payload := make(map[string]any)
			switch r.Type {
			case "user_list":
				if ids, ok := r.Config["userIds"].([]any); ok && len(ids) > 0 {
					payload["userId"] = ids[0]
				}
			case "attribute":
				key, _ := r.Config["attributeKey"].(string)
				val, _ := r.Config["attributeValue"].(string)
				if key != "" {
					payload["attributes"] = map[string]string{key: val}
				}
			}
			if len(payload) > 0 {
				b, _ := json.Marshal(payload)
				payloads = append(payloads, string(b))
			}
		}
		
		// Add a negative payload (random)
		neg, _ := json.Marshal(map[string]any{"userId": "non-existent-user", "attributes": map[string]string{"random": "value"}})
		payloads = append(payloads, string(neg))
	}

	if len(payloads) == 0 {
		// Fallback if no rules found
		payloads = append(payloads, `{"userId":"guest"}`)
	}

	// 3. Select a random flag to slam (Bombardier usually hits one URL at a time)
	// For multi-URL slamming, we'd need to loop or use a proxy, 
	// but we can pick the most complex one (attributes) for the best "slam".
	targetKey := "attributesFeatureFlag"
	targetURL := fmt.Sprintf("%s/%s/evaluate", baseURL, targetKey)
	
	// Use one of our generated payloads
	payload := payloads[0]
	
	fmt.Printf("Slamming %s\n", targetURL)
	fmt.Printf("Payload: %s\n", payload)
	fmt.Println("---------------------------------------")

	// 4. Execute Bombardier
	bombardierPath := "/Users/justinclev/go/bin/bombardier"
	cmd := exec.Command(bombardierPath, 
		"-c", "500", 
		"-d", "10s", 
		"-m", "POST",
		"-H", "X-API-KEY: "+apiKey,
		"-H", "Content-Type: application/json",
		"-b", payload,
		targetURL,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		fmt.Printf("Bombardier failed: %v\n", err)
		fmt.Println("Is 'bombardier' installed? (go install github.com/codesenberg/bombardier@latest)")
	}
}
