package server

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mrvicadai/cortex/internal/notebook"
)

func validateArguments(next mcp.MethodHandler) mcp.MethodHandler {
	return func(ctx context.Context, method string, request mcp.Request) (mcp.Result, error) {
		params, ok := request.GetParams().(*mcp.CallToolParamsRaw)
		if method == "tools/call" && ok {
			if invalidArguments(params.Arguments) {
				return errorResult(&notebook.Fault{Code: "invalid_input", Message: "arguments must be valid UTF-8 and at most 32768 bytes"}), nil
			}
		}
		result, err := next(ctx, method, request)
		if method == "tools/call" && err != nil && len(err.Error()) > 512 {
			return nil, fmt.Errorf("invalid tool request: diagnostic exceeds limit")
		}
		return result, err
	}
}

func invalidArguments(arguments []byte) bool {
	return len(arguments) > 32768 || !utf8.Valid(arguments)
}
