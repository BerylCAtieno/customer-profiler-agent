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

	a2aHandler := a2a.NewA2AHandler(geminiClient)

	router := gin.Default()

	// Endpoints
	router.GET("/.well-known/agent.json", a2aHandler.ServeAgentCard)

	router.POST("/a2a/profiler", a2aHandler.HandleProfiler)

	router.GET("/health", func(c *gin.Context) {
		c.String(200, "OK")
	})

	// server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Customer Profiler Agent starting on port %s", port)
	log.Printf("Agent card available at: http://localhost:%s/.well-known/agent.json", port)
	log.Printf("A2A endpoint available at: http://localhost:%s/a2a/profiler", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
