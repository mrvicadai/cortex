package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mrvicadai/cortex/internal/beliefs"
	"github.com/mrvicadai/cortex/internal/notebook"
	"github.com/mrvicadai/cortex/internal/wire"
)

const serverInstructions = "Cortex provides unified agent context: normative operating beliefs and persistent episodic memory for AI coding agents. Check directory from cortex_context to confirm the active project notebook. Use cortex_get_belief (or beliefs_get) to load authoritative guidance before coding, testing, or opening merge requests."

type ServerOptions struct {
	Notebook *notebook.Store
	Beliefs  *beliefs.Store
}

func New(opts ServerOptions) *mcp.Server {
	s := mcp.NewServer(
		&mcp.Implementation{Name: "cortex", Version: "0.1.0"},
		&mcp.ServerOptions{Instructions: serverInstructions},
	)
	s.AddReceivingMiddleware(validateArguments)

	registerEpisodicTools(s, opts.Notebook)
	registerBeliefTools(s, opts.Beliefs)
	registerCrossCuttingTools(s, opts.Notebook, opts.Beliefs)
	registerResourcesAndPrompts(s, opts.Notebook, opts.Beliefs)

	return s
}

// ----------------------------------------------------------------------------
// Episodic Tools (Notebook Engine)
// ----------------------------------------------------------------------------

func registerEpisodicTools(s *mcp.Server, nb *notebook.Store) {
	if nb == nil {
		return
	}

	// 1. Context (cortex_context + fieldnotes_context)
	contextDesc := "Load essential guidance and complete recent events for the active project directory within an 8 KiB budget. Resume older events with cursor only."
	registerToolPair(s, "cortex_context", "fieldnotes_context", contextDesc, false, nb.Context)

	// 2. Fact Put (cortex_fact_put + fieldnotes_fact_put)
	factPutDesc := "Create or CAS-replace a durable project invariant (e.g. build configs, architecture, @context). Do NOT use for completed tasks or historical actions (e.g. 'PR merged') — use cortex_event_append instead."
	registerToolPair(s, "cortex_fact_put", "fieldnotes_fact_put", factPutDesc, true, nb.PutFact)

	// 3. Fact Retire (cortex_fact_retire + fieldnotes_fact_retire)
	factRetireDesc := "CAS-retire a fact, preserving an opaque tombstone revision. Cannot retire @context. Retry the identical request with the same operation_id."
	registerToolPair(s, "cortex_fact_retire", "fieldnotes_fact_retire", factRetireDesc, true, nb.RetireFact)

	// 4. Event Append (cortex_event_append + fieldnotes_event_append)
	eventAppendDesc := "Append an immutable completed progress milestone or chronological outcome with evidence (e.g. 'PR #42 merged', 'Refactored auth handler', 'Tests passed')."
	registerToolPair(s, "cortex_event_append", "fieldnotes_event_append", eventAppendDesc, false, nb.AppendEvent)

	// 5. Find (cortex_find + fieldnotes_find)
	findDesc := "Browse or search references across facts and events (newest first). Query uses ANDed case-insensitive literal substrings. Resume with cursor."
	registerToolPair(s, "cortex_find", "fieldnotes_find", findDesc, false, nb.Find)

	// 6. Read (cortex_read + fieldnotes_read)
	readDesc := "Read one to ten ordered references. Only complete records, missing statuses, or versioned tombstones are returned; bodies are never chunked."
	registerToolPair(s, "cortex_read", "fieldnotes_read", readDesc, false, nb.Read)
}

func registerToolPair[In, Out any](s *mcp.Server, primary, alias, desc string, destructive bool, fn func(context.Context, In) (Out, error)) {
	addTool(s, primary, desc, destructive, fn)
	if alias != "" {
		addTool(s, alias, desc+" (alias for "+primary+")", destructive, fn)
	}
}

func addTool[In, Out any](s *mcp.Server, name, description string, destructive bool, run func(context.Context, In) (Out, error)) {
	openWorld := false
	mcp.AddTool(s, &mcp.Tool{
		Name:        name,
		Description: description,
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    false,
			DestructiveHint: &destructive,
			IdempotentHint:  true,
			OpenWorldHint:   &openWorld,
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, any, error) {
		out, err := run(ctx, in)
		if err != nil {
			return errorResult(err), nil, nil
		}
		if !wire.Fits(out) {
			return errorResult(&notebook.Fault{Code: "result_too_large", Message: "result exceeds the complete-response budget"}), nil, nil
		}
		return wire.Result(out, false), nil, nil
	})
}

// ----------------------------------------------------------------------------
// Normative Belief Tools
// ----------------------------------------------------------------------------

type BootstrapInput struct {
	Query string `json:"query,omitempty" jsonschema:"Optional keyword filter to filter available belief topics (e.g., 'git', 'testing')."`
}

type GetBeliefsInput struct {
	Topics  []string `json:"topics,omitempty" jsonschema:"Array of topic IDs to retrieve in a single batch (default parameter). Allowed: coding-standards, testing, mr-rules, git-stacking, project-design, walkthrough, datadog-query-craft, premise-check, context-hygiene, governance."`
	TopicID string   `json:"topic_id,omitempty" jsonschema:"Optional single topic ID."`
	Topic   string   `json:"topic,omitempty" jsonschema:"Optional single topic ID."`
}

type BeliefsSearchInput struct {
	Query         string `json:"query" jsonschema:"Search keywords or question across operating rules."`
	InlineContent bool   `json:"inline_content,omitempty" jsonschema:"When true, inlines full markdown guidance for top matching topic(s) to save turns."`
	TopKFull      int    `json:"top_k_full,omitempty" jsonschema:"Number of top matching topics to inline (default: 1, max: 5)."`
}

func formatGetBeliefDescription(b *beliefs.Store) string {
	var sb strings.Builder
	sb.WriteString("Retrieve authoritative guidance for one or more topics before executing that task (e.g. before coding, testing, opening MRs, or designing projects). Defaults to accepting an array of topic IDs. Active topics & triggers:\n")
	for _, t := range b.ListTopics("") {
		sb.WriteString(fmt.Sprintf("• %s: %s\n", t.ID, t.Trigger))
	}
	return sb.String()
}

func registerBeliefTools(s *mcp.Server, b *beliefs.Store) {
	if b == nil {
		return
	}

	// 7. Bootstrap (cortex_bootstrap + beliefs_bootstrap)
	bootstrapDesc := "Session bootstrap: returns core operating principles (ODAV, Least-to-Most, Project-State) and the active task-triggered topics router."
	bootstrapHandler := func(ctx context.Context, in BootstrapInput) (any, error) {
		if strings.TrimSpace(in.Query) == "" {
			return struct {
				Bootstrap string                  `json:"bootstrap"`
				Topics    []beliefs.TopicMetadata `json:"topics"`
			}{
				Bootstrap: b.GetBootstrapPayload(),
				Topics:    b.ListTopics(""),
			}, nil
		}
		return struct {
			Topics []beliefs.TopicMetadata `json:"topics"`
		}{
			Topics: b.ListTopics(in.Query),
		}, nil
	}
	registerToolPair(s, "cortex_bootstrap", "beliefs_bootstrap", bootstrapDesc, false, bootstrapHandler)

	// 8. Get Beliefs (defaults to array of topics)
	getDesc := formatGetBeliefDescription(b)
	getHandler := func(ctx context.Context, in GetBeliefsInput) (any, error) {
		var topicIDs []string
		if len(in.Topics) > 0 {
			topicIDs = append(topicIDs, in.Topics...)
		}
		if in.TopicID != "" && !contains(topicIDs, in.TopicID) {
			topicIDs = append(topicIDs, in.TopicID)
		}
		if in.Topic != "" && !contains(topicIDs, in.Topic) {
			topicIDs = append(topicIDs, in.Topic)
		}

		topics, err := b.GetTopics(topicIDs)
		if err != nil {
			return nil, &notebook.Fault{Code: "topic_not_found", Message: err.Error()}
		}

		if len(topics) == 1 {
			return struct {
				TopicID string `json:"topic_id"`
				Title   string `json:"title"`
				Version string `json:"version"`
				Origin  string `json:"origin"`
				Content string `json:"content"`
			}{
				TopicID: topics[0].ID,
				Title:   topics[0].Title,
				Version: topics[0].Version,
				Origin:  topics[0].Origin,
				Content: topics[0].Content,
			}, nil
		}

		return struct {
			Count  int                     `json:"count"`
			Topics []beliefs.TopicMetadata `json:"topics"`
		}{
			Count:  len(topics),
			Topics: topics,
		}, nil
	}
	registerToolPair(s, "cortex_get_belief", "beliefs_get", getDesc, false, getHandler)

	// 9. Beliefs Search
	searchDesc := "Search normative belief rules using natural language questions or keywords, optionally inlining full markdown guidance."
	searchHandler := func(ctx context.Context, in BeliefsSearchInput) (any, error) {
		res := b.SearchTopics(in.Query, in.InlineContent, in.TopKFull)
		return res, nil
	}
	registerToolPair(s, "cortex_beliefs_search", "beliefs_search", searchDesc, false, searchHandler)
}

// ----------------------------------------------------------------------------
// Cross-Cutting Tools
// ----------------------------------------------------------------------------

type CortexSearchInput struct {
	Query string `json:"query" jsonschema:"Search query spanning both normative rules and episodic project history."`
	Scope string `json:"scope,omitempty" jsonschema:"'all' (default), 'beliefs', or 'fieldnotes'."`
}

func registerCrossCuttingTools(s *mcp.Server, nb *notebook.Store, b *beliefs.Store) {
	searchDesc := "Unified full-text search spanning normative operating beliefs and empirical project history."
	mcp.AddTool(s, &mcp.Tool{
		Name:        "cortex_search",
		Description: searchDesc,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in CortexSearchInput) (*mcp.CallToolResult, any, error) {
		scope := strings.ToLower(strings.TrimSpace(in.Scope))
		if scope == "" {
			scope = "all"
		}

		result := map[string]any{}

		if (scope == "all" || scope == "beliefs") && b != nil {
			beliefMatches := b.SearchTopics(in.Query, false, 0)
			result["beliefs"] = beliefMatches.Matches
		}

		if (scope == "all" || scope == "fieldnotes") && nb != nil {
			findRes, err := nb.Find(ctx, notebook.FindInput{Query: in.Query})
			if err == nil {
				result["fieldnotes"] = findRes
			}
		}

		return wire.Result(result, false), nil, nil
	})
}

// ----------------------------------------------------------------------------
// Resources & Prompts
// ----------------------------------------------------------------------------

func registerResourcesAndPrompts(s *mcp.Server, nb *notebook.Store, b *beliefs.Store) {
	// Resource: cortex://beliefs/{topic}
	if b != nil {
		s.AddResourceTemplate(&mcp.ResourceTemplate{
			URITemplate: "cortex://beliefs/{topic}",
			Name:        "Belief Topic Documentation",
			MIMEType:    "text/markdown",
		}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			uri := req.Params.URI
			topicID := strings.TrimPrefix(uri, "cortex://beliefs/")
			topics, err := b.GetTopics([]string{topicID})
			if err != nil || len(topics) == 0 {
				return nil, fmt.Errorf("belief topic not found: %s", topicID)
			}
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{
					{
						URI:      uri,
						MIMEType: "text/markdown",
						Text:     topics[0].Content,
					},
				},
			}, nil
		})
	}

	// Resource: cortex://context
	if nb != nil {
		s.AddResource(&mcp.Resource{
			URI:      "cortex://context",
			Name:     "Project Context Snapshot",
			MIMEType: "application/json",
		}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			ctxRes, err := nb.Context(ctx, notebook.ContextInput{})
			if err != nil {
				return nil, err
			}
			data, err := json.MarshalIndent(ctxRes, "", "  ")
			if err != nil {
				return nil, err
			}
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{
					{
						URI:      req.Params.URI,
						MIMEType: "application/json",
						Text:     string(data),
					},
				},
			}, nil
		})
	}

	// Prompt: cortex_bootstrap
	s.AddPrompt(&mcp.Prompt{
		Name:        "cortex_bootstrap",
		Description: "Initial bootstrap prompt seeding agent with project context and operating belief router",
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		var text string
		if b != nil {
			text = b.GetBootstrapPayload()
		} else {
			text = "Cortex context initialized."
		}
		return &mcp.GetPromptResult{
			Messages: []*mcp.PromptMessage{
				{
					Role: "user",
					Content: &mcp.TextContent{
						Text: text,
					},
				},
			},
		}, nil
	})
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func errorResult(err error) *mcp.CallToolResult {
	fault := &notebook.Fault{Code: "operation_failed", Message: err.Error()}
	var known *notebook.Fault
	if errors.As(err, &known) {
		fault = known
	}
	return wire.Result(struct {
		Error *notebook.Fault `json:"error"`
	}{fault}, true)
}
