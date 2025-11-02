# Customer Profiler Agent

An AI-powered agent that generates detailed customer profiles based on business ideas, designed to integrate with the Telex.im messaging platform using the A2A protocol.

## Features

- Generates detailed customer profiles from business ideas
- Implements A2A (Agent-to-Agent) protocol
- Provides structured data output with demographics, pain points, and motivations
- Easy integration with Telex.im platform
- Powered by Google's Gemini AI

## Project Structure

```
customer-profiler-agent/
├── cmd/
│   ├── server/
│   │   └── main.go                 # Main server entry point
│   ├── test/
│       └── main.go                 # Test script main entry point
├── internal/
│   ├── a2a/
│   │   └── handler.go               # A2A protocol handlers
│   │   └── models.go               # A2A protocol types
│   ├── agent/
│   │   ├── agent.go               # Agent card loader
│   │   └── agent.json             # Agent configuration
│   │
│   ├── models/
│   │   └── customer.go              # Customer data models
│   └── profiler/
│       └── gemini.go              # Gemini AI client
├── go.mod
└── go.sum
```

## Installation

1. Clone the repository:
```bash
git clone https://github.com/BerylCAtieno/customer-profiler-agent.git
cd customer-profiler-agent
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
export GEMINI_API_KEY="your-gemini-api-key"
export PORT="8080" 
```

## Usage

### Running Locally

```bash
go run cmd/server/main.go
```

The agent will start on `http://localhost:8080` with the following endpoints:

- `/.well-known/agent.json` - Agent card endpoint
- `/a2a/profiler` - A2A protocol endpoint for profile generation
- `/health` - Health check endpoint

### Testing the Agent

#### Test Agent Card
```bash
curl http://localhost:8080/.well-known/agent.json
```

#### Test Profile Generation

Send a JSON-RPC request to generate profiles:

```bash
curl -X POST http://localhost:8080/a2a/profiler \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": "test-123",
    "method": "agent/task",
    "params": {
      "message": {
        "kind": "message",
        "role": "user",
        "parts": [
          {
            "kind": "text",
            "text": "Generate a customer profile for a sustainable fashion e-commerce platform"
          }
        ]
      },
      "configuration": {
        "blocking": true,
        "acceptedOutputModes": ["text", "data"]
      }
    }
  }'
```

## A2A Protocol

The agent implements the A2A (Agent-to-Agent) protocol for seamless integration with messaging platforms.

### Supported Methods

- `agent/task` - Main method for processing profile generation requests

### Message Format

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": "unique-request-id",
  "method": "agent/task",
  "params": {
    "message": {
      "kind": "message",
      "role": "user",
      "parts": [{"kind": "text", "text": "your business idea"}]
    },
    "configuration": {
      "blocking": true
    }
  }
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": "unique-request-id",
  "result": {
    "id": "task-id",
    "kind": "task",
    "status": {
      "state": "completed",
      "timestamp": "2025-01-15T10:30:00Z",
      "message": {
        "kind": "message",
        "role": "agent",
        "parts": [{"kind": "text", "text": "formatted profile..."}]
      }
    },
    "artifacts": [
      {
        "artifactId": "profile-task-id",
        "name": "Customer Profile Data",
        "parts": [{"kind": "data", "data": {...}}]
      }
    ]
  }
}
```

## Integration with Telex.im

1. Deploy your agent to a publicly accessible URL
2. Register your agent with Telex.im using [this](./workflow.json) configuration
3. Users can interact with your agent through the Telex.im messaging platform

## Deployment

### Docker
```bash
docker build -t customer-profiler-agent .
docker run -p 8080:8080 -e GEMINI_API_KEY=your-key customer-profiler-agent
```


## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.