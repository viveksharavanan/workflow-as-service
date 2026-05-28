package mysql

import (
	"context"
	"errors"
	"math"
	"strings"

	"github.com/viveksharavanan/workflow-as-service/internal/model"
)

func (s *MySQLStore) ListUsers(ctx context.Context, prefix string, offset, limit int64) ([]*model.User, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit must be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	var q string
	var args []interface{}

	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		q = `
		SELECT id, first_name, last_name, email, active
		FROM wf_users_master
		ORDER BY id
		LIMIT ? OFFSET ?
		`
		args = []interface{}{limit, offset}
	} else {
		q = `
		SELECT id, first_name, last_name, email, active
		FROM wf_users_master
		WHERE first_name LIKE ?
		UNION
		SELECT id, first_name, last_name, email, active
		FROM wf_users_master
		WHERE last_name LIKE ?
		ORDER BY id
		LIMIT ? OFFSET ?
		`
		args = []interface{}{prefix + "%", prefix + "%", limit, offset}
	}

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*model.User, 0, 10)
	for rows.Next() {
		var elem model.User
		err = rows.Scan(&elem.ID, &elem.FirstName, &elem.LastName, &elem.Email, &elem.Active)
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

func (s *MySQLStore) GetUser(ctx context.Context, uid model.UserID) (*model.User, error) {
	if uid <= 0 {
		return nil, errors.New("user ID should be a positive integer")
	}

	var elem model.User
	row := s.db.QueryRowContext(ctx, "SELECT id, first_name, last_name, email, active FROM wf_users_master WHERE id = ?", uid)
	err := row.Scan(&elem.ID, &elem.FirstName, &elem.LastName, &elem.Email, &elem.Active)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

func (s *MySQLStore) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return nil, errors.New("e-mail address should be non-empty")
	}

	var elem model.User
	row := s.db.QueryRowContext(ctx, "SELECT id, first_name, last_name, email, active FROM wf_users_master WHERE email = ?", email)
	err := row.Scan(&elem.ID, &elem.FirstName, &elem.LastName, &elem.Email, &elem.Active)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

func (s *MySQLStore) IsUserActive(ctx context.Context, uid model.UserID) (bool, error) {
	row := s.db.QueryRowContext(ctx, "SELECT active FROM wf_users_master WHERE id = ?", uid)
	var active bool
	err := row.Scan(&active)
	if err != nil {
		return false, err
	}

	return active, nil
}

func (s *MySQLStore) GetUserGroups(ctx context.Context, uid model.UserID) ([]*model.Group, error) {
	q := `
	SELECT gm.id, gm.name, gm.group_type
	FROM wf_groups_master gm
	JOIN wf_group_users gus ON gus.group_id = gm.id
	JOIN wf_users_master um ON um.id = gus.user_id
	WHERE um.id = ?
	`
	rows, err := s.db.QueryContext(ctx, q, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*model.Group, 0, 2)
	for rows.Next() {
		var elem model.Group
		err = rows.Scan(&elem.ID, &elem.Name, &elem.GroupType)
		if err != nil {
			return nil, err
		}
		ary = append(ary, &elem)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return ary, nil
}

func (s *MySQLStore) GetUserSingletonGroup(ctx context.Context, uid model.UserID) (*model.Group, error) {
	q := `
	SELECT gm.id, gm.name, gm.group_type
	FROM wf_groups_master gm
	JOIN wf_group_users gu ON gu.group_id = gm.id
	WHERE gu.user_id = ?
	AND gm.group_type = 'S'
	`
	var elem model.Group
	row := s.db.QueryRowContext(ctx, q, uid)
	err := row.Scan(&elem.ID, &elem.Name, &elem.GroupType)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}
