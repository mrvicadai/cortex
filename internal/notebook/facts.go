package notebook

import (
	"context"
	"database/sql"
	"time"
)

type PutInput struct {
	OperationID      string `json:"operation_id" jsonschema:"Stable retry key, reused only with the identical request."`
	ExpectedRevision string `json:"expected_revision" jsonschema:"Empty string creates a never-seen ID; otherwise use the current opaque revision from a complete read."`
	Fact             Fact   `json:"fact" jsonschema:"Durable active project invariant or system truth (not for completed milestones/tasks)."`
}

type RetireInput struct {
	OperationID      string     `json:"operation_id" jsonschema:"Stable retry key, reused only with the identical request."`
	ID               string     `json:"id" jsonschema:"ID of the fact to retire."`
	ExpectedRevision string     `json:"expected_revision" jsonschema:"Current opaque revision from a complete read."`
	Reason           string     `json:"reason,omitempty" jsonschema:"Optional explanation for the auto-appended retirement milestone event (defaults to 'Retired fact: <id>')."`
	Evidence         []Evidence `json:"evidence,omitempty" jsonschema:"Optional structured evidence pointers attached to the auto-appended event."`
	SkipEvent        bool       `json:"skip_event,omitempty" jsonschema:"Set true ONLY when retiring obsolete scratch or duplicate facts without needing an audit trail."`
}

func (s *Store) PutFact(ctx context.Context, in PutInput) (Receipt, error) {
	document, err := validateFact(in.Fact)
	if err != nil {
		return Receipt{}, err
	}
	if err := validateRevision(in.ExpectedRevision, true); err != nil {
		return Receipt{}, err
	}
	return s.mutate(ctx, s.project, in.OperationID, "fact_put", in, func(tx *sql.Tx) (Receipt, error) {
		if err := checkRevision(ctx, tx, s.project, in.Fact.ID, in.ExpectedRevision); err != nil {
			return Receipt{}, err
		}
		receipt := Receipt{Reference: Reference{Kind: "fact", ID: in.Fact.ID}, Revision: token()}
		_, err := tx.ExecContext(ctx, `INSERT INTO records(project_id,kind,id,revision,document,search_text) VALUES (?,'fact',?,?,?,?)
		 ON CONFLICT(project_id,kind,id) DO UPDATE SET revision=excluded.revision,retired=0,document=excluded.document,search_text=excluded.search_text`, s.project, in.Fact.ID, receipt.Revision, document, factSearch(in.Fact))
		return receipt, err
	})
}

func (s *Store) RetireFact(ctx context.Context, in RetireInput) (Receipt, error) {
	if err := validateRetireInput(in); err != nil {
		return Receipt{}, err
	}
	return s.mutate(ctx, s.project, in.OperationID, "fact_retire", in, func(tx *sql.Tx) (Receipt, error) {
		if err := checkRevision(ctx, tx, s.project, in.ID, in.ExpectedRevision); err != nil {
			return Receipt{}, err
		}
		receipt := Receipt{Reference: Reference{Kind: "fact", ID: in.ID}, Revision: token(), Retired: true}
		if _, err := tx.ExecContext(ctx, "UPDATE records SET revision=?,retired=1,document='',search_text='' WHERE project_id=? AND kind='fact' AND id=?", receipt.Revision, s.project, in.ID); err != nil {
			return Receipt{}, err
		}
		if !in.SkipEvent {
			body := in.Reason
			if body == "" {
				body = "Retired fact: " + in.ID
			}
			now := time.Now().UTC().Format(time.RFC3339Nano)
			if err := appendMilestoneEvent(ctx, tx, s.project, now, body, in.Evidence); err != nil {
				return Receipt{}, err
			}
		}
		return receipt, nil
	})
}

func validateRetireInput(in RetireInput) error {
	if !validName(in.ID, 128) {
		return invalid("id")
	}
	if in.ID == ContextID {
		return &Fault{Code: "context_required", Message: "replace essential guidance rather than retiring it"}
	}
	if err := validateRevision(in.ExpectedRevision, false); err != nil {
		return err
	}
	if in.Reason != "" {
		if err := validateBody(in.Reason); err != nil {
			return err
		}
	}
	if in.Evidence != nil {
		return validateEvidence(in.Evidence)
	}
	return nil
}
