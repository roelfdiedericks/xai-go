// Package xai provides a Go client for the xAI gRPC API.
//
// The client supports chat completions (blocking and streaming), tool calling,
// image generation, embeddings, tokenization, and more.
//
// # Quick Start
//
//	client, err := xai.FromEnv()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	req := xai.NewChatRequest().
//	    SystemMessage("You are a helpful assistant.").
//	    UserMessage("Hello!")
//
//	resp, err := client.CompleteChat(context.Background(), req)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(resp.Content)
//
// # Authentication
//
// Set the XAI_APIKEY environment variable and use [FromEnv], or pass an API key
// explicitly via [Config]:
//
//	client, err := xai.New(xai.Config{
//	    APIKey: xai.NewSecureString("your-api-key"),
//	})
//
// # Chat Completions
//
// Use [NewChatRequest] to build chat requests with a fluent API:
//
//	req := xai.NewChatRequest().
//	    SystemMessage("You are a helpful assistant.").
//	    UserMessage("What is the capital of France?").
//	    WithMaxTokens(100).
//	    WithTemperature(0.7)
//
//	// Blocking
//	resp, err := client.CompleteChat(ctx, req)
//
//	// Streaming
//	stream, err := client.StreamChat(ctx, req)
//	for {
//	    chunk, err := stream.Next()
//	    if err == io.EOF {
//	        break
//	    }
//	    fmt.Print(chunk.Delta)
//	}
//
// # Multi-Turn Conversations
//
// For multi-turn conversations, use [ChatRequest.WithPreviousResponseId] to chain
// responses using xAI's server-side context storage instead of resending full history:
//
//	// First turn: enable storage
//	req1 := xai.NewChatRequest().
//	    SystemMessage("You are helpful.").
//	    UserMessage("My name is Bob.").
//	    WithStoreMessages(true)
//
//	resp1, _ := client.CompleteChat(ctx, req1)
//
//	// Second turn: chain from previous response
//	req2 := xai.NewChatRequest().
//	    WithPreviousResponseId(resp1.ID).
//	    WithStoreMessages(true).
//	    UserMessage("What is my name?")
//
//	resp2, _ := client.CompleteChat(ctx, req2)
//	// resp2 will reference "Bob" from server-side context
//
// Note: [ChatRequest.WithStoreMessages](true) must be set on each turn to continue the chain.
// Responses are stored for 30 days.
//
// # Tool Calling
//
// Define tools using [NewFunctionTool], [NewWebSearchTool], [NewXSearchTool], etc:
//
//	tool := xai.NewFunctionTool("get_weather", "Get current weather").
//	    WithParameters(`{"type": "object", "properties": {"city": {"type": "string"}}}`)
//
//	req := xai.NewChatRequest().
//	    UserMessage("What's the weather in Paris?").
//	    AddTool(tool).
//	    WithToolChoice(xai.ToolChoiceAuto)
//
// # Image Generation
//
// Generate images using [NewImageRequest]:
//
//	req := xai.NewImageRequest("a sunset over mountains").
//	    WithAspectRatio(xai.ImageAspectRatio16x9)
//
//	resp, err := client.GenerateImage(ctx, req)
//	fmt.Println(resp.Images[0].URL)
//
// # Error Handling
//
// Errors are wrapped as [*Error] with structured information:
//
//	resp, err := client.CompleteChat(ctx, req)
//	if err != nil {
//	    var xaiErr *xai.Error
//	    if errors.As(err, &xaiErr) {
//	        if xaiErr.Code == xai.ErrRateLimit {
//	            // Handle rate limiting
//	        }
//	    }
//	}
package xai
