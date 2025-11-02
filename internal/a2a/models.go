package a2a

import (
	"encoding/json"
	"time"
)

// JSON-RPC types
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// Message types
type MessageParams struct {
	Message       A2AMessage           `json:"message"`
	Configuration MessageConfiguration `json:"configuration"`
}

type A2AMessage struct {
	Kind      string        `json:"kind"`
	Role      string        `json:"role"`
	Parts     []MessagePart `json:"parts"`
	MessageID string        `json:"messageId,omitempty"`
	TaskID    *string       `json:"taskId,omitempty"`
}

type MessagePart struct {
	Kind string      `json:"kind"`
	Text interface{} `json:"text,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

type MessageConfiguration struct {
	AcceptedOutputModes []string `json:"acceptedOutputModes,omitempty"`
	HistoryLength       int      `json:"historyLength,omitempty"`
	Blocking            bool     `json:"blocking,omitempty"`
}

// Task types
type TaskResult struct {
	ID        string       `json:"id"`
	ContextID string       `json:"contextId,omitempty"`
	Status    TaskStatus   `json:"status"`
	Artifacts []Artifact   `json:"artifacts,omitempty"`
	History   []A2AMessage `json:"history,omitempty"`
	Kind      string       `json:"kind"`
}

type TaskStatus struct {
	State     string      `json:"state"`
	Timestamp string      `json:"timestamp"`
	Message   *A2AMessage `json:"message,omitempty"`
}

type Artifact struct {
	ArtifactID string        `json:"artifactId"`
	Name       string        `json:"name"`
	Parts      []MessagePart `json:"parts"`
}

// Helper functions
func TextPart(text string) MessagePart {
	return MessagePart{
		Kind: "text",
		Text: &text,
	}
}

func DataPart(data map[string]interface{}) MessagePart {
	dataBytes, _ := json.Marshal(data)
	return MessagePart{
		Kind: "text",
		Text: string(dataBytes),
	}
}

func Timestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// Task states
const (
	StateWorking       = "working"
	StateInputRequired = "input-required"
	StateCompleted     = "completed"
	StateFailed        = "failed"
)

// Message roles
const (
	RoleUser  = "user"
	RoleAgent = "agent"
)
