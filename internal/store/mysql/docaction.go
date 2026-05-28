package mysql

import (
	"context"
	"errors"
	"math"

	"github.com/viveksharavanan/workflow-as-service/internal/model"
)

func (s *MySQLStore) CreateDocAction(ctx context.Context, name string, reconfirm bool) (model.DocActionID, error) {
	if name == "" {
		return 0, errors.New("document action cannot be empty")
	}

	var flag int
	if reconfirm {
		flag = 1
	}

	res, err := s.db.ExecContext(ctx, "INSERT INTO wf_docactions_master(name, reconfirm) VALUES(?, ?)", name, flag)
	if err != nil {
		return 0, err
	}
	aid, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return model.DocActionID(aid), nil
}

func (s *MySQLStore) ListDocActions(ctx context.Context, offset, limit int64) ([]*model.DocAction, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit must be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	q := `
	SELECT id, name, reconfirm
	FROM wf_docactions_master
	ORDER BY id
	LIMIT ? OFFSET ?
	`
	rows, err := s.db.QueryContext(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*model.DocAction, 0, 10)
	for rows.Next() {
		var elem model.DocAction
		err = rows.Scan(&elem.ID, &elem.Name, &elem.Reconfirm)
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

func (s *MySQLStore) GetDocAction(ctx context.Context, id model.DocActionID) (*model.DocAction, error) {
	if id <= 0 {
		return nil, errors.New("ID should be a positive integer")
	}

	var elem model.DocAction
	row := s.db.QueryRowContext(ctx, "SELECT id, name, reconfirm FROM wf_docactions_master WHERE id = ?", id)
	err := row.Scan(&elem.ID, &elem.Name, &elem.Reconfirm)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

func (s *MySQLStore) GetDocActionByName(ctx context.Context, name string) (*model.DocAction, error) {
	if name == "" {
		return nil, errors.New("document action cannot be empty")
	}

	var elem model.DocAction
	row := s.db.QueryRowContext(ctx, "SELECT id, name, reconfirm FROM wf_docactions_master WHERE name = ?", name)
	err := row.Scan(&elem.ID, &elem.Name, &elem.Reconfirm)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

func (s *MySQLStore) RenameDocAction(ctx context.Context, id model.DocActionID, name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	_, err := s.db.ExecContext(ctx, "UPDATE wf_docactions_master SET name = ? WHERE id = ?", name, id)
	return err
}
