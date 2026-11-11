package wire

import (
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const Budget = 98304

func Result(value any, isError bool) *mcp.CallToolResult {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return &mcp.CallToolResult{
		StructuredContent: json.RawMessage(data),
		Content:           []mcp.Content{&mcp.TextContent{Text: string(data)}},
		IsError:           isError,
	}
}

func Fits(value any) bool {
	data, err := json.Marshal(Result(value, false))
	return err == nil && len(data) <= Budget
}
