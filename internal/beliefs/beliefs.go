package beliefs

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

//go:embed embedded/*
var embeddedFS embed.FS

const PackageVersion = "0.1.0"

type FoundationalPrinciple struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type TopicMetadata struct {
	ID      string `json:"topic_id"`
	Title   string `json:"title"`
	Trigger string `json:"trigger"`
	Version string `json:"version"`
	Origin  string `json:"origin"` // "packaged" | "user_override"
	Content string `json:"content,omitempty"`
}

var defaultTopicDefinitions = []struct {
	id      string
	title   string
	trigger string
	file    string
}{
	{
		id:      "coding-standards",
		title:   "Coding Standards",
		trigger: "Before writing, refactoring, or fixing code",
		file:    "embedded/topics/coding-standards.md",
	},
	{
		id:      "testing",
		title:   "Testing & Verification",
		trigger: "Before defining proof, running verification, or claiming completion",
		file:    "embedded/topics/testing.md",
	},
	{
		id:      "mr-rules",
		title:   "Merge Request / Pull Request Rules",
		trigger: "Before opening or updating an MR or PR",
		file:    "embedded/topics/mr-rules.md",
	},
	{
		id:      "git-stacking",
		title:   "Git Stacking & Topology",
		trigger: "Creating worktrees, branch topology, or stacking",
		file:    "embedded/topics/git-stacking.md",
	},
	{
		id:      "project-design",
		title:   "Project Design & Technical Specifications",
		trigger: "Designing a project, writing an RFC, or technical specification",
		file:    "embedded/topics/project-design.md",
	},
	{
		id:      "walkthrough",
		title:   "Project Walkthroughs",
		trigger: "Authoring a whole-project walkthrough or onboarding explainer",
		file:    "embedded/topics/walkthrough.md",
	},
	{
		id:      "datadog-query-craft",
		title:   "Datadog Query Craft & Telemetry",
		trigger: "Querying Datadog telemetry, APM traces, error breakdowns, or logs",
		file:    "embedded/topics/datadog-query-craft.md",
	},
	{
		id:      "premise-check",
		title:   "Premise Verification",
		trigger: "Checking a claimed defect or dependency premise before committing work",
		file:    "embedded/topics/premise-check.md",
	},
	{
		id:      "context-hygiene",
		title:   "Context Hygiene & Token Conservation",
		trigger: "Reducing token overhead in MCP queries, tool output, or shell commands",
		file:    "embedded/topics/context-hygiene.md",
	},
	{
		id:      "governance",
		title:   "Knowledge-Base Governance",
		trigger: "Maintaining belief rules or performing self-improvement",
		file:    "embedded/topics/governance.md",
	},
	{
		id:      "custom-beliefs",
		title:   "Custom Belief Authoring",
		trigger: "Before adding or authoring custom markdown beliefs (--additional-beliefs)",
		file:    "embedded/topics/custom-beliefs.md",
	},
}

var defaultFoundationalDefinitions = []struct {
	id    string
	title string
	file  string
}{
	{
		id:    "odav",
		title: "Observe, Decide, Act, Verify (ODAV)",
		file:  "embedded/foundational/odav.md",
	},
	{
		id:    "least-to-most",
		title: "Least-to-Most Problem Solving",
		file:  "embedded/foundational/least-to-most.md",
	},
	{
		id:    "project-state",
		title: "Project Memory & State Policy",
		file:  "embedded/foundational/project-state.md",
	},
}

type Store struct {
	mu           sync.RWMutex
	overrideDir  string
	foundational []FoundationalPrinciple
	topics       map[string]TopicMetadata
	topicOrder   []string
}

func NewStore(overrideDir string) (*Store, error) {
	if overrideDir == "" {
		overrideDir = os.Getenv("CORTEX_ADDITIONAL_BELIEFS")
	}

	s := &Store{
		overrideDir: overrideDir,
		topics:      make(map[string]TopicMetadata),
	}

	// 1. Load foundational principles
	for _, f := range defaultFoundationalDefinitions {
		data, err := embeddedFS.ReadFile(f.file)
		if err != nil {
			return nil, fmt.Errorf("reading embedded foundational %s: %w", f.file, err)
		}
		s.foundational = append(s.foundational, FoundationalPrinciple{
			ID:      f.id,
			Title:   f.title,
			Content: strings.TrimSpace(string(data)),
		})
	}

	// 2. Load embedded topics
	for _, t := range defaultTopicDefinitions {
		data, err := embeddedFS.ReadFile(t.file)
		if err != nil {
			return nil, fmt.Errorf("reading embedded topic %s: %w", t.file, err)
		}
		s.topics[t.id] = TopicMetadata{
			ID:      t.id,
			Title:   t.title,
			Trigger: t.trigger,
			Version: PackageVersion,
			Origin:  "packaged",
			Content: strings.TrimSpace(string(data)),
		}
		s.topicOrder = append(s.topicOrder, t.id)
	}

	// 3. Scan user overrides if directory configured
	if overrideDir != "" {
		if err := s.loadOverrides(overrideDir); err != nil {
			// Log or return error
			return nil, fmt.Errorf("loading overrides from %s: %w", overrideDir, err)
		}
	}

	return s, nil
}

func (s *Store) loadOverrides(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("override path %s is not a directory", dir)
	}

	nestedDomains := make(map[string][]string)

	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		return s.loadOverrideFile(dir, path, nestedDomains)
	})
	if err != nil {
		return err
	}

	s.synthesizeDomainRouters(nestedDomains)
	return nil
}

func (s *Store) loadOverrideFile(baseDir, fullPath string, nestedDomains map[string][]string) error {
	relPath, err := filepath.Rel(baseDir, fullPath)
	if err != nil {
		return err
	}
	slashPath := filepath.ToSlash(relPath)
	defaultID := strings.TrimSuffix(slashPath, ".md")

	contentBytes, err := os.ReadFile(fullPath)
	if err != nil {
		return err
	}

	meta, content := parseOverrideMarkdown(defaultID, string(contentBytes))
	isNested := strings.Contains(slashPath, "/")

	if isNested {
		domain := strings.Split(slashPath, "/")[0]
		nestedDomains[domain] = append(nestedDomains[domain], meta.ID)
	}

	s.registerOverride(meta, content, isNested)
	return nil
}

func (s *Store) registerOverride(meta TopicMetadata, content string, isNested bool) {
	existing, found := s.topics[meta.ID]
	title := meta.Title
	trigger := meta.Trigger

	if found {
		if title == meta.ID {
			title = existing.Title
		}
		if trigger == "Custom user override belief" {
			trigger = existing.Trigger
		}
	} else if !isNested {
		s.topicOrder = append(s.topicOrder, meta.ID)
	}

	s.topics[meta.ID] = TopicMetadata{
		ID:      meta.ID,
		Title:   title,
		Trigger: trigger,
		Version: PackageVersion + "-override",
		Origin:  "user_override",
		Content: content,
	}
}

func (s *Store) synthesizeDomainRouters(nestedDomains map[string][]string) {
	var domains []string
	for d := range nestedDomains {
		domains = append(domains, d)
	}
	sort.Strings(domains)

	for _, domain := range domains {
		if _, exists := s.topics[domain]; exists {
			continue // Author provided explicit top-level domain router (e.g. security.md)
		}

		children := nestedDomains[domain]
		routerContent := s.buildDomainRouterContent(domain, children)

		meta := TopicMetadata{
			ID:      domain,
			Title:   formatDomainTitle(domain),
			Trigger: fmt.Sprintf("Before working in the %s domain or with %s sub-rules", domain, domain),
			Version: PackageVersion + "-override",
			Origin:  "user_override",
			Content: routerContent,
		}

		s.topics[domain] = meta
		s.topicOrder = append(s.topicOrder, domain)
	}
}

func (s *Store) buildDomainRouterContent(domain string, children []string) string {
	var sb strings.Builder
	title := formatDomainTitle(domain)
	sb.WriteString(fmt.Sprintf("# %s\n\n", title))
	sb.WriteString(fmt.Sprintf("Scope: Domain router for `%s`. Call `beliefs_get(topics: [\"<sub-topic-id>\"])` to retrieve specific sub-rules.\n\n", domain))
	sb.WriteString("## Sub-Topics Router\n\n")
	sb.WriteString("| Sub-Topic ID | Title | When to Call |\n")
	sb.WriteString("| :--- | :--- | :--- |\n")

	for _, childID := range children {
		child := s.topics[childID]
		sb.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n", child.ID, child.Title, child.Trigger))
	}
	return sb.String()
}

func formatDomainTitle(domain string) string {
	if len(domain) == 0 {
		return "Domain Router"
	}
	return strings.ToUpper(domain[:1]) + domain[1:] + " (Domain Router)"
}

func parseOverrideMarkdown(defaultID, raw string) (TopicMetadata, string) {
	raw = strings.TrimSpace(raw)
	meta := TopicMetadata{
		ID:      defaultID,
		Title:   defaultID,
		Trigger: "Custom user override belief",
	}

	content := raw
	if strings.HasPrefix(raw, "---\n") || strings.HasPrefix(raw, "---\r\n") {
		meta, content = extractFrontmatter(raw, meta)
	}

	if meta.Title == defaultID {
		meta.Title = extractHeaderTitle(content, defaultID)
	}

	return meta, strings.TrimSpace(content)
}

func extractFrontmatter(raw string, base TopicMetadata) (TopicMetadata, string) {
	endIdx := strings.Index(raw[3:], "\n---")
	if endIdx == -1 {
		return base, raw
	}
	fmBody := raw[3 : 3+endIdx]
	rest := strings.TrimPrefix(raw[3+endIdx+4:], "\n")
	rest = strings.TrimPrefix(rest, "\r\n")

	for _, line := range strings.Split(fmBody, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "id:") {
			base.ID = cleanYAMLValue(strings.TrimPrefix(line, "id:"))
		} else if strings.HasPrefix(line, "title:") {
			base.Title = cleanYAMLValue(strings.TrimPrefix(line, "title:"))
		} else if strings.HasPrefix(line, "trigger:") {
			base.Trigger = cleanYAMLValue(strings.TrimPrefix(line, "trigger:"))
		} else if strings.HasPrefix(line, "when:") {
			base.Trigger = cleanYAMLValue(strings.TrimPrefix(line, "when:"))
		}
	}
	return base, rest
}

func cleanYAMLValue(val string) string {
	val = strings.TrimSpace(val)
	if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) ||
		(strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
		if len(val) >= 2 {
			val = val[1 : len(val)-1]
		}
	}
	return strings.TrimSpace(val)
}

func extractHeaderTitle(content, fallback string) string {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
		}
	}
	return fallback
}

func (s *Store) ListTopics(query string) []TopicMetadata {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []TopicMetadata
	tokens := tokenize(query)

	for _, id := range s.topicOrder {
		topic, exists := s.topics[id]
		if !exists {
			continue
		}
		if len(tokens) == 0 || matchesTokens(topic, tokens) {
			result = append(result, TopicMetadata{
				ID:      topic.ID,
				Title:   topic.Title,
				Trigger: topic.Trigger,
				Version: topic.Version,
				Origin:  topic.Origin,
			})
		}
	}

	return result
}

func (s *Store) GetTopics(topicIDs []string) ([]TopicMetadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []TopicMetadata
	for _, id := range topicIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		topic, exists := s.topics[id]
		if !exists {
			return nil, fmt.Errorf("topic '%s' not found. Available topics: %s", id, strings.Join(s.topicOrder, ", "))
		}
		result = append(result, topic)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no topic specified. Available topics: %s", strings.Join(s.topicOrder, ", "))
	}

	return result, nil
}

type SearchMatch struct {
	TopicID     string `json:"topic_id"`
	Title       string `json:"title"`
	Trigger     string `json:"trigger"`
	MatchedLine string `json:"matched_line,omitempty"`
	Score       int    `json:"score"`
}

type SearchResult struct {
	Matches       []SearchMatch   `json:"matches"`
	InlineContent []TopicMetadata `json:"inline_content,omitempty"`
}

func (s *Store) SearchTopics(query string, inlineContent bool, topK int) SearchResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tokens := tokenize(query)
	if len(tokens) == 0 {
		return SearchResult{}
	}

	var matches []SearchMatch
	for _, id := range s.topicOrder {
		topic := s.topics[id]
		score := 0
		lowerID := strings.ToLower(topic.ID)
		lowerTitle := strings.ToLower(topic.Title)
		lowerTrigger := strings.ToLower(topic.Trigger)

		for _, token := range tokens {
			if strings.Contains(lowerID, token) {
				score += 10
			}
			if strings.Contains(lowerTitle, token) {
				score += 8
			}
			if strings.Contains(lowerTrigger, token) {
				score += 5
			}
		}

		lines := strings.Split(topic.Content, "\n")
		var bestMatchedLine string

		for _, line := range lines {
			lowerLine := strings.ToLower(line)
			lineMatches := 0
			for _, token := range tokens {
				if strings.Contains(lowerLine, token) {
					lineMatches++
				}
			}
			if lineMatches > 0 {
				score += lineMatches * 2
				if bestMatchedLine == "" && len(strings.TrimSpace(line)) > 10 {
					bestMatchedLine = strings.TrimSpace(line)
				}
			}
		}

		if score > 0 {
			matches = append(matches, SearchMatch{
				TopicID:     topic.ID,
				Title:       topic.Title,
				Trigger:     topic.Trigger,
				MatchedLine: bestMatchedLine,
				Score:       score,
			})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	if len(matches) > 5 {
		matches = matches[:5]
	}

	res := SearchResult{Matches: matches}

	if inlineContent && len(matches) > 0 {
		if topK <= 0 {
			topK = 1
		}
		if topK > len(matches) {
			topK = len(matches)
		}
		for i := 0; i < topK; i++ {
			t := s.topics[matches[i].TopicID]
			res.InlineContent = append(res.InlineContent, t)
		}
	}

	return res
}

func (s *Store) GetBootstrapPayload() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString("# Cortex Agent Context: Beliefs Bootstrap\n\n")
	sb.WriteString("You are operating under the following foundational principles and topic guidance.\n\n")
	sb.WriteString("## Foundational Operating Principles (Always Active)\n\n")

	for i, p := range s.foundational {
		sb.WriteString(fmt.Sprintf("### %s\n\n%s\n\n", p.Title, p.Content))
		if i+1 < len(s.foundational) {
			sb.WriteString("---\n\n")
		}
	}

	sb.WriteString("---\n\n## Task-Triggered Topics Router\n\n")
	sb.WriteString("Before planning or executing any of the following tasks, call `beliefs_get(topics: [\"<id>\"])` to retrieve authoritative instructions:\n\n")
	sb.WriteString("| Topic ID | Topic Name | When to Call |\n")
	sb.WriteString("| :--- | :--- | :--- |\n")

	for _, id := range s.topicOrder {
		t := s.topics[id]
		sb.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n", t.ID, t.Title, t.Trigger))
	}

	sb.WriteString("\n### Usage Instruction\n")
	sb.WriteString("- When entering a task phase, call `beliefs_get(topics: [\"<id>\"])` (or `cortex_get_belief`). Batch multiple topics when crossing boundaries.\n")
	sb.WriteString("- Use `beliefs_search(query: \"<text>\")` with `inline_content: true` to instantly find and retrieve relevant rules.\n")

	return sb.String()
}

func (s *Store) GetInlineSnippet() string {
	return `<!-- CORTEX_BELIEFS_START -->
## Agent Operating Beliefs & Context

**MANDATORY PROTOCOL**: At the start of every session, you MUST call 'beliefs_bootstrap' (or 'cortex_bootstrap') to load the core operating principles (ODAV, Least-to-Most, Project Memory) and the active task-triggered topics router.

Before executing any task phase (e.g. coding, testing, git stacking, designing, opening MRs), you MUST consult the relevant topic(s) discovered from the bootstrap router using 'beliefs_get(topics: ["<topic-id>"])' (or 'cortex_get_belief').

**DEFAULT TO COMMIT**: Every verified code change or completed slice MUST be committed immediately as a git checkpoint before proceeding or claiming completion, unless explicitly instructed otherwise.
<!-- CORTEX_BELIEFS_END -->
`
}

func (s *Store) GetRenderedTableBlock() string {
	return s.GetInlineSnippet()
}

func tokenize(query string) []string {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(query)))
	return fields
}

func matchesTokens(topic TopicMetadata, tokens []string) bool {
	combined := strings.ToLower(fmt.Sprintf("%s %s %s", topic.ID, topic.Title, topic.Trigger))
	for _, token := range tokens {
		if strings.Contains(combined, token) {
			return true
		}
	}
	return false
}
