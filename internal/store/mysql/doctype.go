package mysql

import (
	"context"
	"errors"
	"math"

	"github.com/viveksharavanan/workflow-as-service/internal/model"
)

func (s *MySQLStore) CreateDocType(ctx context.Context, name string) (model.DocTypeID, error) {
	if name == "" {
		return 0, errors.New("name cannot be empty")
	}

	res, err := s.db.ExecContext(ctx, "INSERT INTO wf_doctypes_master(name) VALUES(?)", name)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return model.DocTypeID(id), nil
}

func (s *MySQLStore) ListDocTypes(ctx context.Context, offset, limit int64) ([]*model.DocType, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit must be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	q := `
	SELECT id, name
	FROM wf_doctypes_master
	ORDER BY id
	LIMIT ? OFFSET ?
	`
	rows, err := s.db.QueryContext(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*model.DocType, 0, 10)
	for rows.Next() {
		var elem model.DocType
		err = rows.Scan(&elem.ID, &elem.Name)
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

func (s *MySQLStore) GetDocType(ctx context.Context, id model.DocTypeID) (*model.DocType, error) {
	if id <= 0 {
		return nil, errors.New("ID should be a positive integer")
	}

	var elem model.DocType
	row := s.db.QueryRowContext(ctx, "SELECT id, name FROM wf_doctypes_master WHERE id = ?", id)
	err := row.Scan(&elem.ID, &elem.Name)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

func (s *MySQLStore) GetDocTypeByName(ctx context.Context, name string) (*model.DocType, error) {
	if name == "" {
		return nil, errors.New("document type cannot be empty")
	}

	var elem model.DocType
	row := s.db.QueryRowContext(ctx, "SELECT id, name FROM wf_doctypes_master WHERE name = ?", name)
	err := row.Scan(&elem.ID, &elem.Name)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

func (s *MySQLStore) RenameDocType(ctx context.Context, id model.DocTypeID, name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	_, err := s.db.ExecContext(ctx, "UPDATE wf_doctypes_master SET name = ? WHERE id = ?", name, id)
	return err
}

func (s *MySQLStore) ListTransitions(ctx context.Context, dtype model.DocTypeID, from model.DocStateID) (map[model.DocStateID]*model.TransitionMap, error) {
	q := `
	SELECT dst.from_state_id, dsm1.name, dst.docaction_id, dam.name, dam.reconfirm, dst.to_state_id, dsm2.name
	FROM wf_docstate_transitions dst
	JOIN wf_docstates_master dsm1 ON dsm1.id = dst.from_state_id
	JOIN wf_docstates_master dsm2 ON dsm2.id = dst.to_state_id
	JOIN wf_docactions_master dam ON dam.id = dst.docaction_id
	WHERE dst.doctype_id = ?
	`
	var args []interface{}
	args = append(args, dtype)
	if from > 0 {
		q += `AND dst.from_state_id = ?
		`
		args = append(args, from)
	}

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := map[model.DocStateID]*model.TransitionMap{}
	for rows.Next() {
		var dsfrom model.DocState
		var t model.Transition
		err := rows.Scan(&dsfrom.ID, &dsfrom.Name, &t.Upon.ID, &t.Upon.Name, &t.Upon.Reconfirm, &t.To.ID, &t.To.Name)
		if err != nil {
			return nil, err
		}

		var elem *model.TransitionMap
		ok := false
		if elem, ok = res[dsfrom.ID]; !ok {
			elem = &model.TransitionMap{}
			elem.From = dsfrom
			elem.Transitions = map[model.DocActionID]model.Transition{}
		}

		elem.Transitions[t.Upon.ID] = t
		res[dsfrom.ID] = elem
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *MySQLStore) GetTransitionTargets(ctx context.Context, dtype model.DocTypeID, state model.DocStateID) (map[model.DocActionID]model.DocStateID, error) {
	q := `
	SELECT docaction_id, to_state_id
	FROM wf_docstate_transitions
	WHERE doctype_id = ?
	AND from_state_id = ?
	`
	rows, err := s.db.QueryContext(ctx, q, dtype, state)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hash := make(map[model.DocActionID]model.DocStateID)
	for rows.Next() {
		var da model.DocActionID
		var ds model.DocStateID
		err := rows.Scan(&da, &ds)
		if err != nil {
			return nil, err
		}
		hash[da] = ds
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return hash, nil
}

func (s *MySQLStore) AddTransition(ctx context.Context, dtype model.DocTypeID, state model.DocStateID,
	action model.DocActionID, toState model.DocStateID) error {
	q := `
	INSERT INTO wf_docstate_transitions(doctype_id, from_state_id, docaction_id, to_state_id)
	VALUES(?, ?, ?, ?)
	`
	_, err := s.db.ExecContext(ctx, q, dtype, state, action, toState)
	return err
}

func (s *MySQLStore) RemoveTransition(ctx context.Context, dtype model.DocTypeID, state model.DocStateID, action model.DocActionID) error {
	q := `
	DELETE FROM wf_docstate_transitions
	WHERE doctype_id = ?
	AND from_state_id =?
	AND docaction_id = ?
	`
	_, err := s.db.ExecContext(ctx, q, dtype, state, action)
	return err
}
