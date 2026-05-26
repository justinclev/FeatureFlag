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

type Rule struct {
	Type   string         `json:"type"`
	Config map[string]any `json:"config"`
	Value  bool           `json:"value"`
}

type CreateFlagRequest struct {
	Key               string `json:"key"`
	Name              string `json:"name"`
	Enabled           bool   `json:"enabled"`
	DefaultValue      bool   `json:"defaultValue"`
	RuleMatchStrategy string `json:"ruleMatchStrategy,omitempty"`
	Rules             []Rule `json:"rules"`
}

func main() {
	flags := []CreateFlagRequest{
		{
			Key:          "defaultfeatureflag",
			Name:         "Default Flag (No Rules)",
			Enabled:      true,
			DefaultValue: true,
			Rules:        []Rule{},
		},
		{
			Key:               "attributesfeatureflag",
			Name:              "Attributes Flag (Market/Product)",
			Enabled:           true,
			DefaultValue:      false,
			RuleMatchStrategy: "any",
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
			Key:          "userfeatureflag",
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
	
	// 1. Check if it exists and get ID
	getReq, _ := http.NewRequest("GET", baseURL, nil)
	getReq.Header.Set("X-API-KEY", apiKey)
	getResp, err := http.DefaultClient.Do(getReq)
	if err == nil {
		defer getResp.Body.Close()
		var flags []struct { ID string `json:"id"`; Key string `json:"key"` }
		json.NewDecoder(getResp.Body).Decode(&flags)
		for _, existing := range flags {
			if strings.EqualFold(existing.Key, f.Key) {
				// Delete by ID
				delReq, _ := http.NewRequest("DELETE", baseURL+"/"+existing.ID, nil)
				delReq.Header.Set("X-API-KEY", apiKey)
				http.DefaultClient.Do(delReq)
				break
			}
		}
	}

	// 2. Create
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
