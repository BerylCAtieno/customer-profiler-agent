package main

import (
	"log"
	"os"

	"github.com/BerylCAtieno/customer-profiler-agent/internal/a2a"
	"github.com/BerylCAtieno/customer-profiler-agent/internal/profiler"
	"github.com/gin-gonic/gin"
)

func main() {

	// Get API key from environment
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable is required")
	}

	// Initialize Gemini client
	geminiClient, err := profiler.NewGeminiClient(apiKey)
	if err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
	}
	defer geminiClient.Close()

	// Create handler
	a2aHandler := a2a.NewA2AHandler(geminiClient)

	// Setup Gin router
	router := gin.Default()

	// Agent card endpoint
	router.GET("/.well-known/agent.json", a2aHandler.ServeAgentCard)

	// A2A protocol endpoint
	router.POST("/a2a/profiler", a2aHandler.HandleProfiler)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.String(200, "OK")
	})

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Customer Profiler Agent starting on port %s", port)
	log.Printf("Agent card available at: http://localhost:%s/.well-known/agent.json", port)
	log.Printf("A2A endpoint available at: http://localhost:%s/a2a/profiler", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
