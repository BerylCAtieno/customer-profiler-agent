package a2a

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/BerylCAtieno/customer-profiler-agent/internal/agent"
	"github.com/BerylCAtieno/customer-profiler-agent/internal/models"
	"github.com/BerylCAtieno/customer-profiler-agent/internal/profiler"
	"github.com/gin-gonic/gin"
)

type A2AHandler struct {
	geminiClient *profiler.GeminiClient
}

func NewA2AHandler(geminiClient *profiler.GeminiClient) *A2AHandler {
	return &A2AHandler{
		geminiClient: geminiClient,
	}
}

// HandleProfiler processes A2A messages
func (h *A2AHandler) HandleProfiler(c *gin.Context) {
	// Parse JSON-RPC request
	var rpcReq JSONRPCRequest
	if err := c.ShouldBindJSON(&rpcReq); err != nil {
		log.Printf("Failed to decode request: %v", err)
		h.sendErrorResponse(c, "", "Invalid request format", -32700)
		return
	}

	// Validate JSON-RPC version
	if rpcReq.JSONRPC != "2.0" {
		h.sendErrorResponse(c, rpcReq.ID, "Invalid JSON-RPC version", -32600)
		return
	}

	// Handle different methods
	switch rpcReq.Method {
	case "agent/task":
		h.handleTask(c, rpcReq)
	default:
		h.sendErrorResponse(c, rpcReq.ID, fmt.Sprintf("Method not found: %s", rpcReq.Method), -32601)
	}
}

func (h *A2AHandler) handleTask(c *gin.Context, rpcReq JSONRPCRequest) {
	// Parse message parameters
	paramsJSON, err := json.Marshal(rpcReq.Params)
	if err != nil {
		h.sendErrorResponse(c, rpcReq.ID, "Failed to parse parameters", -32602)
		return
	}

	var msgParams MessageParams
	if err := json.Unmarshal(paramsJSON, &msgParams); err != nil {
		log.Printf("Failed to unmarshal params: %v", err)
		h.sendErrorResponse(c, rpcReq.ID, "Invalid parameters", -32602)
		return
	}

	// Extract business idea from user message
	businessIdea := h.extractBusinessIdea(msgParams.Message)
	if businessIdea == "" {
		result := h.createErrorTaskResult(
			rpcReq.ID,
			"Please provide a business idea to generate customer profiles.",
		)
		h.sendSuccessResponse(c, rpcReq.ID, result)
		return
	}

	// Generate customer profiles
	ctx := context.Background()
	profileResp, err := h.geminiClient.GenerateCustomerProfiles(ctx, businessIdea)
	if err != nil {
		log.Printf("Failed to generate profiles: %v", err)
		result := h.createErrorTaskResult(
			rpcReq.ID,
			fmt.Sprintf("Failed to generate customer profiles: %v", err),
		)
		h.sendSuccessResponse(c, rpcReq.ID, result)
		return
	}

	// Create successful task result
	result := h.createSuccessTaskResult(rpcReq.ID, profileResp)
	h.sendSuccessResponse(c, rpcReq.ID, result)
}

// ServeAgentCard serves the agent card using Gin
func (h *A2AHandler) ServeAgentCard(c *gin.Context) {
	if err := agent.LoadAgentCard(); err != nil {
		log.Printf("Error loading agent card: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Agent card not available"})
		return
	}

	c.Data(http.StatusOK, "application/json", agent.AgentCardData)
}

func (h *A2AHandler) extractBusinessIdea(msg A2AMessage) string {
	var texts []string
	for _, part := range msg.Parts {
		if part.Kind == "text" && part.Text != nil {
			texts = append(texts, *part.Text)
		}
	}
	return strings.TrimSpace(strings.Join(texts, " "))
}

func (h *A2AHandler) createSuccessTaskResult(taskID string, profileResp *models.ProfileResponse) TaskResult {
	// Format the profile data nicely
	responseText := h.formatProfileResponse(profileResp)

	// Create data artifact with the full profile
	profileData := map[string]interface{}{
		"businessIdea": profileResp.BusinessIdea,
		"profiles":     profileResp.Profiles,
	}

	return TaskResult{
		ID:   taskID,
		Kind: "task",
		Status: TaskStatus{
			State:     StateCompleted,
			Timestamp: Timestamp(),
			Message: &A2AMessage{
				Kind: "message",
				Role: RoleAgent,
				Parts: []MessagePart{
					TextPart(responseText),
				},
			},
		},
		Artifacts: []Artifact{
			{
				ArtifactID: fmt.Sprintf("profile-%s", taskID),
				Name:       "Customer Profile Data",
				Parts: []MessagePart{
					DataPart(profileData),
				},
			},
		},
	}
}

func (h *A2AHandler) createErrorTaskResult(taskID string, errorMsg string) TaskResult {
	return TaskResult{
		ID:   taskID,
		Kind: "task",
		Status: TaskStatus{
			State:     StateFailed,
			Timestamp: Timestamp(),
			Message: &A2AMessage{
				Kind: "message",
				Role: RoleAgent,
				Parts: []MessagePart{
					TextPart(errorMsg),
				},
			},
		},
	}
}

func (h *A2AHandler) formatProfileResponse(profileResp *models.ProfileResponse) string {
	if len(profileResp.Profiles) == 0 {
		return "No customer profiles generated."
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("# Customer Profile for: %s\n\n", profileResp.BusinessIdea))

	for i, profile := range profileResp.Profiles {
		if i > 0 {
			builder.WriteString("\n---\n\n")
		}

		builder.WriteString("**Demographics:**\n")
		builder.WriteString(fmt.Sprintf("- Age: %s\n", profile.Age))
		builder.WriteString(fmt.Sprintf("- Gender: %s\n", profile.Gender))
		builder.WriteString(fmt.Sprintf("- Location: %s\n", profile.Location))
		builder.WriteString(fmt.Sprintf("- Occupation: %s\n", profile.Occupation))
		builder.WriteString(fmt.Sprintf("- Income: %s\n\n", profile.Income))

		if len(profile.PainPoints) > 0 {
			builder.WriteString("**Pain Points:**\n")
			for _, pp := range profile.PainPoints {
				builder.WriteString(fmt.Sprintf("- %s\n", strings.TrimSpace(pp)))
			}
			builder.WriteString("\n")
		}

		if len(profile.Motivations) > 0 {
			builder.WriteString("**Motivations:**\n")
			for _, m := range profile.Motivations {
				builder.WriteString(fmt.Sprintf("- %s\n", strings.TrimSpace(m)))
			}
			builder.WriteString("\n")
		}

		if len(profile.Interests) > 0 {
			builder.WriteString("**Interests:**\n")
			for _, interest := range profile.Interests {
				builder.WriteString(fmt.Sprintf("- %s\n", strings.TrimSpace(interest)))
			}
			builder.WriteString("\n")
		}

		if len(profile.PreferredChannels) > 0 {
			builder.WriteString("**Preferred Channels:**\n")
			for _, channel := range profile.PreferredChannels {
				builder.WriteString(fmt.Sprintf("- %s\n", strings.TrimSpace(channel)))
			}
		}
	}

	return builder.String()
}

func (h *A2AHandler) sendSuccessResponse(c *gin.Context, id string, result interface{}) {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	c.JSON(http.StatusOK, response)
}

func (h *A2AHandler) sendErrorResponse(c *gin.Context, id string, message string, code int) {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
	c.JSON(http.StatusOK, response) // JSON-RPC errors are sent with 200 OK
}
