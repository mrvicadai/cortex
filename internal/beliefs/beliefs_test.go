package beliefs_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mrvicadai/cortex/internal/beliefs"
)

func TestEmbeddedBeliefsLoad(t *testing.T) {
	store, err := beliefs.NewStore("")
	if err != nil {
		t.Fatalf("failed to create beliefs store: %v", err)
	}

	topics := store.ListTopics("")
	if len(topics) != 13 {
		t.Fatalf("expected 13 default topics, got %d", len(topics))
	}

	for _, topic := range topics {
		if topic.Origin != "packaged" {
			t.Errorf("expected origin 'packaged', got %s for topic %s", topic.Origin, topic.ID)
		}
		if topic.Version != "0.1.0" {
			t.Errorf("expected version '0.1.0', got %s for topic %s", topic.Version, topic.ID)
		}
	}

	// Verify custom-beliefs topic exists and has valid instructions
	cb, err := store.GetTopics([]string{"custom-beliefs"})
	if err != nil {
		t.Fatalf("getting custom-beliefs topic: %v", err)
	}
	if cb[0].Title != "Custom Belief Authoring" {
		t.Errorf("expected title 'Custom Belief Authoring', got %s", cb[0].Title)
	}
	if !strings.Contains(cb[0].Content, "YAML frontmatter") {
		t.Errorf("expected custom-beliefs content to mention YAML frontmatter")
	}

	// Verify feasibility-audit topic exists and encodes Stage 1 Dependency Matrix and Gate 1 hard stop
	fa, err := store.GetTopics([]string{"feasibility-audit"})
	if err != nil {
		t.Fatalf("getting feasibility-audit topic: %v", err)
	}
	if fa[0].Title != "Feasibility & Dependency Audit (Stage 1)" {
		t.Errorf("expected title 'Feasibility & Dependency Audit (Stage 1)', got %s", fa[0].Title)
	}
	if !strings.Contains(fa[0].Content, "Dependency & Capability Matrix") || !strings.Contains(fa[0].Content, "Gate 1 hard stop") || !strings.Contains(fa[0].Content, "cortex_find") {
		t.Errorf("expected feasibility-audit content to include matrix, Gate 1 hard stop, and Cortex lookup")
	}

	// Verify project-planning topic exists and encodes Stage 3 gates and sizing rules
	pp, err := store.GetTopics([]string{"project-planning"})
	if err != nil {
		t.Fatalf("getting project-planning topic: %v", err)
	}
	if pp[0].Title != "Project Planning & Task Breakdown (Stage 3)" {
		t.Errorf("expected title 'Project Planning & Task Breakdown (Stage 3)', got %s", pp[0].Title)
	}
	if !strings.Contains(pp[0].Content, "One-Sentence Diff Test") || !strings.Contains(pp[0].Content, "Gate 2 hard stop") || !strings.Contains(pp[0].Content, "Gate 3") {
		t.Errorf("expected project-planning content to include sizing gates and Gate 2/3 rules")
	}

	// Verify rfc-authoring topic exists and encodes Stage 2 gates, ADR policy, Mermaid, and Concreteness rules
	rfc, err := store.GetTopics([]string{"rfc-authoring"})
	if err != nil {
		t.Fatalf("getting rfc-authoring topic: %v", err)
	}
	if rfc[0].Title != "RFC Authoring & Agent-Ready Blueprints (Stage 2)" {
		t.Errorf("expected title 'RFC Authoring & Agent-Ready Blueprints (Stage 2)', got %s", rfc[0].Title)
	}
	if !strings.Contains(rfc[0].Content, "Concreteness & readability rules") || !strings.Contains(rfc[0].Content, "RFC vs. ADR policy") || !strings.Contains(rfc[0].Content, "stateDiagram-v2") || !strings.Contains(rfc[0].Content, "Failure Modes & Concurrency") || !strings.Contains(rfc[0].Content, "Must Be Empty Before Stage 3") {
		t.Errorf("expected rfc-authoring content to include ADR policy, readability rules, Mermaid guidance, Failure Modes, and Gate 2")
	}

	gov, err := store.GetTopics([]string{"governance"})
	if err != nil {
		t.Fatalf("getting governance topic: %v", err)
	}
	if !strings.Contains(gov[0].Content, "120 physical lines") {
		t.Errorf("expected governance content to enforce 120 physical lines cap")
	}

	// Verify removed project-design topic is no longer registered
	if _, err := store.GetTopics([]string{"project-design"}); err == nil {
		t.Errorf("expected project-design topic to be removed")
	}
}

func TestGetTopicsBatch(t *testing.T) {
	store, err := beliefs.NewStore("")
	if err != nil {
		t.Fatalf("failed to create beliefs store: %v", err)
	}

	// 1. Single topic
	single, err := store.GetTopics([]string{"coding-standards"})
	if err != nil {
		t.Fatalf("failed to get single topic: %v", err)
	}
	if len(single) != 1 || single[0].ID != "coding-standards" {
		t.Fatalf("expected coding-standards topic, got %+v", single)
	}
	if len(single[0].Content) < 100 {
		t.Errorf("topic content unexpectedly small: %d bytes", len(single[0].Content))
	}

	// 2. Batch topics
	batch, err := store.GetTopics([]string{"coding-standards", "testing", "mr-rules"})
	if err != nil {
		t.Fatalf("failed to get batch topics: %v", err)
	}
	if len(batch) != 3 {
		t.Fatalf("expected 3 topics, got %d", len(batch))
	}
	if batch[0].ID != "coding-standards" || batch[1].ID != "testing" || batch[2].ID != "mr-rules" {
		t.Errorf("unexpected batch order: %+v", batch)
	}

	// 3. Foundational principles accessible via GetTopics
	foundationalGot, err := store.GetTopics([]string{"project-state", "odav", "least-to-most"})
	if err != nil {
		t.Fatalf("failed to get foundational topics: %v", err)
	}
	if len(foundationalGot) != 3 {
		t.Fatalf("expected 3 foundational topics, got %d", len(foundationalGot))
	}
	if foundationalGot[0].Origin != "foundational" || foundationalGot[0].ID != "project-state" {
		t.Errorf("unexpected foundational topic metadata: %+v", foundationalGot[0])
	}

	// 4. Unknown topic returns error
	_, err = store.GetTopics([]string{"non-existent-topic"})
	if err == nil {
		t.Error("expected error for unknown topic, got nil")
	}
}

func TestSearchTopicsWithInline(t *testing.T) {
	store, err := beliefs.NewStore("")
	if err != nil {
		t.Fatalf("failed to create beliefs store: %v", err)
	}

	// Search without inline
	res := store.SearchTopics("merge conflicts", false, 0)
	if len(res.Matches) == 0 {
		t.Fatal("expected search matches for 'merge conflicts'")
	}
	if len(res.InlineContent) != 0 {
		t.Errorf("expected no inline content, got %d", len(res.InlineContent))
	}

	// Search with inline
	inlineRes := store.SearchTopics("inversion", true, 1)
	if len(inlineRes.Matches) == 0 {
		t.Fatal("expected search matches for 'inversion'")
	}
	if len(inlineRes.InlineContent) != 1 {
		t.Fatalf("expected 1 inline content topic, got %d", len(inlineRes.InlineContent))
	}
	if inlineRes.InlineContent[0].ID != "testing" {
		t.Errorf("expected top inline topic to be 'testing', got %s", inlineRes.InlineContent[0].ID)
	}
	if len(inlineRes.InlineContent[0].Content) < 50 {
		t.Errorf("inline content unexpectedly empty")
	}
	// Search foundational principles
	foundationalSearch := store.SearchTopics("project-state", false, 0)
	if len(foundationalSearch.Matches) == 0 {
		t.Fatalf("expected matches for foundational topic 'project-state', got 0")
	}
	if foundationalSearch.Matches[0].TopicID != "project-state" {
		t.Errorf("expected top match to be 'project-state', got %s", foundationalSearch.Matches[0].TopicID)
	}
}

func TestUserOverrides(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Override existing topic: coding-standards.md
	overrideFile := filepath.Join(tempDir, "coding-standards.md")
	overrideContent := "# Custom Team Coding Standards\n\nAlways write tests first and prefer functional composition."
	if err := os.WriteFile(overrideFile, []byte(overrideContent), 0600); err != nil {
		t.Fatalf("writing override file: %v", err)
	}

	// 2. Add brand new custom topic: custom-cloud.md
	customFile := filepath.Join(tempDir, "custom-cloud.md")
	customContent := "# Cloud Architecture\n\nDeploy all microservices with Terraform."
	if err := os.WriteFile(customFile, []byte(customContent), 0600); err != nil {
		t.Fatalf("writing custom file: %v", err)
	}

	// 3. Add nested directory topic: security/secrets.md
	secDir := filepath.Join(tempDir, "security")
	if err := os.MkdirAll(secDir, 0700); err != nil {
		t.Fatalf("creating security dir: %v", err)
	}
	nestedSecFile := filepath.Join(secDir, "secrets.md")
	nestedSecContent := "# Corporate Secrets\n\nNever commit tokens to git."
	if err := os.WriteFile(nestedSecFile, []byte(nestedSecContent), 0600); err != nil {
		t.Fatalf("writing nested file: %v", err)
	}

	// 4. Add nested directory topic with frontmatter: infra/k8s.md
	infraDir := filepath.Join(tempDir, "infra")
	if err := os.MkdirAll(infraDir, 0700); err != nil {
		t.Fatalf("creating infra dir: %v", err)
	}
	nestedInfraFile := filepath.Join(infraDir, "k8s.md")
	nestedInfraContent := "---\nid: k8s-infra\ntitle: Kubernetes Infrastructure\ntrigger: Before editing helm charts or deployment yamls\n---\n# K8s Standards\n\nAlways specify resource limits."
	if err := os.WriteFile(nestedInfraFile, []byte(nestedInfraContent), 0600); err != nil {
		t.Fatalf("writing nested infra file: %v", err)
	}

	// 5. Add custom topic with quoted frontmatter values
	quotedFile := filepath.Join(tempDir, "payments.md")
	quotedContent := "---\nid: \"payments-compliance\"\ntitle: 'Payment Card Security'\ntrigger: \"When handling credit card numbers or processing payments\"\n---\n# PCI Compliance\nNever log PAN."
	if err := os.WriteFile(quotedFile, []byte(quotedContent), 0600); err != nil {
		t.Fatalf("writing quoted file: %v", err)
	}

	store, err := beliefs.NewStore(tempDir)
	if err != nil {
		t.Fatalf("failed to create beliefs store with overrides: %v", err)
	}

	topics := store.ListTopics("")
	if len(topics) != 17 { // 13 original + 1 custom-cloud + 1 payments-compliance + 2 synthesized routers (security, infra)
		t.Fatalf("expected 17 topics in root router, got %d", len(topics))
	}

	// Verify root router contains domain routers, NOT raw nested sub-topics
	topicMap := make(map[string]bool)
	for _, tMeta := range topics {
		topicMap[tMeta.ID] = true
	}
	if topicMap["security/secrets"] || topicMap["k8s-infra"] {
		t.Errorf("nested sub-topics should not appear in root router table")
	}
	if !topicMap["security"] || !topicMap["infra"] {
		t.Errorf("domain routers 'security' and 'infra' should appear in root router table")
	}

	// Verify synthesized domain router content
	routerGot, err := store.GetTopics([]string{"security"})
	if err != nil {
		t.Fatalf("getting synthesized security domain router: %v", err)
	}
	if !strings.Contains(routerGot[0].Content, "Sub-Topics Router") || !strings.Contains(routerGot[0].Content, "security/secrets") {
		t.Errorf("expected synthesized router table to list sub-topics, got: %s", routerGot[0].Content)
	}

	// Verify overridden topic
	got, err := store.GetTopics([]string{"coding-standards"})
	if err != nil {
		t.Fatalf("getting overridden topic: %v", err)
	}
	if got[0].Origin != "user_override" {
		t.Errorf("expected origin 'user_override', got %s", got[0].Origin)
	}
	if got[0].Content != overrideContent {
		t.Errorf("expected overridden content, got %s", got[0].Content)
	}

	// Verify new custom topic
	customGot, err := store.GetTopics([]string{"custom-cloud"})
	if err != nil {
		t.Fatalf("getting new custom topic: %v", err)
	}
	if customGot[0].Origin != "user_override" {
		t.Errorf("expected origin 'user_override', got %s", customGot[0].Origin)
	}
	if customGot[0].Content != customContent {
		t.Errorf("expected custom content, got %s", customGot[0].Content)
	}

	// Verify nested topic without frontmatter
	secGot, err := store.GetTopics([]string{"security/secrets"})
	if err != nil {
		t.Fatalf("getting nested security topic: %v", err)
	}
	if secGot[0].Title != "Corporate Secrets" {
		t.Errorf("expected title 'Corporate Secrets', got %s", secGot[0].Title)
	}
	if secGot[0].Content != nestedSecContent {
		t.Errorf("expected nested content, got %s", secGot[0].Content)
	}

	// Verify quoted frontmatter parsing
	quotedGot, err := store.GetTopics([]string{"payments-compliance"})
	if err != nil {
		t.Fatalf("getting quoted payments topic: %v", err)
	}
	if quotedGot[0].Title != "Payment Card Security" {
		t.Errorf("expected title without quotes 'Payment Card Security', got %q", quotedGot[0].Title)
	}
	if quotedGot[0].Trigger != "When handling credit card numbers or processing payments" {
		t.Errorf("expected trigger without quotes, got %q", quotedGot[0].Trigger)
	}
	if quotedGot[0].Content != "# PCI Compliance\nNever log PAN." {
		t.Errorf("expected parsed content, got %q", quotedGot[0].Content)
	}

	// 6. Verify that an explicit author-provided domain router overrides auto-synthesizer
	authorRouterFile := filepath.Join(tempDir, "security.md")
	authorContent := "# Custom Security Index\n\nHand-crafted author security router."
	if err := os.WriteFile(authorRouterFile, []byte(authorContent), 0600); err != nil {
		t.Fatalf("writing author router file: %v", err)
	}

	storeWithExplicit, err := beliefs.NewStore(tempDir)
	if err != nil {
		t.Fatalf("failed to reload store: %v", err)
	}
	explicitGot, err := storeWithExplicit.GetTopics([]string{"security"})
	if err != nil {
		t.Fatalf("getting explicit security topic: %v", err)
	}
	if explicitGot[0].Content != authorContent {
		t.Errorf("expected author-provided domain router content, got: %s", explicitGot[0].Content)
	}
}

func TestBootstrapAndSnippet(t *testing.T) {
	store, err := beliefs.NewStore("")
	if err != nil {
		t.Fatalf("failed to create beliefs store: %v", err)
	}

	bootstrap := store.GetBootstrapPayload()
	if len(bootstrap) < 500 {
		t.Errorf("bootstrap payload unexpectedly small: %d", len(bootstrap))
	}

	snippet := store.GetInlineSnippet()
	if len(snippet) < 100 {
		t.Errorf("snippet unexpectedly small: %d", len(snippet))
	}
	if !strings.Contains(snippet, "DEFAULT TO COMMIT") {
		t.Errorf("snippet missing 'DEFAULT TO COMMIT' rule")
	}
	if !strings.Contains(snippet, "RECORD PROGRESS MILESTONES") {
		t.Errorf("snippet missing 'RECORD PROGRESS MILESTONES' rule")
	}
	if !strings.Contains(snippet, "cortex_event_append") {
		t.Errorf("snippet missing 'cortex_event_append'")
	}
	if !strings.Contains(snippet, "cortex_bootstrap({})") || !strings.Contains(snippet, "zero arguments") {
		t.Errorf("snippet missing 'cortex_bootstrap({})' or 'zero arguments' directive")
	}
	if !strings.Contains(snippet, "cortex_context({})") {
		t.Errorf("snippet missing 'cortex_context({})' directive")
	}
	if strings.Contains(snippet, "beliefs_bootstrap") || strings.Contains(snippet, "fieldnotes_") {
		t.Errorf("snippet unexpectedly contains legacy aliases")
	}

	if !strings.Contains(bootstrap, "MANDATORY COMMIT CHECKPOINT") {
		t.Errorf("bootstrap missing 'MANDATORY COMMIT CHECKPOINT' in ODAV")
	}
	if !strings.Contains(bootstrap, "MANDATORY SLICE COMMIT") {
		t.Errorf("bootstrap missing 'MANDATORY SLICE COMMIT' in least-to-most")
	}
	if !strings.Contains(bootstrap, "MANDATORY MILESTONE LOGGING") {
		t.Errorf("bootstrap missing 'MANDATORY MILESTONE LOGGING' in project-state")
	}
	if strings.Contains(bootstrap, "fieldnotes_") {
		t.Errorf("bootstrap unexpectedly contains legacy 'fieldnotes_' references")
	}

	rendered := store.GetRenderedTableBlock()
	if len(rendered) < 300 {
		t.Errorf("rendered table block unexpectedly small: %d", len(rendered))
	}
}
