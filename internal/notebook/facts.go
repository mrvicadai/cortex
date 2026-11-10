package notebook

import (
	"context"
	"database/sql"
)

type PutInput struct {
	OperationID      string `json:"operation_id" jsonschema:"Stable retry key, reused only with the identical request."`
	ExpectedRevision string `json:"expected_revision" jsonschema:"Empty string creates a never-seen ID; otherwise use the current opaque revision from a complete read."`
	Fact             Fact   `json:"fact"`
}

type RetireInput struct {
	OperationID      string `json:"operation_id"`
	ID               string `json:"id"`
	ExpectedRevision string `json:"expected_revision"`
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
	if !validName(in.ID, 128) {
		return Receipt{}, invalid("id")
	}
	if in.ID == ContextID {
		return Receipt{}, &Fault{Code: "context_required", Message: "replace essential guidance rather than retiring it"}
	}
	if err := validateRevision(in.ExpectedRevision, false); err != nil {
		return Receipt{}, err
	}
	return s.mutate(ctx, s.project, in.OperationID, "fact_retire", in, func(tx *sql.Tx) (Receipt, error) {
		if err := checkRevision(ctx, tx, s.project, in.ID, in.ExpectedRevision); err != nil {
			return Receipt{}, err
		}
		receipt := Receipt{Reference: Reference{Kind: "fact", ID: in.ID}, Revision: token(), Retired: true}
		_, err := tx.ExecContext(ctx, "UPDATE records SET revision=?,retired=1,document='',search_text='' WHERE project_id=? AND kind='fact' AND id=?", receipt.Revision, s.project, in.ID)
		return receipt, err
	})
}
