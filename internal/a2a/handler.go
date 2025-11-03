package a2a

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/BerylCAtieno/customer-profiler-agent/internal/agent"
	"github.com/BerylCAtieno/customer-profiler-agent/internal/models"
	"github.com/BerylCAtieno/customer-profiler-agent/internal/profiler"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type A2AHandler struct {
	geminiClient *profiler.GeminiClient
}

func NewA2AHandler(geminiClient *profiler.GeminiClient) *A2AHandler {
	return &A2AHandler{
		geminiClient: geminiClient,
	}
}

// RequestLoggingMiddleware logs all incoming requests
func RequestLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read the body
		bodyBytes, _ := io.ReadAll(c.Request.Body)

		// Log the raw request
		log.Printf("=== INCOMING REQUEST ===")
		log.Printf("Method: %s", c.Request.Method)
		log.Printf("Path: %s", c.Request.URL.Path)
		log.Printf("Headers: %v", c.Request.Header)
		log.Printf("Body: %s", string(bodyBytes))
		log.Printf("========================")

		// Restore the body for the handler
		c.Request.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

		c.Next()

		// Log response status
		log.Printf("=== RESPONSE ===")
		log.Printf("Status: %d", c.Writer.Status())
		log.Printf("================")
	}
}

// HandleProfiler processes A2A messages
func (h *A2AHandler) HandleProfiler(c *gin.Context) {
	// Read and log the raw body first
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("ERROR: Failed to read request body: %v", err)
		h.sendErrorResponse(c, "", "Failed to read request body", -32700)
		return
	}

	log.Printf("=== RAW REQUEST BODY ===")
	log.Printf("%s", string(bodyBytes))
	log.Printf("========================")

	// Restore body for JSON parsing
	c.Request.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	// Try to parse as raw JSON first to see structure
	var rawJSON map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &rawJSON); err == nil {
		log.Printf("=== PARSED JSON STRUCTURE ===")
		prettyJSON, _ := json.MarshalIndent(rawJSON, "", "  ")
		log.Printf("%s", string(prettyJSON))
		log.Printf("============================")
	}

	// Restore body again for binding
	c.Request.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	// Parse JSON-RPC request
	var rpcReq JSONRPCRequest
	if err := c.ShouldBindJSON(&rpcReq); err != nil {
		log.Printf("ERROR: Failed to decode request as JSON-RPC: %v", err)
		log.Printf("Trying alternative direct message parsing...")

		// Try parsing without JSON-RPC wrapper
		h.handleDirectMessage(c, bodyBytes)
		return
	}

	log.Printf("=== PARSED RPC REQUEST ===")
	log.Printf("JSONRPC: %s", rpcReq.JSONRPC)
	log.Printf("ID: %s", rpcReq.ID)
	log.Printf("Method: %s", rpcReq.Method)
	log.Printf("Params: %+v", rpcReq.Params)
	log.Printf("==========================")

	// Validate JSON-RPC version
	if rpcReq.JSONRPC != "2.0" {
		log.Printf("WARN: Invalid JSON-RPC version: %s", rpcReq.JSONRPC)
		h.sendErrorResponse(c, rpcReq.ID, "Invalid JSON-RPC version", -32600)
		return
	}

	// Handle different methods
	switch rpcReq.Method {
	case "agent/task":
		h.handleTask(c, rpcReq)
	case "message/send":
		h.handleTask(c, rpcReq)
	default:
		log.Printf("ERROR: Unknown method: %s", rpcReq.Method)
		h.sendErrorResponse(c, rpcReq.ID, fmt.Sprintf("Method not found: %s", rpcReq.Method), -32601)
	}
}

// handleDirectMessage tries to handle message without JSON-RPC wrapper
func (h *A2AHandler) handleDirectMessage(c *gin.Context, bodyBytes []byte) {
	log.Printf("STATE: ATTEMPTING DIRECT MESSAGE PARSE")

	var msgParams MessageParams
	if err := json.Unmarshal(bodyBytes, &msgParams); err != nil {
		log.Printf("ERROR: Failed to parse as direct message: %v", err)
		h.sendErrorResponse(c, "", "Invalid request format", -32700)
		return
	}

	log.Printf("Successfully parsed as direct message")

	// Extract business idea
	businessIdea := h.extractBusinessIdea(msgParams.Message)
	log.Printf("Extracted business idea: '%s'", businessIdea)

	if businessIdea == "" {
		result := h.createErrorTaskResult(
			"direct-message",
			"Please provide a business idea to generate customer profiles.",
		)
		h.sendSuccessResponse(c, "direct-message", result)
		return
	}

	log.Printf("STATE: Calling Gemini client to generate profiles for: %s", businessIdea)
	// Generate customer profiles
	ctx := context.Background()
	profileResp, err := h.geminiClient.GenerateCustomerProfiles(ctx, businessIdea)
	if err != nil {
		log.Printf("ERROR: Failed to generate profiles: %v", err)
		result := h.createErrorTaskResult(
			"direct-message",
			fmt.Sprintf("Failed to generate customer profiles: %v", err),
		)
		h.sendSuccessResponse(c, "direct-message", result)
		return
	}

	log.Printf("STATE: Profile generation succeeded. Sending StateCompleted TaskResult.")
	// Create successful task result
	result := h.createSuccessTaskResult("direct-message", profileResp)
	h.sendSuccessResponse(c, "direct-message", result)
}

func (h *A2AHandler) handleTask(c *gin.Context, rpcReq JSONRPCRequest) {
	log.Printf("STATE: HANDLING JSON-RPC TASK")

	// Parse message parameters
	paramsJSON, err := json.Marshal(rpcReq.Params)
	if err != nil {
		log.Printf("ERROR: Failed to marshal params: %v", err)
		h.sendErrorResponse(c, rpcReq.ID, "Failed to parse parameters", -32602)
		return
	}

	var msgParams MessageParams
	if err := json.Unmarshal(paramsJSON, &msgParams); err != nil {
		log.Printf("ERROR: Failed to unmarshal params: %v", err)
		log.Printf("Params structure: %+v", rpcReq.Params)
		h.sendErrorResponse(c, rpcReq.ID, "Invalid parameters", -32602)
		return
	}

	// Extract business idea from user message
	businessIdea := h.extractBusinessIdea(msgParams.Message)
	log.Printf("Extracted business idea: '%s'", businessIdea)

	if businessIdea == "" {
		log.Printf("WARN: No business idea found in message")
		result := h.createErrorTaskResult(
			rpcReq.ID,
			"Please provide a business idea to generate customer profiles.",
		)
		h.sendSuccessResponse(c, rpcReq.ID, result)
		return
	}

	log.Printf("STATE: Calling Gemini client to generate profiles for: %s", businessIdea)

	// Generate customer profiles
	ctx := context.Background()
	profileResp, err := h.geminiClient.GenerateCustomerProfiles(ctx, businessIdea)
	if err != nil {
		log.Printf("ERROR: Failed to generate profiles: %v", err)
		result := h.createErrorTaskResult(
			rpcReq.ID,
			fmt.Sprintf("Failed to generate customer profiles: %v", err),
		)
		h.sendSuccessResponse(c, rpcReq.ID, result)
		return
	}

	log.Printf("Successfully generated %d profile(s)", len(profileResp.Profiles))

	log.Printf("STATE: Profile generation succeeded. Sending StateCompleted TaskResult.")
	// Create successful task result
	result := h.createSuccessTaskResult(rpcReq.ID, profileResp)

	h.sendSuccessResponse(c, rpcReq.ID, result)
}

// ServeAgentCard serves the agent card using Gin
func (h *A2AHandler) ServeAgentCard(c *gin.Context) {
	if err := agent.LoadAgentCard(); err != nil {
		log.Printf("ERROR: Error loading agent card: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Agent card not available"})
		return
	}

	log.Printf("Serving agent card")
	c.Data(http.StatusOK, "application/json", agent.AgentCardData)
}

func (h *A2AHandler) extractBusinessIdea(msg A2AMessage) string {
	var texts []string

	for _, part := range msg.Parts {
		// Handle direct text parts
		if part.Kind == "text" && part.Text != nil {
			// Since Text is interface{}, we need to type assert to string
			if textStr, ok := part.Text.(string); ok && textStr != "" {
				texts = append(texts, textStr)
			}
		}

		// Handle data parts containing conversation history
		if part.Kind == "data" && part.Data != nil {
			var dataArray []map[string]interface{}
			var dataBytes []byte
			var err error

			// part.Data could be json.RawMessage, []byte, or already unmarshaled
			switch v := part.Data.(type) {
			case json.RawMessage:
				dataBytes = v
			case []byte:
				dataBytes = v
			case string:
				dataBytes = []byte(v)
			case []map[string]interface{}:
				// Already unmarshaled
				dataArray = v
			default:
				// Try to marshal and unmarshal
				dataBytes, err = json.Marshal(v)
				if err != nil {
					log.Printf("WARN: Failed to marshal data part: %v", err)
					continue
				}
			}

			// If we have bytes, unmarshal them
			if len(dataBytes) > 0 {
				if err := json.Unmarshal(dataBytes, &dataArray); err != nil {
					log.Printf("WARN: Failed to unmarshal data part: %v", err)
					continue
				}
			}

			// Extract text from the last message in the data array (most recent user message)
			// Look for user messages that look like requests (not system responses)
			for i := len(dataArray) - 1; i >= 0; i-- {
				item := dataArray[i]
				if kind, ok := item["kind"].(string); ok && kind == "text" {
					if text, ok := item["text"].(string); ok && text != "" {
						// Clean HTML tags if present
						cleanText := strings.TrimSpace(text)
						cleanText = strings.ReplaceAll(cleanText, "<p>", "")
						cleanText = strings.ReplaceAll(cleanText, "</p>", "")
						cleanText = strings.TrimSpace(cleanText)

						// Skip if it looks like a system response (contains "Generating", "Creating", etc.)
						if strings.Contains(strings.ToLower(cleanText), "generating") ||
							strings.Contains(strings.ToLower(cleanText), "creating") ||
							cleanText == "." || cleanText == ".." || cleanText == "..." ||
							cleanText == "ce..." {
							continue
						}

						// This looks like a user request
						if cleanText != "" {
							texts = append(texts, cleanText)
							break // Use the most recent valid user message
						}
					}
				}
			}
		}
	}

	result := strings.TrimSpace(strings.Join(texts, " "))
	log.Printf("extractBusinessIdea: '%s'", result)
	return result
}

func (h *A2AHandler) createSuccessTaskResult(taskID string, profileResp *models.ProfileResponse) TaskResult {
	// Format the profile data nicely
	responseText := h.formatProfileResponse(profileResp)

	artifactID := uuid.New().String()
	messageID := uuid.New().String()

	return TaskResult{
		ID:   taskID,
		Kind: "task",
		Status: TaskStatus{
			State:     StateCompleted,
			Timestamp: Timestamp(),
			Message: &A2AMessage{
				Kind:      "message",
				Role:      RoleAgent,
				MessageID: messageID,
				TaskID:    taskID,
				Parts: []MessagePart{
					TextPart(responseText),
				},
			},
		},
		Artifacts: []Artifact{
			{
				ArtifactID: artifactID,
				Name:       "Customer Profile Data",
				Parts: []MessagePart{
					TextPart(responseText),
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
				Kind: "text",
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
		builder.WriteString(fmt.Sprintf("- Income: %s\n", profile.Income))

		if len(profile.PainPoints) > 0 {
			builder.WriteString("\n**Pain Points:**\n") // Single \n before section
			for _, pp := range profile.PainPoints {
				builder.WriteString(fmt.Sprintf("- %s\n", strings.TrimSpace(pp)))
			}
		}

		if len(profile.Motivations) > 0 {
			builder.WriteString("\n**Motivations:**\n") // Single \n before section
			for _, m := range profile.Motivations {
				builder.WriteString(fmt.Sprintf("- %s\n", strings.TrimSpace(m)))
			}
		}

		if len(profile.Interests) > 0 {
			builder.WriteString("\n**Interests:**\n") // Single \n before section
			for _, interest := range profile.Interests {
				builder.WriteString(fmt.Sprintf("- %s\n", strings.TrimSpace(interest)))
			}
		}

		if len(profile.PreferredChannels) > 0 {
			builder.WriteString("\n**Preferred Channels:**\n") // Single \n before section
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

	log.Printf("=== SENDING RESPONSE (Status 200) ===")
	responseJSON, _ := json.MarshalIndent(response, "", "  ")
	log.Printf("%s", string(responseJSON))
	log.Printf("====================================")

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

	log.Printf("=== SENDING RPC ERROR RESPONSE (Status 200) ===")
	log.Printf("Code: %d, Message: %s", code, message)
	log.Printf("==============================================")

	c.JSON(http.StatusOK, response) // JSON-RPC errors are sent with 200 OK
}
