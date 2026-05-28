package mysql

import (
	"context"
	"errors"
	"math"

	"github.com/viveksharavanan/workflow-as-service/internal/model"
)

func (s *MySQLStore) CreateDocState(ctx context.Context, name string) (model.DocStateID, error) {
	if name == "" {
		return 0, errors.New("name cannot be empty")
	}

	res, err := s.db.ExecContext(ctx, "INSERT INTO wf_docstates_master(name) VALUES(?)", name)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return model.DocStateID(id), nil
}

func (s *MySQLStore) ListDocStates(ctx context.Context, offset, limit int64) ([]*model.DocState, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit must be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	q := `
	SELECT id, name
	FROM wf_docstates_master
	ORDER BY id
	LIMIT ? OFFSET ?
	`
	rows, err := s.db.QueryContext(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*model.DocState, 0, 10)
	for rows.Next() {
		var elem model.DocState
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

func (s *MySQLStore) GetDocState(ctx context.Context, id model.DocStateID) (*model.DocState, error) {
	if id <= 0 {
		return nil, errors.New("ID should be a positive integer")
	}

	var elem model.DocState
	q := `
	SELECT name
	FROM wf_docstates_master
	WHERE id = ?
	`
	row := s.db.QueryRowContext(ctx, q, id)
	err := row.Scan(&elem.Name)
	if err != nil {
		return nil, err
	}

	elem.ID = id
	return &elem, nil
}

func (s *MySQLStore) GetDocStateByName(ctx context.Context, name string) (*model.DocState, error) {
	if name == "" {
		return nil, errors.New("document state name should be non-empty")
	}

	var elem model.DocState
	row := s.db.QueryRowContext(ctx, "SELECT id, name FROM wf_docstates_master WHERE name = ?", name)
	err := row.Scan(&elem.ID, &elem.Name)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

func (s *MySQLStore) RenameDocState(ctx context.Context, id model.DocStateID, name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	_, err := s.db.ExecContext(ctx, "UPDATE wf_docstates_master SET name = ? WHERE id = ?", name, id)
	return err
}
