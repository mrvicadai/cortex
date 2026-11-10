package beliefs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mrvicadai/cortex/internal/beliefs"
)

func TestEmbeddedBeliefsLoad(t *testing.T) {
	store, err := beliefs.NewStore("")
	if err != nil {
		t.Fatalf("failed to create beliefs store: %v", err)
	}

	topics := store.ListTopics("")
	if len(topics) != 10 {
		t.Fatalf("expected 10 default topics, got %d", len(topics))
	}

	for _, topic := range topics {
		if topic.Origin != "packaged" {
			t.Errorf("expected origin 'packaged', got %s for topic %s", topic.Origin, topic.ID)
		}
		if topic.Version != "0.1.0" {
			t.Errorf("expected version '0.1.0', got %s for topic %s", topic.Version, topic.ID)
		}
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

	// 3. Unknown topic returns error
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

	store, err := beliefs.NewStore(tempDir)
	if err != nil {
		t.Fatalf("failed to create beliefs store with overrides: %v", err)
	}

	topics := store.ListTopics("")
	if len(topics) != 11 { // 10 original + 1 new custom
		t.Fatalf("expected 11 topics with custom override, got %d", len(topics))
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

	rendered := store.GetRenderedTableBlock()
	if len(rendered) < 300 {
		t.Errorf("rendered table block unexpectedly small: %d", len(rendered))
	}
}
