package xai

import (
	"encoding/json"

	v1 "github.com/roelfdiedericks/xai-go/proto/xai/api/v1"
)

// Tool represents a tool that can be used by the model.
// Use the specific tool constructors (NewFunctionTool, NewWebSearchTool, etc.)
// to create tools.
type Tool interface {
	toProto() *v1.Tool
}

// ToolChoice controls how the model uses tools.
type ToolChoice int

const (
	// ToolChoiceAuto lets the model decide whether to use tools.
	ToolChoiceAuto ToolChoice = iota
	// ToolChoiceNone prevents the model from using tools.
	ToolChoiceNone
	// ToolChoiceRequired forces the model to use at least one tool.
	ToolChoiceRequired
)

func (tc ToolChoice) toProto() *v1.ToolChoice {
	switch tc {
	case ToolChoiceNone:
		return &v1.ToolChoice{
			ToolChoice: &v1.ToolChoice_Mode{
				Mode: v1.ToolMode_TOOL_MODE_NONE,
			},
		}
	case ToolChoiceRequired:
		return &v1.ToolChoice{
			ToolChoice: &v1.ToolChoice_Mode{
				Mode: v1.ToolMode_TOOL_MODE_REQUIRED,
			},
		}
	default: // Auto
		return &v1.ToolChoice{
			ToolChoice: &v1.ToolChoice_Mode{
				Mode: v1.ToolMode_TOOL_MODE_AUTO,
			},
		}
	}
}

// FunctionTool represents a function that can be called by the model.
type FunctionTool struct {
	// Name is the function name.
	Name string
	// Description describes what the function does.
	Description string
	// Parameters is a JSON Schema describing the function parameters.
	Parameters json.RawMessage
	// Strict enables strict schema validation (not yet supported by xAI).
	Strict bool
}

// NewFunctionTool creates a new function tool with the given name and description.
func NewFunctionTool(name, description string) *FunctionTool {
	return &FunctionTool{
		Name:        name,
		Description: description,
	}
}

// WithParameters sets the function parameters schema.
func (f *FunctionTool) WithParameters(params any) *FunctionTool {
	switch p := params.(type) {
	case json.RawMessage:
		f.Parameters = p
	case []byte:
		f.Parameters = p
	case string:
		f.Parameters = json.RawMessage(p)
	default:
		// Try to marshal as JSON
		b, err := json.Marshal(p)
		if err == nil {
			f.Parameters = b
		}
	}
	return f
}

func (f *FunctionTool) toProto() *v1.Tool {
	fn := &v1.Function{
		Name:        f.Name,
		Description: f.Description,
		Strict:      f.Strict,
	}
	if f.Parameters != nil {
		fn.Parameters = string(f.Parameters)
	}
	return &v1.Tool{
		Tool: &v1.Tool_Function{
			Function: fn,
		},
	}
}

// WebSearchTool enables web search capabilities.
type WebSearchTool struct{}

// NewWebSearchTool creates a new web search tool.
func NewWebSearchTool() *WebSearchTool {
	return &WebSearchTool{}
}

func (w *WebSearchTool) toProto() *v1.Tool {
	return &v1.Tool{
		Tool: &v1.Tool_WebSearch{
			WebSearch: &v1.WebSearch{},
		},
	}
}

// XSearchTool enables X (Twitter) search capabilities.
type XSearchTool struct{}

// NewXSearchTool creates a new X search tool.
func NewXSearchTool() *XSearchTool {
	return &XSearchTool{}
}

func (x *XSearchTool) toProto() *v1.Tool {
	return &v1.Tool{
		Tool: &v1.Tool_XSearch{
			XSearch: &v1.XSearch{},
		},
	}
}

// CodeExecutionTool enables code execution in a sandboxed environment.
type CodeExecutionTool struct{}

// NewCodeExecutionTool creates a new code execution tool.
func NewCodeExecutionTool() *CodeExecutionTool {
	return &CodeExecutionTool{}
}

func (c *CodeExecutionTool) toProto() *v1.Tool {
	return &v1.Tool{
		Tool: &v1.Tool_CodeExecution{
			CodeExecution: &v1.CodeExecution{},
		},
	}
}

// CollectionsSearchTool enables searching within document collections.
type CollectionsSearchTool struct {
	// CollectionIDs specifies which collections to search.
	CollectionIDs []string
}

// NewCollectionsSearchTool creates a new collections search tool.
func NewCollectionsSearchTool(collectionIDs ...string) *CollectionsSearchTool {
	return &CollectionsSearchTool{
		CollectionIDs: collectionIDs,
	}
}

func (c *CollectionsSearchTool) toProto() *v1.Tool {
	return &v1.Tool{
		Tool: &v1.Tool_CollectionsSearch{
			CollectionsSearch: &v1.CollectionsSearch{
				CollectionIds: c.CollectionIDs,
			},
		},
	}
}

// AttachmentSearchTool enables searching within attachments.
type AttachmentSearchTool struct {
	// Limit is the maximum number of files to search (optional).
	Limit *int32
}

// NewAttachmentSearchTool creates a new attachment search tool.
func NewAttachmentSearchTool() *AttachmentSearchTool {
	return &AttachmentSearchTool{}
}

// WithLimit sets the maximum number of files to search.
func (a *AttachmentSearchTool) WithLimit(n int32) *AttachmentSearchTool {
	a.Limit = &n
	return a
}

func (a *AttachmentSearchTool) toProto() *v1.Tool {
	as := &v1.AttachmentSearch{}
	if a.Limit != nil {
		as.Limit = a.Limit
	}
	return &v1.Tool{
		Tool: &v1.Tool_AttachmentSearch{
			AttachmentSearch: as,
		},
	}
}

// MCPTool enables MCP (Model Context Protocol) tool usage.
type MCPTool struct {
	// ServerLabel identifies the MCP server.
	ServerLabel string
	// ServerURL is the URL of the MCP server.
	ServerURL string
}

// NewMCPTool creates a new MCP tool.
func NewMCPTool(serverLabel, serverURL string) *MCPTool {
	return &MCPTool{
		ServerLabel: serverLabel,
		ServerURL:   serverURL,
	}
}

func (m *MCPTool) toProto() *v1.Tool {
	return &v1.Tool{
		Tool: &v1.Tool_Mcp{
			Mcp: &v1.MCP{
				ServerLabel: m.ServerLabel,
				ServerUrl:   m.ServerURL,
			},
		},
	}
}

// ToolCallType indicates whether a tool call is client-side or server-side.
type ToolCallType int

const (
	// ToolCallTypeClient indicates a client-side tool call (you must execute it).
	ToolCallTypeClient ToolCallType = iota
	// ToolCallTypeServer indicates a server-side tool call (xAI executed it).
	ToolCallTypeServer
)

// ToolCallStatus indicates the status of a tool call.
type ToolCallStatus int

const (
	// ToolCallStatusPending indicates the tool call is pending execution.
	ToolCallStatusPending ToolCallStatus = iota
	// ToolCallStatusCompleted indicates the tool call completed successfully.
	ToolCallStatusCompleted
	// ToolCallStatusFailed indicates the tool call failed.
	ToolCallStatusFailed
)

// ToolCallInfo represents a tool call made by the model.
type ToolCallInfo struct {
	// ID is the unique identifier for this tool call.
	ID string
	// Type indicates if this is a client-side or server-side tool call.
	Type ToolCallType
	// Status is the current status of the tool call.
	Status ToolCallStatus
	// ErrorMessage contains an error message if the call failed.
	ErrorMessage string
	// Function contains the function call details.
	Function *FunctionCall
}

// FunctionCall represents a function call made by the model.
type FunctionCall struct {
	// Name is the function name.
	Name string
	// Arguments is the JSON-encoded arguments.
	Arguments string
}

// IsClientSide returns true if this is a client-side tool call that you must execute.
func (tc *ToolCallInfo) IsClientSide() bool {
	return tc.Type == ToolCallTypeClient
}

// IsServerSide returns true if this is a server-side tool call that xAI executed.
func (tc *ToolCallInfo) IsServerSide() bool {
	return tc.Type == ToolCallTypeServer
}

// toolCallFromProto converts a proto ToolCall to a ToolCallInfo.
func toolCallFromProto(tc *v1.ToolCall) *ToolCallInfo {
	if tc == nil {
		return nil
	}

	info := &ToolCallInfo{
		ID: tc.GetId(),
	}

	// Type - determine if server-side or client-side based on type
	switch tc.GetType() {
	case v1.ToolCallType_TOOL_CALL_TYPE_CLIENT_SIDE_TOOL:
		info.Type = ToolCallTypeClient
	case v1.ToolCallType_TOOL_CALL_TYPE_WEB_SEARCH_TOOL,
		v1.ToolCallType_TOOL_CALL_TYPE_X_SEARCH_TOOL,
		v1.ToolCallType_TOOL_CALL_TYPE_CODE_EXECUTION_TOOL,
		v1.ToolCallType_TOOL_CALL_TYPE_COLLECTIONS_SEARCH_TOOL,
		v1.ToolCallType_TOOL_CALL_TYPE_MCP_TOOL,
		v1.ToolCallType_TOOL_CALL_TYPE_ATTACHMENT_SEARCH_TOOL:
		info.Type = ToolCallTypeServer
	default:
		info.Type = ToolCallTypeClient
	}

	// Status
	switch tc.GetStatus() {
	case v1.ToolCallStatus_TOOL_CALL_STATUS_COMPLETED:
		info.Status = ToolCallStatusCompleted
	case v1.ToolCallStatus_TOOL_CALL_STATUS_FAILED:
		info.Status = ToolCallStatusFailed
	default:
		info.Status = ToolCallStatusPending
	}

	// Error message
	if tc.ErrorMessage != nil {
		info.ErrorMessage = *tc.ErrorMessage
	}

	// Function
	if fn := tc.GetFunction(); fn != nil {
		info.Function = &FunctionCall{
			Name:      fn.GetName(),
			Arguments: fn.GetArguments(),
		}
	}

	return info
}

// IsClientSideTool checks if a tool call matches one of the registered client-side tools.
// This is useful for distinguishing which tool calls you need to execute versus
// which were executed server-side by xAI.
func IsClientSideTool(call *ToolCallInfo, registeredTools []Tool) bool {
	if call == nil || call.Function == nil {
		return false
	}

	for _, tool := range registeredTools {
		if fn, ok := tool.(*FunctionTool); ok {
			if fn.Name == call.Function.Name {
				return true
			}
		}
	}
	return false
}
