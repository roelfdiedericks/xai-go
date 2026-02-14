# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.0] - 2026-02-14

### Added

- **Tool call history reconstruction** - `AssistantContent.ToolCalls` for representing assistant tool calls in conversation history
- **`HistoryToolCall` struct** - Represents tool calls with ID, Name, and Arguments
- **Integration test** - `TestChatWithToolCallHistory` validates tool call history reconstruction

### Changed

- **BREAKING: Struct-based message content API** - All message methods now use typed content structs:
  - `SystemMessage(SystemContent{Text: "..."})` instead of `SystemMessage("...")`
  - `UserMessage(UserContent{Text: "...", ImageURL: "..."})` instead of `UserMessage("...")` / `UserWithImage(...)`
  - `AssistantMessage(AssistantContent{Text: "...", ToolCalls: [...]})` instead of `AssistantMessage("...")`
  - `DeveloperMessage(DeveloperContent{Text: "..."})` instead of `DeveloperMessage("...")`
  - `ToolResult(ToolContent{CallID: "...", Result: "..."})` instead of `ToolResult("...", "...")`
- Removed `UserWithImage()` - use `UserMessage(UserContent{Text: "...", ImageURL: "..."})` instead

## [0.3.0] - 2026-02-14

### Added

- **Server-side conversation context** - `WithPreviousResponseId()` for efficient multi-turn conversations
- **Message storage control** - `WithStoreMessages()` to enable/disable server-side context storage
- **Encrypted content support** - `WithEncryptedContent()` for reasoning trace preservation
- **`/context` command** - Toggle between response_id (server) and history (client) context modes
- **`/context store` command** - Toggle server-side message storage in interactive client
- **Integration tests** - Tests for multi-turn conversations with and without `previous_response_id`

### Changed

- Interactive client now defaults to `response_id` mode with storage enabled
- Response ID captured and chained automatically in interactive client

## [0.2.0] - 2026-02-14

### Added

- **Configurable keepalive** - `KeepaliveTime`, `KeepaliveTimeout`, `KeepalivePermitWithoutStream` in Config
- **Interactive client tools** - Server-side tools (web search, X search, code execution) enabled by default
- **Tool display** - Real-time tool call visibility during streaming with status and arguments
- **`/tools` command** - Toggle tools on/off in interactive client
- **`/image-model` command** - Change image model in interactive client

### Changed

- Default image format now explicitly set to URL (fixes "Invalid format" error)
- Tool calls displayed inline during streaming with `[pending]`/`[completed]` status

## [0.1.0] - 2026-02-14

### Added

- Initial release
- **Chat completions** - Blocking and streaming responses via gRPC
- **Tool support** - Function calling, web search, X search, code execution, collections search, attachment search, MCP
- **Image generation** - Generate images with configurable aspect ratio, resolution, and format
- **Embeddings** - Text and image embeddings
- **Tokenization** - Token counting and inspection
- **Model management** - List and query language, embedding, and image models
- **Document search** - Search within document collections
- **Deferred completions** - Async request processing with polling
- **Stored completions** - Retrieve and delete stored completions
- **API key info** - Query API key status and permissions
- **Secure API key handling** - Memory-zeroing SecureString wrapper
- **Structured errors** - Error types with codes, retryability, and gRPC status mapping
- **Interactive test client** - REPL for manual testing with chat and image generation

[0.4.0]: https://github.com/roelfdiedericks/xai-go/releases/tag/v0.4.0
[0.3.0]: https://github.com/roelfdiedericks/xai-go/releases/tag/v0.3.0
[0.2.0]: https://github.com/roelfdiedericks/xai-go/releases/tag/v0.2.0
[0.1.0]: https://github.com/roelfdiedericks/xai-go/releases/tag/v0.1.0
