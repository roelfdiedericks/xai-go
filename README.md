# xai-go

Go client library for the xAI gRPC API.

## Installation

```bash
go get github.com/roelfdiedericks/xai-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    xai "github.com/roelfdiedericks/xai-go"
)

func main() {
    // Create client from environment variable XAI_APIKEY
    client, err := xai.FromEnv()
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()

    // Simple chat completion
    req := xai.NewChatRequest().
        SystemMessage("You are a helpful assistant.").
        UserMessage("What is the capital of France?")

    resp, err := client.CompleteChat(ctx, req)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(resp.Content)
}
```

## Features

- **Chat completions** - Blocking and streaming responses
- **Tool support** - Function calling, web search, X search, code execution, and more
- **Model management** - List and get model information
- **Embeddings** - Generate text and image embeddings
- **Tokenization** - Count and inspect tokens
- **Image generation** - Generate images from text prompts
- **Document search** - Search within document collections
- **Deferred completions** - Async request processing

## Configuration

```go
// From environment variable
client, err := xai.FromEnv()

// With explicit configuration
client, err := xai.New(xai.Config{
    APIKey:       xai.NewSecureString("your-api-key"),
    Endpoint:     "api.x.ai:443",                 // optional
    Timeout:      120 * time.Second,              // optional
    DefaultModel: "grok-4-1-fast-reasoning",      // optional
})
```

## Chat Completions

### Blocking

```go
req := xai.NewChatRequest().
    SystemMessage("You are a helpful assistant.").
    UserMessage("Hello!").
    WithMaxTokens(100).
    WithTemperature(0.7)

resp, err := client.CompleteChat(ctx, req)
// Uses client.DefaultModel() if WithModel() not called
```

### Streaming

```go
stream, err := client.StreamChat(ctx, req)
if err != nil {
    log.Fatal(err)
}

for {
    chunk, err := stream.Next()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatal(err)
    }
    fmt.Print(chunk.Delta)
}
```

### Multi-Turn Conversations with Server-Side Context

Instead of sending the full conversation history with each request, you can use xAI's server-side context storage with `previous_response_id`. This is more efficient and required for preserving reasoning traces in reasoning models.

```go
// First turn: enable storage
req1 := xai.NewChatRequest().
    SystemMessage("You are a helpful assistant.").
    UserMessage("My name is Alice.").
    WithStoreMessages(true)  // Enable server-side storage

resp1, err := client.CompleteChat(ctx, req1)
// resp1.ID contains the response ID

// Second turn: reference previous response, only send new message
req2 := xai.NewChatRequest().
    WithPreviousResponseId(resp1.ID).  // Chain from previous
    WithStoreMessages(true).            // Keep storing for further turns
    UserMessage("What is my name?")

resp2, err := client.CompleteChat(ctx, req2)
// The model remembers "Alice" from server-side context
```

Key points:
- `WithStoreMessages(true)` must be set on **each turn** to continue the chain
- Responses are stored for 30 days
- You still pay for the full context, but benefit from cached prompt tokens
- Use `WithEncryptedContent(true)` for reasoning trace preservation

## Tool Calling

```go
// Define a function tool
addTool := xai.NewFunctionTool("add", "Add two numbers").
    WithParameters(`{
        "type": "object",
        "properties": {
            "a": {"type": "number"},
            "b": {"type": "number"}
        },
        "required": ["a", "b"]
    }`)

req := xai.NewChatRequest().
    UserMessage("What is 2 + 3?").
    AddTool(addTool).
    WithToolChoice(xai.ToolChoiceAuto)

resp, err := client.CompleteChat(ctx, req)
if resp.HasToolCalls() {
    for _, tc := range resp.ToolCalls {
        if tc.Function != nil {
            fmt.Printf("Call %s with args: %s\n", 
                tc.Function.Name, tc.Function.Arguments)
        }
    }
}
```

### Built-in Tools

```go
// Web search
req.AddTool(xai.NewWebSearchTool())

// X (Twitter) search
req.AddTool(xai.NewXSearchTool())

// Code execution
req.AddTool(xai.NewCodeExecutionTool())

// Collections search
req.AddTool(xai.NewCollectionsSearchTool("collection-id-1"))
```

## Image Generation

```go
req := xai.NewImageRequest("a futuristic cityscape at sunset").
    WithAspectRatio(xai.ImageAspectRatio16x9).
    WithResolution(xai.ImageResolution2K)
// Uses xai.DefaultImageModel ("grok-2-image") if WithModel() not called

resp, err := client.GenerateImage(ctx, req)
if err != nil {
    log.Fatal(err)
}

for _, img := range resp.Images {
    fmt.Println("Image URL:", img.URL)
}
```

Available options:
- `WithModel(name)` - specify image model (default: `grok-2-image`)
- `WithAspectRatio()` - `1x1`, `16x9`, `9x16`, `4x3`, `3x4`
- `WithResolution()` - `1K` (default), `2K`
- `WithFormat()` - `URL` (default), `Base64`
- `WithCount(n)` - number of images (1-10)

## Error Handling

```go
resp, err := client.CompleteChat(ctx, req)
if err != nil {
    var xaiErr *xai.Error
    if errors.As(err, &xaiErr) {
        switch xaiErr.Code {
        case xai.ErrAuth:
            log.Fatal("Authentication failed")
        case xai.ErrRateLimit:
            log.Printf("Rate limited, retry after: %v", xaiErr.RetryAfter)
        default:
            log.Fatal(xaiErr)
        }
    }
}
```

## Development

### Prerequisites

- Go 1.23+
- [buf](https://buf.build) CLI for proto generation

### Setup

```bash
# Clone with submodules
git clone --recursive https://github.com/roelfdiedericks/xai-go.git

# Or update submodules if already cloned
git submodule update --init --recursive

# Generate proto code
make proto

# Build
make build
```

### Makefile Targets

| Target | Description |
|--------|-------------|
| `make proto` | Generate Go code from xai-proto (only if sources changed) |
| `make proto-force` | Force regenerate proto code |
| `make build` | Build the library |
| `make test` | Run unit tests |
| `make test-integration` | Run integration tests (requires XAI_APIKEY) |
| `make test-interactive` | Start interactive chat REPL (requires XAI_APIKEY) |
| `make test-automated` | Run automated API verification tests |
| `make test-all` | Run all tests |
| `make lint` | Run golangci-lint |
| `make audit` | Run lint + govulncheck |
| `make tidy` | Run go mod tidy |
| `make submodule-update` | Update xai-proto submodule to latest |
| `make install-buf` | Install buf CLI |

### Interactive Testing

The `test-interactive` target starts a chat REPL for manual testing:

```bash
make test-interactive
```

Commands within the REPL:
- `/help` - Show available commands
- `/model` - List available chat models
- `/model <name>` - Change the chat model
- `/system <prompt>` - Change system prompt
- `/stream` - Toggle streaming mode
- `/clear` - Clear conversation history
- `/info` - Show current settings
- `/context` - Show context mode (response_id vs history)
- `/context mode` - Toggle between server-side (response_id) and client-side (history) context
- `/context store` - Toggle server-side message storage
- `/tools` - Show enabled tools
- `/tools <name>` - Toggle tool (web, x, code, all, off)
- `/image <prompt>` - Generate an image (options: `-wide`, `-tall`, `-2k`)
- `/image-model` - Show/change image model
- `/image-models` - List available image models
- `/quit` - Exit

You can also run it directly with flags:

```bash
go run ./cmd/minimal-client -model grok-4-1-fast-reasoning -system "You are a pirate" -stream=false
```

### Project Layout

```
xai-go/
├── xai-proto/          # Git submodule — xAI protobuf definitions
├── proto/              # Generated Go code (gitignored)
│   └── xai/api/v1/     # Chat, models, auth, embed, etc.
├── tests/              # Unit tests
├── integration/        # Integration tests (require API key)
├── cmd/minimal-client/ # Interactive chat REPL
├── *.go                # Library source files
├── buf.gen.go.yaml     # Buf generation config
├── Makefile
└── go.mod
```

## API Coverage

This library provides full coverage of the xAI gRPC API:

| Service | RPCs |
|---------|------|
| Chat | GetCompletion, GetCompletionChunk, StartDeferredCompletion, GetDeferredCompletion, GetStoredCompletion, DeleteStoredCompletion |
| Models | ListLanguageModels, GetLanguageModel, ListEmbeddingModels, GetEmbeddingModel, ListImageGenerationModels, GetImageGenerationModel |
| Embedder | Embed |
| Tokenize | TokenizeText |
| Auth | GetApiKeyInfo |
| Sample | SampleText, SampleTextStreaming |
| Image | GenerateImage |
| Documents | Search |

## License

See [LICENSE](LICENSE) for details.
