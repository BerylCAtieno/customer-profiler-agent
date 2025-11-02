package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
)

type TestClient struct {
	baseURL string
	client  *http.Client
}

func NewTestClient(baseURL string) *TestClient {
	return &TestClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func main() {
	baseURL := flag.String("url", "http://localhost:8080", "Base URL of the agent")
	testType := flag.String("test", "all", "Test type: all, health, agent-card, profile, custom")
	businessIdea := flag.String("idea", "", "Business idea for profile generation (for custom test)")
	flag.Parse()

	client := NewTestClient(*baseURL)

	printHeader("Customer Profiler Agent - Test Suite")
	fmt.Printf("%sBase URL: %s%s\n\n", colorCyan, *baseURL, colorReset)

	switch *testType {
	case "all":
		client.runAllTests()
	case "health":
		client.testHealthCheck()
	case "agent-card":
		client.testAgentCard()
	case "profile":
		client.testProfileGeneration()
	case "custom":
		if *businessIdea == "" {
			printError("Business idea is required for custom test. Use -idea flag")
			os.Exit(1)
		}
		client.testCustomProfile(*businessIdea)
	default:
		printError(fmt.Sprintf("Unknown test type: %s", *testType))
		fmt.Println("\nAvailable tests: all, health, agent-card, profile, custom")
		os.Exit(1)
	}
}

func (tc *TestClient) runAllTests() {
	tests := []struct {
		name string
		fn   func() bool
	}{
		{"Health Check", tc.testHealthCheck},
		{"Agent Card", tc.testAgentCard},
		{"Profile Generation", tc.testProfileGeneration},
	}

	passed := 0
	failed := 0

	for _, test := range tests {
		if test.fn() {
			passed++
		} else {
			failed++
		}
		fmt.Println()
	}

	printHeader("Test Summary")
	fmt.Printf("%sPassed: %d%s\n", colorGreen, passed, colorReset)
	fmt.Printf("%sFailed: %d%s\n", colorRed, failed, colorReset)
	fmt.Printf("Total: %d\n", passed+failed)

	if failed > 0 {
		os.Exit(1)
	}
}

func (tc *TestClient) testHealthCheck() bool {
	printTestHeader("Testing Health Check Endpoint")

	url := fmt.Sprintf("%s/health", tc.baseURL)
	fmt.Printf("GET %s\n", url)

	resp, err := tc.client.Get(url)
	if err != nil {
		printError(fmt.Sprintf("Request failed: %v", err))
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		printError(fmt.Sprintf("Expected status 200, got %d", resp.StatusCode))
		return false
	}

	if string(body) != "OK" {
		printError(fmt.Sprintf("Expected body 'OK', got '%s'", string(body)))
		return false
	}

	printSuccess("Health check passed")
	return true
}

func (tc *TestClient) testAgentCard() bool {
	printTestHeader("Testing Agent Card Endpoint")

	url := fmt.Sprintf("%s/.well-known/agent.json", tc.baseURL)
	fmt.Printf("GET %s\n", url)

	resp, err := tc.client.Get(url)
	if err != nil {
		printError(fmt.Sprintf("Request failed: %v", err))
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		printError(fmt.Sprintf("Expected status 200, got %d", resp.StatusCode))
		fmt.Printf("Response: %s\n", string(body))
		return false
	}

	// Parse JSON to validate it's valid
	var agentCard map[string]interface{}
	if err := json.Unmarshal(body, &agentCard); err != nil {
		printError(fmt.Sprintf("Invalid JSON response: %v", err))
		return false
	}

	// Check required fields
	requiredFields := []string{"name", "description", "version", "capabilities", "endpoints"}
	for _, field := range requiredFields {
		if _, ok := agentCard[field]; !ok {
			printError(fmt.Sprintf("Missing required field: %s", field))
			return false
		}
	}

	printSuccess("Agent card is valid")
	printJSON(body)
	return true
}

func (tc *TestClient) testProfileGeneration() bool {
	businessIdea := "A sustainable fashion e-commerce platform targeting eco-conscious millennials"
	return tc.testCustomProfile(businessIdea)
}

func (tc *TestClient) testCustomProfile(businessIdea string) bool {
	printTestHeader("Testing Profile Generation")

	url := fmt.Sprintf("%s/a2a/profiler", tc.baseURL)
	fmt.Printf("POST %s\n", url)
	fmt.Printf("%sBusiness Idea:%s %s\n\n", colorCyan, colorReset, businessIdea)

	// Create JSON-RPC request
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      fmt.Sprintf("test-%d", time.Now().Unix()),
		"method":  "agent/task",
		"params": map[string]interface{}{
			"message": map[string]interface{}{
				"kind": "message",
				"role": "user",
				"parts": []map[string]interface{}{
					{
						"kind": "text",
						"text": businessIdea,
					},
				},
			},
			"configuration": map[string]interface{}{
				"blocking":            true,
				"acceptedOutputModes": []string{"text", "data"},
			},
		},
	}

	jsonData, _ := json.MarshalIndent(request, "", "  ")
	fmt.Printf("%sRequest:%s\n", colorYellow, colorReset)
	fmt.Println(string(jsonData))
	fmt.Println()

	resp, err := tc.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		printError(fmt.Sprintf("Request failed: %v", err))
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		printError(fmt.Sprintf("Expected status 200, got %d", resp.StatusCode))
		fmt.Printf("Response: %s\n", string(body))
		return false
	}

	// Parse JSON-RPC response
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		printError(fmt.Sprintf("Invalid JSON response: %v", err))
		return false
	}

	// Check for errors
	if errObj, ok := response["error"]; ok {
		printError("Request returned an error")
		errJSON, _ := json.MarshalIndent(errObj, "", "  ")
		fmt.Println(string(errJSON))
		return false
	}

	// Check result
	result, ok := response["result"].(map[string]interface{})
	if !ok {
		printError("Invalid result format")
		return false
	}

	// Check task status
	status, ok := result["status"].(map[string]interface{})
	if !ok {
		printError("Invalid status format")
		return false
	}

	state, _ := status["state"].(string)
	if state != "completed" {
		printError(fmt.Sprintf("Expected state 'completed', got '%s'", state))
		return false
	}

	printSuccess("Profile generation completed successfully")

	// Display the response message
	if msg, ok := status["message"].(map[string]interface{}); ok {
		if parts, ok := msg["parts"].([]interface{}); ok {
			fmt.Printf("\n%sGenerated Profile:%s\n", colorGreen, colorReset)
			fmt.Println(strings.Repeat("=", 80))
			for _, part := range parts {
				if p, ok := part.(map[string]interface{}); ok {
					if text, ok := p["text"].(string); ok {
						fmt.Println(text)
					}
				}
			}
			fmt.Println(strings.Repeat("=", 80))
		}
	}

	// Display artifacts if any
	if artifacts, ok := result["artifacts"].([]interface{}); ok && len(artifacts) > 0 {
		fmt.Printf("\n%sArtifacts:%s\n", colorPurple, colorReset)
		artifactsJSON, _ := json.MarshalIndent(artifacts, "", "  ")
		fmt.Println(string(artifactsJSON))
	}

	return true
}

func printHeader(text string) {
	fmt.Printf("\n%s%s%s\n", colorBlue, strings.Repeat("=", len(text)+4), colorReset)
	fmt.Printf("%s= %s =%s\n", colorBlue, text, colorReset)
	fmt.Printf("%s%s%s\n\n", colorBlue, strings.Repeat("=", len(text)+4), colorReset)
}

func printTestHeader(text string) {
	fmt.Printf("%s[TEST] %s%s\n", colorCyan, text, colorReset)
	fmt.Println(strings.Repeat("-", 80))
}

func printSuccess(text string) {
	fmt.Printf("%s✓ %s%s\n", colorGreen, text, colorReset)
}

func printError(text string) {
	fmt.Printf("%s✗ %s%s\n", colorRed, text, colorReset)
}

func printJSON(data []byte) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err == nil {
		fmt.Printf("\n%sResponse:%s\n%s\n", colorYellow, colorReset, prettyJSON.String())
	}
}
