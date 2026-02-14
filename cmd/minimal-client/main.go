// Package main provides an interactive xAI chat client for testing.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	xai "github.com/roelfdiedericks/xai-go"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Flags
	model := flag.String("model", "", "Model to use (default: grok-3)")
	system := flag.String("system", "You are a helpful assistant.", "System prompt")
	stream := flag.Bool("stream", true, "Stream responses")
	runTests := flag.Bool("test", false, "Run automated tests instead of interactive mode")
	flag.Parse()

	// Create client from environment
	client, err := xai.FromEnv()
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}
	defer client.Close()

	if *runTests {
		return runAutomatedTests(client)
	}

	return runInteractive(client, *model, *system, *stream)
}

// toolSettings tracks which tools are enabled
type toolSettings struct {
	webSearch     bool
	xSearch       bool
	codeExecution bool
}

func (t *toolSettings) String() string {
	var enabled []string
	if t.webSearch {
		enabled = append(enabled, "web")
	}
	if t.xSearch {
		enabled = append(enabled, "x")
	}
	if t.codeExecution {
		enabled = append(enabled, "code")
	}
	if len(enabled) == 0 {
		return "none"
	}
	return strings.Join(enabled, ", ")
}

func runInteractive(client *xai.Client, model, systemPrompt string, stream bool) error {
	if model == "" {
		model = client.DefaultModel()
	}
	imageModel := client.DefaultImageModel()

	// Enable server-side tools by default
	tools := &toolSettings{
		webSearch:     true,
		xSearch:       true,
		codeExecution: true,
	}

	fmt.Println("=== xAI Interactive Chat ===")
	fmt.Printf("Model: %s\n", model)
	fmt.Printf("Streaming: %v\n", stream)
	fmt.Printf("Tools: %s\n", tools)
	fmt.Println("Commands: /help, /model, /system, /stream, /tools, /image, /quit")
	fmt.Println("---")

	reader := bufio.NewReader(os.Stdin)
	var history []*message

	for {
		fmt.Print("\nYou: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("\nGoodbye!")
				return nil
			}
			return err
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Handle commands
		if strings.HasPrefix(input, "/") {
			switch {
			case input == "/quit" || input == "/exit" || input == "/q":
				fmt.Println("Goodbye!")
				return nil

			case input == "/help" || input == "/h":
				printHelp()
				continue

			case input == "/clear" || input == "/c":
				history = nil
				fmt.Println("Conversation cleared.")
				continue

			case input == "/stream" || input == "/s":
				stream = !stream
				fmt.Printf("Streaming: %v\n", stream)
				continue

			case input == "/info" || input == "/i":
				printInfo(client, model, systemPrompt, stream, len(history))
				continue

			case input == "/model" || input == "/m":
				if err := listModels(client); err != nil {
					fmt.Printf("Error: %v\n", err)
				}
				continue

			case strings.HasPrefix(input, "/model "):
				model = strings.TrimPrefix(input, "/model ")
				fmt.Printf("Model changed to: %s\n", model)
				continue

			case strings.HasPrefix(input, "/system "):
				systemPrompt = strings.TrimPrefix(input, "/system ")
				fmt.Printf("System prompt changed to: %s\n", systemPrompt)
				continue

			case input == "/image":
				fmt.Println("Usage: /image <prompt>")
				fmt.Println("Options: -wide, -tall, -2k, -model <name>")
				fmt.Println("Example: /image -wide a sunset over mountains")
				fmt.Printf("Current image model: %s\n", imageModel)
				continue

			case input == "/image-models" || input == "/im":
				if err := listImageModels(client); err != nil {
					fmt.Printf("Error: %v\n", err)
				}
				continue

			case input == "/image-model" || input == "/imodel":
				fmt.Printf("Current image model: %s\n", imageModel)
				fmt.Println("Use /image-models to list available models")
				continue

			case strings.HasPrefix(input, "/image-model "):
				imageModel = strings.TrimPrefix(input, "/image-model ")
				fmt.Printf("Image model changed to: %s\n", imageModel)
				continue

			case strings.HasPrefix(input, "/image "):
				prompt := strings.TrimPrefix(input, "/image ")
				if err := generateImage(client, imageModel, prompt); err != nil {
					fmt.Printf("Error: %v\n", err)
				}
				continue

			case input == "/tools" || input == "/t":
				fmt.Printf("Tools enabled: %s\n", tools)
				fmt.Println("Usage: /tools web|x|code|all|off")
				continue

			case strings.HasPrefix(input, "/tools "):
				arg := strings.TrimPrefix(input, "/tools ")
				switch arg {
				case "web":
					tools.webSearch = !tools.webSearch
					fmt.Printf("Web search: %v\n", tools.webSearch)
				case "x":
					tools.xSearch = !tools.xSearch
					fmt.Printf("X search: %v\n", tools.xSearch)
				case "code":
					tools.codeExecution = !tools.codeExecution
					fmt.Printf("Code execution: %v\n", tools.codeExecution)
				case "all":
					tools.webSearch = true
					tools.xSearch = true
					tools.codeExecution = true
					fmt.Println("All tools enabled")
				case "off", "none":
					tools.webSearch = false
					tools.xSearch = false
					tools.codeExecution = false
					fmt.Println("All tools disabled")
				default:
					fmt.Println("Unknown tool. Options: web, x, code, all, off")
				}
				continue

			default:
				fmt.Println("Unknown command. Type /help for available commands.")
				continue
			}
		}

		// Add user message to history
		history = append(history, &message{role: "user", content: input})

		// Build request
		req := xai.NewChatRequest().
			SystemMessage(systemPrompt).
			WithModel(model)

		// Add history
		for _, msg := range history {
			switch msg.role {
			case "user":
				req.UserMessage(msg.content)
			case "assistant":
				req.AssistantMessage(msg.content)
			}
		}

		// Add enabled tools
		if tools.webSearch {
			req.AddTool(xai.NewWebSearchTool())
		}
		if tools.xSearch {
			req.AddTool(xai.NewXSearchTool())
		}
		if tools.codeExecution {
			req.AddTool(xai.NewCodeExecutionTool())
		}

		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)

		fmt.Print("\nAssistant: ")

		var response string
		if stream {
			response, err = streamResponse(ctx, client, req)
		} else {
			response, err = blockingResponse(ctx, client, req)
		}
		cancel()

		if err != nil {
			fmt.Printf("\nError: %v\n", err)
			// Remove the failed user message from history
			history = history[:len(history)-1]
			continue
		}

		fmt.Println()

		// Add assistant response to history
		history = append(history, &message{role: "assistant", content: response})
	}
}

type message struct {
	role    string
	content string
}

func streamResponse(ctx context.Context, client *xai.Client, req *xai.ChatRequest) (string, error) {
	stream, err := client.StreamChat(ctx, req)
	if err != nil {
		return "", err
	}

	var content strings.Builder
	var toolCalls []*xai.ToolCallInfo
	var citations []string
	contentStarted := false
	hadTools := false

	for {
		chunk, err := stream.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return content.String(), err
		}

		// Announce tools as they arrive with status
		for _, tc := range chunk.ToolCalls {
			hadTools = true
			status := "pending"
			switch tc.Status {
			case xai.ToolCallStatusCompleted:
				status = "completed"
			case xai.ToolCallStatusFailed:
				status = "failed"
			}

			if tc.Function != nil {
				args := tc.Function.Arguments
				if len(args) > 80 {
					args = args[:77] + "..."
				}
				fmt.Printf("  Tool: %s [%s] %s\n", tc.Function.Name, status, args)
			} else if tc.ID != "" {
				fmt.Printf("  Tool: %s [%s]\n", tc.ID, status)
			}
		}

		// Print newline before first content if we had tool calls
		if chunk.Delta != "" && !contentStarted {
			if hadTools {
				fmt.Println()
			}
			contentStarted = true
		}

		fmt.Print(chunk.Delta)
		content.WriteString(chunk.Delta)

		// Collect tool calls and citations from chunks
		toolCalls = append(toolCalls, chunk.ToolCalls...)
		if len(chunk.Citations) > 0 {
			citations = chunk.Citations
		}
	}

	// Display tool calls if any
	displayToolCalls(toolCalls)
	displayCitations(citations)

	return content.String(), nil
}

func blockingResponse(ctx context.Context, client *xai.Client, req *xai.ChatRequest) (string, error) {
	resp, err := client.CompleteChat(ctx, req)
	if err != nil {
		return "", err
	}
	fmt.Print(resp.Content)

	// Display tool calls if any
	displayToolCalls(resp.ToolCalls)
	displayCitations(resp.Citations)

	return resp.Content, nil
}

func displayToolCalls(toolCalls []*xai.ToolCallInfo) {
	if len(toolCalls) == 0 {
		return
	}

	fmt.Println()
	fmt.Println("\n[Tool Calls]")
	for _, tc := range toolCalls {
		typeStr := "client"
		if tc.IsServerSide() {
			typeStr = "server"
		}

		statusStr := "pending"
		switch tc.Status {
		case xai.ToolCallStatusCompleted:
			statusStr = "completed"
		case xai.ToolCallStatusFailed:
			statusStr = "failed"
		}

		if tc.Function != nil {
			fmt.Printf("  - %s (%s, %s)\n", tc.Function.Name, typeStr, statusStr)
			if tc.Function.Arguments != "" && tc.Function.Arguments != "{}" {
				fmt.Printf("    Args: %s\n", tc.Function.Arguments)
			}
		} else {
			fmt.Printf("  - [%s] (%s, %s)\n", tc.ID, typeStr, statusStr)
		}

		if tc.ErrorMessage != "" {
			fmt.Printf("    Error: %s\n", tc.ErrorMessage)
		}
	}
}

func displayCitations(citations []string) {
	if len(citations) == 0 {
		return
	}

	fmt.Println("\n[Citations]")
	for i, c := range citations {
		fmt.Printf("  %d. %s\n", i+1, c)
	}
}

func printHelp() {
	fmt.Print(`
Commands:
  /help, /h            Show this help
  /quit, /exit, /q     Exit the chat
  /clear, /c           Clear conversation history
  /stream, /s          Toggle streaming mode
  /info, /i            Show current settings
  /model, /m           List available chat models
  /model <name>        Change the chat model
  /system <prompt>     Change the system prompt
  /tools, /t           Show enabled tools
  /tools <name>        Toggle tool (web, x, code, all, off)
  /image <prompt>      Generate an image (options: -wide, -tall, -2k)
  /image-model         Show current image model
  /image-model <name>  Change the image model
  /image-models, /im   List available image models
`)
}

func printInfo(client *xai.Client, model, systemPrompt string, stream bool, historyLen int) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println()
	fmt.Printf("Model: %s\n", model)
	fmt.Printf("System: %s\n", systemPrompt)
	fmt.Printf("Streaming: %v\n", stream)
	fmt.Printf("History: %d messages\n", historyLen)

	info, err := client.GetAPIKeyInfo(ctx)
	if err == nil {
		fmt.Printf("API Key: %s (Status: %s)\n", info.RedactedKey, info.Status)
	}
}

func listModels(client *xai.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	models, err := client.ListModels(ctx)
	if err != nil {
		return err
	}

	fmt.Println("\nAvailable models:")
	for _, m := range models {
		aliases := ""
		if len(m.Aliases) > 0 {
			aliases = fmt.Sprintf(" (aliases: %s)", strings.Join(m.Aliases, ", "))
		}
		fmt.Printf("  - %s%s\n", m.Name, aliases)
	}
	return nil
}

func listImageModels(client *xai.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	models, err := client.ListImageModels(ctx)
	if err != nil {
		return err
	}

	if len(models) == 0 {
		fmt.Println("\nNo image models available.")
		return nil
	}

	fmt.Println("\nAvailable image models:")
	for _, m := range models {
		aliases := ""
		if len(m.Aliases) > 0 {
			aliases = fmt.Sprintf(" (aliases: %s)", strings.Join(m.Aliases, ", "))
		}
		fmt.Printf("  - %s%s\n", m.Name, aliases)
	}
	return nil
}

func generateImage(client *xai.Client, defaultModel, input string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Parse options from input
	var prompt string
	model := defaultModel
	var aspectRatio *xai.ImageAspectRatio
	var resolution *xai.ImageResolution

	// Split input and parse flags
	remaining := input
parseLoop:
	for strings.HasPrefix(remaining, "-") {
		parts := strings.SplitN(remaining, " ", 2)
		if len(parts) < 2 {
			break
		}
		opt := parts[0]
		remaining = parts[1]

		switch opt {
		case "-wide", "-16x9":
			ar := xai.ImageAspectRatio16x9
			aspectRatio = &ar
		case "-tall", "-9x16":
			ar := xai.ImageAspectRatio9x16
			aspectRatio = &ar
		case "-4x3":
			ar := xai.ImageAspectRatio4x3
			aspectRatio = &ar
		case "-3x4":
			ar := xai.ImageAspectRatio3x4
			aspectRatio = &ar
		case "-2k", "-hd":
			res := xai.ImageResolution2K
			resolution = &res
		case "-model":
			// Next part is the model name
			modelParts := strings.SplitN(remaining, " ", 2)
			model = modelParts[0]
			if len(modelParts) > 1 {
				remaining = modelParts[1]
			} else {
				remaining = ""
			}
		default:
			// Unknown option, stop parsing - put opt back
			remaining = opt + " " + remaining
			break parseLoop
		}
	}
	prompt = strings.TrimSpace(remaining)

	if prompt == "" {
		return fmt.Errorf("no prompt provided")
	}

	req := xai.NewImageRequest(prompt).WithModel(model)
	if aspectRatio != nil {
		req.WithAspectRatio(*aspectRatio)
	}
	if resolution != nil {
		req.WithResolution(*resolution)
	}

	fmt.Println("Generating image...")

	resp, err := client.GenerateImage(ctx, req)
	if err != nil {
		return err
	}

	if len(resp.Images) == 0 {
		return fmt.Errorf("no images generated")
	}

	fmt.Printf("Model: %s\n", resp.Model)
	for i, img := range resp.Images {
		if img.URL != "" {
			fmt.Printf("Image %d: %s\n", i+1, img.URL)
		} else if img.Base64 != "" {
			fmt.Printf("Image %d: [base64 data, %d bytes]\n", i+1, len(img.Base64))
		}
	}

	return nil
}

func runAutomatedTests(client *xai.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Test 1: Get API key info
	fmt.Println("=== Getting API key info ===")
	info, err := client.GetAPIKeyInfo(ctx)
	if err != nil {
		return fmt.Errorf("getting API key info: %w", err)
	}
	fmt.Printf("Key: %s (ID: %s)\n", info.RedactedKey, info.KeyID)
	fmt.Printf("Status: %s\n", info.Status)
	fmt.Printf("ACLs: %v\n", info.ACLs)
	fmt.Println()

	// Test 2: List models
	fmt.Println("=== Listing models ===")
	models, err := client.ListModels(ctx)
	if err != nil {
		return fmt.Errorf("listing models: %w", err)
	}
	for _, m := range models {
		fmt.Printf("- %s (context: %d tokens)\n", m.Name, m.MaxPromptLength)
	}
	fmt.Println()

	// Test 3: Simple chat completion
	fmt.Println("=== Chat completion ===")
	req := xai.NewChatRequest().
		SystemMessage("You are a helpful assistant. Be concise.").
		UserMessage("What is the capital of France?").
		WithMaxTokens(100)

	resp, err := client.CompleteChat(ctx, req)
	if err != nil {
		return fmt.Errorf("chat completion: %w", err)
	}
	fmt.Printf("Response: %s\n", resp.Content)
	fmt.Printf("Tokens: prompt=%d, completion=%d\n",
		resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
	fmt.Println()

	// Test 4: Streaming chat
	fmt.Println("=== Streaming chat ===")
	streamReq := xai.NewChatRequest().
		SystemMessage("You are a helpful assistant.").
		UserMessage("Count from 1 to 5, one number per line.").
		WithMaxTokens(100)

	stream, err := client.StreamChat(ctx, streamReq)
	if err != nil {
		return fmt.Errorf("streaming chat: %w", err)
	}

	fmt.Print("Streaming: ")
	for {
		chunk, err := stream.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("stream error: %w", err)
		}
		fmt.Print(chunk.Delta)
	}
	fmt.Println()
	fmt.Println()

	// Test 5: Tokenization
	fmt.Println("=== Tokenization ===")
	text := "Hello, world! This is a test."
	tokenResp, err := client.Tokenize(ctx, client.DefaultModel(), text)
	if err != nil {
		return fmt.Errorf("tokenization: %w", err)
	}
	fmt.Printf("Text: %q\n", text)
	fmt.Printf("Token count: %d\n", tokenResp.TokenCount())
	fmt.Println()

	fmt.Println("=== All tests passed! ===")
	return nil
}
