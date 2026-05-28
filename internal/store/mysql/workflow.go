package mysql

import (
	"context"
	"errors"
	"math"

	"github.com/viveksharavanan/workflow-as-service/internal/model"
)

func (s *MySQLStore) CreateWorkflow(ctx context.Context, name string, dtype model.DocTypeID, state model.DocStateID) (model.WorkflowID, error) {
	if name == "" {
		return 0, errors.New("name should not be empty")
	}
	if dtype <= 0 {
		return 0, errors.New("document type should be a positive integer")
	}
	if state <= 1 {
		return 0, errors.New("initial document state should be an integer > 1")
	}

	q := `
	INSERT INTO wf_workflows(name, doctype_id, docstate_id, active)
	VALUES(?, ?, ?, 1)
	`
	res, err := s.db.ExecContext(ctx, q, name, dtype, state)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return model.WorkflowID(id), nil
}

func (s *MySQLStore) ListWorkflows(ctx context.Context, offset, limit int64) ([]*model.Workflow, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit must be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	q := `
	SELECT wf.id, wf.name, dtm.id, dtm.name, dsm.id, dsm.name, wf.active
	FROM wf_workflows wf
	JOIN wf_doctypes_master dtm ON wf.doctype_id = dtm.id
	JOIN wf_docstates_master dsm ON wf.docstate_id = dsm.id
	ORDER BY wf.id
	LIMIT ? OFFSET ?
	`
	rows, err := s.db.QueryContext(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*model.Workflow, 0, 10)
	for rows.Next() {
		var elem model.Workflow
		err = rows.Scan(&elem.ID, &elem.Name, &elem.DocType.ID, &elem.DocType.Name,
			&elem.BeginState.ID, &elem.BeginState.Name, &elem.Active)
		if err != nil {
			return nil, err
		}
		ary = append(ary, &elem)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ary, nil
}

func (s *MySQLStore) GetWorkflow(ctx context.Context, id model.WorkflowID) (*model.Workflow, error) {
	q := `
	SELECT wf.id, wf.name, dtm.id, dtm.name, dsm.id, dsm.name, wf.active
	FROM wf_workflows wf
	JOIN wf_doctypes_master dtm ON dtm.id = wf.doctype_id
	JOIN wf_docstates_master dsm ON dsm.id = wf.docstate_id
	WHERE wf.id = ?
	`
	row := s.db.QueryRowContext(ctx, q, id)
	var elem model.Workflow
	err := row.Scan(&elem.ID, &elem.Name, &elem.DocType.ID, &elem.DocType.Name,
		&elem.BeginState.ID, &elem.BeginState.Name, &elem.Active)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

func (s *MySQLStore) GetWorkflowByDocType(ctx context.Context, dtid model.DocTypeID) (*model.Workflow, error) {
	q := `
	SELECT wf.id, wf.name, dtm.id, dtm.name, dsm.id, dsm.name, wf.active
	FROM wf_workflows wf
	JOIN wf_doctypes_master dtm ON dtm.id = wf.doctype_id
	JOIN wf_docstates_master dsm ON dsm.id = wf.docstate_id
	WHERE wf.doctype_id = ?
	`
	row := s.db.QueryRowContext(ctx, q, dtid)
	var elem model.Workflow
	err := row.Scan(&elem.ID, &elem.Name, &elem.DocType.ID, &elem.DocType.Name,
		&elem.BeginState.ID, &elem.BeginState.Name, &elem.Active)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

func (s *MySQLStore) GetWorkflowByName(ctx context.Context, name string) (*model.Workflow, error) {
	q := `
	SELECT wf.id, wf.name, dtm.id, dtm.name, dsm.id, dsm.name, wf.active
	FROM wf_workflows wf
	JOIN wf_doctypes_master dtm ON wf.doctype_id = dtm.id
	JOIN wf_docstates_master dsm ON wf.docstate_id = dsm.id
	WHERE wf.name = ?
	`
	row := s.db.QueryRowContext(ctx, q, name)
	var elem model.Workflow
	err := row.Scan(&elem.ID, &elem.Name, &elem.DocType.ID, &elem.DocType.Name,
		&elem.BeginState.ID, &elem.BeginState.Name, &elem.Active)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

func (s *MySQLStore) RenameWorkflow(ctx context.Context, id model.WorkflowID, name string) error {
	if name == "" {
		return errors.New("name should be non-empty")
	}

	q := `
	UPDATE wf_workflows SET name = ?
	WHERE id = ?
	`
	_, err := s.db.ExecContext(ctx, q, name, id)
	return err
}

func (s *MySQLStore) SetWorkflowActive(ctx context.Context, id model.WorkflowID, active bool) error {
	var flag int
	if active {
		flag = 1
	}
	q := `
	UPDATE wf_workflows SET active = ?
	WHERE id = ?
	`
	_, err := s.db.ExecContext(ctx, q, flag, id)
	return err
}

func (s *MySQLStore) AddWorkflowNode(ctx context.Context, dtype model.DocTypeID, state model.DocStateID,
	ac model.AccessContextID, wid model.WorkflowID, name string, ntype model.NodeType) (model.NodeID, error) {
	if name == "" {
		return 0, errors.New("name should not be empty")
	}

	q := `
	INSERT INTO wf_workflow_nodes(doctype_id, docstate_id, ac_id, workflow_id, name, type)
	VALUES(?, ?, ?, ?, ?, ?)
	`
	res, err := s.db.ExecContext(ctx, q, dtype, state, ac, wid, name, string(ntype))
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return model.NodeID(id), nil
}

func (s *MySQLStore) RemoveWorkflowNode(ctx context.Context, wid model.WorkflowID, nid model.NodeID) error {
	q := `
	DELETE FROM wf_workflow_nodes
	WHERE workflow_id = ?
	AND id = ?
	`
	_, err := s.db.ExecContext(ctx, q, wid, nid)
	return err
}
