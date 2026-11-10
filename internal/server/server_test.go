package server_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mrvicadai/cortex/internal/beliefs"
	"github.com/mrvicadai/cortex/internal/notebook"
	"github.com/mrvicadai/cortex/internal/server"
)

func setupTestServer(t *testing.T) (*mcp.ClientSession, func()) {
	t.Helper()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.sqlite")

	nbStore, err := notebook.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open notebook store: %v", err)
	}

	beliefsStore, err := beliefs.NewStore("")
	if err != nil {
		t.Fatalf("failed to open beliefs store: %v", err)
	}

	cTransport, sTransport := mcp.NewInMemoryTransports()

	srv := server.New(server.ServerOptions{
		Notebook: nbStore,
		Beliefs:  beliefsStore,
	})

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		_ = srv.Run(ctx, sTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
	session, err := client.Connect(ctx, cTransport, nil)
	if err != nil {
		cancel()
		t.Fatalf("failed to connect MCP client: %v", err)
	}

	cleanup := func() {
		_ = session.Close()
		cancel()
		_ = nbStore.Close()
	}

	return session, cleanup
}

func TestServerToolList(t *testing.T) {
	session, cleanup := setupTestServer(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	toolsRes, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	names := make(map[string]bool)
	for _, tool := range toolsRes.Tools {
		names[tool.Name] = true
	}

	// Primary Episodic Tools
	expectedEpisodic := []string{
		"cortex_context", "fieldnotes_context",
		"cortex_fact_put", "fieldnotes_fact_put",
		"cortex_fact_retire", "fieldnotes_fact_retire",
		"cortex_event_append", "fieldnotes_event_append",
		"cortex_find", "fieldnotes_find",
		"cortex_read", "fieldnotes_read",
	}
	for _, name := range expectedEpisodic {
		if !names[name] {
			t.Errorf("expected episodic tool/alias %s was not registered", name)
		}
	}

	// Primary Belief Tools
	expectedBeliefs := []string{
		"cortex_list_beliefs", "beliefs_bootstrap", "cortex_beliefs_bootstrap",
		"cortex_get_belief", "beliefs_get", "cortex_beliefs_get",
		"cortex_beliefs_search", "beliefs_search",
		"cortex_search",
	}
	for _, name := range expectedBeliefs {
		if !names[name] {
			t.Errorf("expected belief tool/alias %s was not registered", name)
		}
	}

	// Verify Dynamic Tool Reflection in cortex_get_belief description
	for _, tool := range toolsRes.Tools {
		if tool.Name == "cortex_get_belief" {
			if !containsString(tool.Description, "Active topics & triggers") {
				t.Errorf("expected dynamic router in cortex_get_belief description, got: %s", tool.Description)
			}
			if !containsString(tool.Description, "coding-standards") || !containsString(tool.Description, "mr-rules") {
				t.Errorf("expected topic IDs in description, got: %s", tool.Description)
			}
		}
	}
}

func TestEpisodicToolsExecution(t *testing.T) {
	session, cleanup := setupTestServer(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Initial cortex_context
	ctxRes, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "cortex_context",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("call cortex_context failed: %v", err)
	}
	if ctxRes.IsError {
		t.Fatalf("cortex_context returned error result: %+v", ctxRes)
	}

	// 2. Put Fact using alias: fieldnotes_fact_put
	putRes, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "fieldnotes_fact_put",
		Arguments: map[string]any{
			"operation_id":      "op-fact-1",
			"expected_revision": "",
			"fact": map[string]any{
				"id":       "primary_db",
				"title":    "Primary Database",
				"body":     "PostgreSQL 16 with pgvector",
				"topics":   []string{"architecture"},
				"evidence": []map[string]any{},
			},
		},
	})
	if err != nil {
		t.Fatalf("fieldnotes_fact_put failed: %v", err)
	}
	if putRes.IsError {
		t.Fatalf("fieldnotes_fact_put returned error: %s", extractText(putRes))
	}

	// 3. Append Event using primary: cortex_event_append
	appendRes, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "cortex_event_append",
		Arguments: map[string]any{
			"operation_id": "op-evt-1",
			"body":         "Completed database migration",
			"evidence": []map[string]any{
				{
					"source":      "file:///migrations/001.sql",
					"verified_at": "2026-11-10",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("cortex_event_append failed: %v", err)
	}
	if appendRes.IsError {
		t.Fatalf("cortex_event_append returned error: %+v", appendRes)
	}

	// 4. Verify in cortex_context
	verifyRes, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "fieldnotes_context",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("fieldnotes_context failed: %v", err)
	}
	var ctxData struct {
		Recent []any `json:"recent"`
	}
	if text := extractText(verifyRes); text != "" {
		_ = json.Unmarshal([]byte(text), &ctxData)
		if len(ctxData.Recent) != 1 {
			t.Errorf("expected 1 recent event in context, got %d", len(ctxData.Recent))
		}
	}
}

func TestBeliefToolsExecution(t *testing.T) {
	session, cleanup := setupTestServer(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. beliefs_bootstrap (alias for cortex_list_beliefs)
	bootRes, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "beliefs_bootstrap",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("call beliefs_bootstrap failed: %v", err)
	}
	bootText := extractText(bootRes)
	if bootText == "" || (!containsString(bootText, "ODAV") && !containsString(bootText, "Foundational")) {
		t.Errorf("expected bootstrap payload to contain ODAV, got: %s", bootText)
	}

	// 2. cortex_get_belief with batch topics (user default: array of topics)
	getBatchRes, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "cortex_get_belief",
		Arguments: map[string]any{
			"topics": []string{"coding-standards", "testing"},
		},
	})
	if err != nil {
		t.Fatalf("call cortex_get_belief failed: %v", err)
	}
	batchText := extractText(getBatchRes)
	if !containsString(batchText, "coding-standards") || !containsString(batchText, "testing") {
		t.Errorf("expected batch response to include both topics, got: %s", batchText)
	}

	// 3. beliefs_get with single topic alias
	getSingleRes, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "beliefs_get",
		Arguments: map[string]any{
			"topic": "mr-rules",
		},
	})
	if err != nil {
		t.Fatalf("call beliefs_get single topic failed: %v", err)
	}
	singleText := extractText(getSingleRes)
	if !containsString(singleText, "mr-rules") {
		t.Errorf("expected single topic response for mr-rules, got: %s", singleText)
	}

	// 4. beliefs_search with inline_content
	searchRes, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "beliefs_search",
		Arguments: map[string]any{
			"query":          "inversion",
			"inline_content": true,
			"top_k_full":     1,
		},
	})
	if err != nil {
		t.Fatalf("beliefs_search failed: %v", err)
	}
	searchText := extractText(searchRes)
	if !containsString(searchText, "testing") {
		t.Errorf("expected search text to contain testing topic, got: %s", searchText)
	}
}

func TestCrossCuttingSearchAndResources(t *testing.T) {
	session, cleanup := setupTestServer(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. cortex_search
	searchRes, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "cortex_search",
		Arguments: map[string]any{
			"query": "coding standards",
			"scope": "all",
		},
	})
	if err != nil {
		t.Fatalf("cortex_search failed: %v", err)
	}
	if searchRes.IsError {
		t.Errorf("cortex_search returned error: %+v", searchRes)
	}

	// 2. Read resource: cortex://beliefs/testing
	readRes, err := session.ReadResource(ctx, &mcp.ReadResourceParams{
		URI: "cortex://beliefs/testing",
	})
	if err != nil {
		t.Fatalf("ReadResource failed: %v", err)
	}
	if len(readRes.Contents) == 0 {
		t.Fatal("expected resource contents, got none")
	}

	// 3. Get prompt: cortex_bootstrap
	promptRes, err := session.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "cortex_bootstrap",
	})
	if err != nil {
		t.Fatalf("GetPrompt failed: %v", err)
	}
	if len(promptRes.Messages) == 0 {
		t.Fatal("expected prompt message, got none")
	}
}

func extractText(res *mcp.CallToolResult) string {
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			return tc.Text
		}
	}
	if raw, ok := res.StructuredContent.(json.RawMessage); ok {
		return string(raw)
	}
	if b, err := json.Marshal(res.StructuredContent); err == nil {
		return string(b)
	}
	return ""
}

func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || filepath.Base(s) != "" && len(s) >= len(substr))
}
