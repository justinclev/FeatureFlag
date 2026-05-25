package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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
	Key          string `json:"key"`
	Name         string `json:"name"`
	Enabled      bool   `json:"enabled"`
	DefaultValue bool   `json:"defaultValue"`
	Rules        []Rule `json:"rules"`
}

func main() {
	flags := []CreateFlagRequest{
		{
			Key:          "defaultFeatureFlag",
			Name:         "Default Flag (No Rules)",
			Enabled:      true,
			DefaultValue: true,
			Rules:        []Rule{},
		},
		{
			Key:          "attributesFeatureFlag",
			Name:         "Attributes Flag (Market/Product)",
			Enabled:      true,
			DefaultValue: false,
			Rules: []Rule{
				{
					Type: "attribute",
					Config: map[string]any{
						"attributeKey":   "Market",
						"attributeOp":    "eq",
						"attributeValue": "US",
					},
					Value: true,
				},
				{
					Type: "attribute",
					Config: map[string]any{
						"attributeKey":   "Product",
						"attributeOp":    "eq",
						"attributeValue": "Pro",
					},
					Value: true,
				},
			},
		},
		{
			Key:          "userFeatureFlag",
			Name:         "User ID Flag",
			Enabled:      true,
			DefaultValue: false,
			Rules: []Rule{
				{
					Type: "user_list",
					Config: map[string]any{
						"userIds": []string{"user-123", "admin-99"},
					},
					Value: true,
				},
			},
		},
	}

	for _, f := range flags {
		createFlag(f)
	}
}

func createFlag(f CreateFlagRequest) {
	fmt.Printf("Creating flag: %s...", f.Key)
	
	// Delete first to ensure clean state (ignore error if not found)
	deleteReq, _ := http.NewRequest("DELETE", baseURL+"/"+f.Key, nil)
	deleteReq.Header.Set("X-API-KEY", apiKey)
	http.DefaultClient.Do(deleteReq)

	body, _ := json.Marshal(f)
	req, _ := http.NewRequest("POST", baseURL, bytes.NewBuffer(body))
	req.Header.Set("X-API-KEY", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf(" FAIL: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		fmt.Println(" SUCCESS")
	} else {
		fmt.Printf(" FAIL: HTTP %d\n", resp.StatusCode)
	}
}
