package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/viveksharavanan/workflow-as-service/internal/model"
)

func (s *MySQLStore) CreateRole(ctx context.Context, name string) (model.RoleID, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, errors.New("name cannot not be empty")
	}

	res, err := s.db.ExecContext(ctx, "INSERT INTO wf_roles_master(name) VALUES(?)", name)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return model.RoleID(id), nil
}

func (s *MySQLStore) ListRoles(ctx context.Context, offset, limit int64) ([]*model.Role, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit must be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	q := `
	SELECT id, name
	FROM wf_roles_master
	ORDER BY id
	LIMIT ? OFFSET ?
	`
	rows, err := s.db.QueryContext(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*model.Role, 0, 10)
	for rows.Next() {
		var elem model.Role
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

func (s *MySQLStore) GetRole(ctx context.Context, id model.RoleID) (*model.Role, error) {
	if id <= 0 {
		return nil, errors.New("ID must be a positive integer")
	}

	var elem model.Role
	row := s.db.QueryRowContext(ctx, "SELECT id, name FROM wf_roles_master WHERE id = ?", id)
	err := row.Scan(&elem.ID, &elem.Name)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

func (s *MySQLStore) GetRoleByName(ctx context.Context, name string) (*model.Role, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("role cannot be empty")
	}

	var elem model.Role
	row := s.db.QueryRowContext(ctx, "SELECT id, name FROM wf_roles_master WHERE name = ?", name)
	err := row.Scan(&elem.ID, &elem.Name)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

func (s *MySQLStore) RenameRole(ctx context.Context, id model.RoleID, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("name cannot be empty")
	}

	_, err := s.db.ExecContext(ctx, "UPDATE wf_roles_master SET name = ? WHERE id = ?", name, id)
	return err
}

func (s *MySQLStore) DeleteRole(ctx context.Context, id model.RoleID) error {
	if id <= 0 {
		return errors.New("role ID must be a positive integer")
	}

	row := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM wf_ac_group_roles WHERE role_id = ?", id)
	var n int64
	err := row.Scan(&n)
	if err != nil {
		return err
	}
	if n > 0 {
		return errors.New("role is being used in at least one access context; cannot delete")
	}

	_, err = s.db.ExecContext(ctx, "DELETE FROM wf_role_docactions WHERE role_id = ?", id)
	if err != nil {
		return err
	}
	res, err := s.db.ExecContext(ctx, "DELETE FROM wf_roles_master WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, err = res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("expected number of affected rows : 1; actual affected : %d", n)
	}

	return nil
}

func (s *MySQLStore) AddRolePermissions(ctx context.Context, rid model.RoleID, dtype model.DocTypeID, actions []model.DocActionID) error {
	q := `
	INSERT INTO wf_role_docactions(role_id, doctype_id, docaction_id)
	VALUES(?, ?, ?)
	`
	for _, action := range actions {
		_, err := s.db.ExecContext(ctx, q, rid, dtype, action)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *MySQLStore) RemoveRolePermissions(ctx context.Context, rid model.RoleID, dtype model.DocTypeID, actions []model.DocActionID) error {
	q := `
	DELETE FROM wf_role_docactions
	WHERE role_id = ?
	AND doctype_id = ?
	AND docaction_id = ?
	`
	for _, action := range actions {
		_, err := s.db.ExecContext(ctx, q, rid, dtype, action)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *MySQLStore) ListRolePermissions(ctx context.Context, rid model.RoleID) (map[string]struct {
	DocTypeID model.DocTypeID
	Actions   []*model.DocAction
}, error) {
	q := `
	SELECT dtm.id, dtm.name, dam.id, dam.name, dam.reconfirm
	FROM wf_doctypes_master dtm
	JOIN wf_role_docactions rdas ON dtm.id = rdas.doctype_id
	JOIN wf_docactions_master dam ON dam.id = rdas.docaction_id
	WHERE rdas.role_id = ?
	`
	rows, err := s.db.QueryContext(ctx, q, rid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	das := make(map[string]struct {
		DocTypeID model.DocTypeID
		Actions   []*model.DocAction
	})
	for rows.Next() {
		var dt model.DocType
		var da model.DocAction
		err = rows.Scan(&dt.ID, &dt.Name, &da.ID, &da.Name, &da.Reconfirm)
		if err != nil {
			return nil, err
		}
		st, ok := das[dt.Name]
		if !ok {
			st.DocTypeID = dt.ID
			st.Actions = make([]*model.DocAction, 0, 1)
		}
		st.Actions = append(st.Actions, &da)
		das[dt.Name] = st
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return das, nil
}

func (s *MySQLStore) RoleHasPermission(ctx context.Context, rid model.RoleID, dtype model.DocTypeID, action model.DocActionID) (bool, error) {
	q := `
	SELECT rdas.id FROM wf_role_docactions rdas
	JOIN wf_doctypes_master dtm ON rdas.doctype_id = dtm.id
	JOIN wf_docactions_master dam ON rdas.docaction_id = dam.id
	WHERE rdas.role_id = ?
	AND dtm.id = ?
	AND dam.id = ?
	ORDER BY rdas.id
	LIMIT 1
	`
	row := s.db.QueryRowContext(ctx, q, rid, dtype, action)
	var n int64
	err := row.Scan(&n)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
