package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/viveksharavanan/workflow-as-service/internal/model"
)

func (s *MySQLStore) ListNodes(ctx context.Context, wid model.WorkflowID) ([]*model.Node, error) {
	q := `
	SELECT id, doctype_id, docstate_id, workflow_id, name, type
	FROM wf_workflow_nodes
	WHERE workflow_id = ?
	`
	rows, err := s.db.QueryContext(ctx, q, wid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*model.Node, 0, 5)
	for rows.Next() {
		var elem model.Node
		err = rows.Scan(&elem.ID, &elem.DocType, &elem.State, &elem.Wflow, &elem.Name, &elem.NodeType)
		if err != nil {
			return nil, err
		}
		elem.SetFunc(nil) // sets default NodeFunc
		ary = append(ary, &elem)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ary, nil
}

func (s *MySQLStore) GetNode(ctx context.Context, id model.NodeID) (*model.Node, error) {
	if id <= 0 {
		return nil, errors.New("node ID must be a positive integer")
	}

	var elem model.Node
	var acID sql.NullInt64
	q := `
	SELECT id, doctype_id, docstate_id, ac_id, workflow_id, name, type
	FROM wf_workflow_nodes
	WHERE id = ?
	`
	row := s.db.QueryRowContext(ctx, q, id)
	err := row.Scan(&elem.ID, &elem.DocType, &elem.State, &acID, &elem.Wflow, &elem.Name, &elem.NodeType)
	if err != nil {
		return nil, err
	}
	if acID.Valid {
		elem.AccCtx = model.AccessContextID(acID.Int64)
	}

	elem.SetFunc(nil) // sets default NodeFunc
	return &elem, nil
}

func (s *MySQLStore) GetNodeByState(ctx context.Context, dtype model.DocTypeID, state model.DocStateID) (*model.Node, error) {
	var elem model.Node
	var acID sql.NullInt64
	q := `
	SELECT id, doctype_id, docstate_id, ac_id, workflow_id, name, type
	FROM wf_workflow_nodes
	WHERE doctype_id = ?
	AND docstate_id = ?
	`
	row := s.db.QueryRowContext(ctx, q, dtype, state)
	err := row.Scan(&elem.ID, &elem.DocType, &elem.State, &acID, &elem.Wflow, &elem.Name, &elem.NodeType)
	if err != nil {
		return nil, err
	}
	if acID.Valid {
		elem.AccCtx = model.AccessContextID(acID.Int64)
	}

	elem.SetFunc(nil) // sets default NodeFunc
	return &elem, nil
}

func (s *MySQLStore) GetReportingGroups(ctx context.Context, acid model.AccessContextID, gid model.GroupID) ([]model.GroupID, error) {
	q := `
	SELECT reports_to
	FROM wf_ac_group_hierarchy
	WHERE ac_id = ?
	AND group_id = ?
	ORDER BY group_id
	LIMIT 1
	`
	rows, err := s.db.QueryContext(ctx, q, acid, gid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.GroupID
	for rows.Next() {
		var id int64
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		result = append(result, model.GroupID(id))
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return result, nil
}

func (s *MySQLStore) GetEventParticipants(ctx context.Context, dtype model.DocTypeID, docID model.DocumentID) ([]model.GroupID, error) {
	q := `
	SELECT DISTINCT (group_id)
	FROM wf_docevents
	WHERE doctype_id = ?
	AND doc_id = ?
	`
	rows, err := s.db.QueryContext(ctx, q, dtype, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.GroupID
	for rows.Next() {
		var id int64
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		result = append(result, model.GroupID(id))
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return result, nil
}

func (s *MySQLStore) CreateMessage(ctx context.Context, msg *model.Message) (model.MessageID, error) {
	q := `
	INSERT INTO wf_messages(doctype_id, doc_id, docevent_id, title, data)
	VALUES(?, ?, ?, ?, ?)
	`
	res, err := s.db.ExecContext(ctx, q, msg.DocType.ID, msg.DocID, msg.Event, msg.Title, msg.Data)
	if err != nil {
		return 0, err
	}
	msgid, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return model.MessageID(msgid), nil
}

func (s *MySQLStore) DeliverMessage(ctx context.Context, msgID model.MessageID, recipients []model.GroupID) error {
	q := `
	INSERT INTO wf_mailboxes(group_id, message_id, unread, ctime)
	VALUES(?, ?, 1, NOW())
	`
	for _, gid := range recipients {
		_, err := s.db.ExecContext(ctx, q, gid, msgID)
		if err != nil {
			return err
		}
	}

	return nil
}
