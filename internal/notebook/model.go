package notebook

import "fmt"

const ContextID = "@context"

type Fault struct {
	Code            string `json:"code"`
	Message         string `json:"message"`
	CurrentRevision string `json:"current_revision,omitempty"`
}

func (e *Fault) Error() string { return e.Code + ": " + e.Message }

func invalid(field string) error {
	return &Fault{Code: "invalid_input", Message: fmt.Sprintf("invalid %s", field)}
}

type Evidence struct {
	Source     string `json:"source"`
	VerifiedAt string `json:"verified_at,omitempty"`
}

type Reference struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

type Fact struct {
	ID       string     `json:"id"`
	Title    string     `json:"title"`
	Body     string     `json:"body"`
	Topics   []string   `json:"topics"`
	Evidence []Evidence `json:"evidence"`
}

type ContextInput struct {
	Cursor string `json:"cursor,omitempty" jsonschema:"Optional continuation for older complete events; essential guidance is included only on the initial page."`
}

type ContextResult struct {
	Directory  string     `json:"directory"`
	Context    *ReadItem  `json:"context"`
	Recent     []ReadItem `json:"recent"`
	HasMore    bool       `json:"has_more"`
	NextCursor string     `json:"next_cursor,omitempty"`
}

type ReadItem struct {
	Reference
	Status   string `json:"status"`
	Revision string `json:"revision,omitempty"`
	Fact     *Fact  `json:"fact,omitempty"`
	Event    *Event `json:"event,omitempty"`
}
