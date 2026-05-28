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

func (s *MySQLStore) CreateSingletonGroup(ctx context.Context, uid model.UserID) (model.GroupID, error) {
	q := `
	INSERT INTO wf_groups_master(name, group_type)
	SELECT u.email, 'S'
	FROM wf_users_master u
	WHERE u.id = ?
	`
	res, err := s.db.ExecContext(ctx, q, uid)
	if err != nil {
		return 0, err
	}
	gid, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	_, err = s.db.ExecContext(ctx, "INSERT INTO wf_group_users(group_id, user_id) VALUES(?, ?)", gid, uid)
	if err != nil {
		return 0, err
	}

	return model.GroupID(gid), nil
}

func (s *MySQLStore) CreateGroup(ctx context.Context, name string, gtype string) (model.GroupID, error) {
	name = strings.TrimSpace(name)
	gtype = strings.TrimSpace(gtype)
	if name == "" || gtype == "" {
		return 0, errors.New("group name and type must not be empty")
	}
	switch gtype {
	case "G": // General
	// Nothing to do

	default:
		return 0, errors.New("unknown group type")
	}

	res, err := s.db.ExecContext(ctx, "INSERT INTO wf_groups_master(name, group_type) VALUES(?, ?)", name, gtype)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return model.GroupID(id), nil
}

func (s *MySQLStore) ListGroups(ctx context.Context, offset, limit int64) ([]*model.Group, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit must be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	q := `
	SELECT id, name, group_type
	FROM wf_groups_master
	ORDER BY id
	LIMIT ? OFFSET ?
	`
	rows, err := s.db.QueryContext(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*model.Group, 0, 10)
	for rows.Next() {
		var g model.Group
		err = rows.Scan(&g.ID, &g.Name, &g.GroupType)
		if err != nil {
			return nil, err
		}
		ary = append(ary, &g)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ary, nil
}

func (s *MySQLStore) GetGroup(ctx context.Context, id model.GroupID) (*model.Group, error) {
	if id <= 0 {
		return nil, errors.New("group ID should be a positive integer")
	}

	var elem model.Group
	row := s.db.QueryRowContext(ctx, "SELECT id, name, group_type FROM wf_groups_master WHERE id = ?", id)
	err := row.Scan(&elem.ID, &elem.Name, &elem.GroupType)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

func (s *MySQLStore) GetGroupType(ctx context.Context, id model.GroupID) (string, error) {
	var gt string
	row := s.db.QueryRowContext(ctx, "SELECT group_type FROM wf_groups_master WHERE id = ?", id)
	err := row.Scan(&gt)
	if err != nil {
		return "", err
	}
	return gt, nil
}

func (s *MySQLStore) RenameGroup(ctx context.Context, id model.GroupID, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("name cannot be empty")
	}

	var elem model.Group
	row := s.db.QueryRowContext(ctx, "SELECT id, name, group_type FROM wf_groups_master WHERE id = ?", id)
	err := row.Scan(&elem.ID, &elem.Name, &elem.GroupType)
	if err != nil {
		return err
	}
	if elem.GroupType == "S" {
		return errors.New("cannot rename a singleton group")
	}

	_, err = s.db.ExecContext(ctx, "UPDATE wf_groups_master SET name = ? WHERE id = ?", name, id)
	return err
}

func (s *MySQLStore) DeleteGroup(ctx context.Context, id model.GroupID) error {
	if id <= 0 {
		return errors.New("group ID must be a positive integer")
	}

	row := s.db.QueryRowContext(ctx, "SELECT group_type FROM wf_groups_master WHERE id = ?", id)
	var gtype string
	err := row.Scan(&gtype)
	if err != nil {
		return err
	}
	if gtype == "S" {
		return errors.New("singleton groups cannot be deleted")
	}

	row = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM wf_ac_group_roles WHERE group_id = ?", id)
	var n int64
	err = row.Scan(&n)
	if err != nil {
		return err
	}
	if n > 0 {
		return errors.New("group is being used in at least one access context; cannot delete")
	}

	_, err = s.db.ExecContext(ctx, "DELETE FROM wf_group_users WHERE group_id = ?", id)
	if err != nil {
		return err
	}
	res, err := s.db.ExecContext(ctx, "DELETE FROM wf_groups_master WHERE id = ?", id)
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

func (s *MySQLStore) ListGroupUsers(ctx context.Context, gid model.GroupID) ([]*model.User, error) {
	q := `
	SELECT um.id, um.first_name, um.last_name, um.email, um.active
	FROM wf_users_master um
	JOIN wf_group_users gu ON gu.user_id = um.id
	WHERE gu.group_id = ?
	`
	rows, err := s.db.QueryContext(ctx, q, gid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*model.User, 0, 2)
	for rows.Next() {
		var elem model.User
		err = rows.Scan(&elem.ID, &elem.FirstName, &elem.LastName, &elem.Email, &elem.Active)
		if err != nil {
			return nil, err
		}
		ary = append(ary, &elem)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return ary, nil
}

func (s *MySQLStore) GroupHasUser(ctx context.Context, gid model.GroupID, uid model.UserID) (bool, error) {
	q := `
	SELECT id FROM wf_group_users
	WHERE group_id = ?
	AND user_id = ?
	ORDER BY id
	LIMIT 1
	`
	var id int64
	row := s.db.QueryRowContext(ctx, q, gid, uid)
	err := row.Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		return false, errors.New("given user is not part of the specified group")
	case err != nil:
		return false, err
	default:
		return true, nil
	}
}

func (s *MySQLStore) GetSingletonUser(ctx context.Context, gid model.GroupID) (*model.User, error) {
	q := `
	SELECT um.id, um.first_name, um.last_name, um.email, um.active
	FROM wf_users_master um
	JOIN wf_group_users gus ON gus.user_id = um.id
	JOIN wf_groups_master gm ON gus.group_id = gm.id
	WHERE gm.id = ?
	AND gm.group_type = 'S'
	ORDER BY um.id
	LIMIT 1
	`

	var elem model.User
	row := s.db.QueryRowContext(ctx, q, gid)
	err := row.Scan(&elem.ID, &elem.FirstName, &elem.LastName, &elem.Email, &elem.Active)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

func (s *MySQLStore) AddGroupUser(ctx context.Context, gid model.GroupID, uid model.UserID) error {
	if gid <= 0 || uid <= 0 {
		return errors.New("group ID and user ID must be positive integers")
	}

	var gtype string
	row := s.db.QueryRowContext(ctx, "SELECT group_type FROM wf_groups_master WHERE id = ?", gid)
	err := row.Scan(&gtype)
	if err != nil {
		return err
	}
	if gtype == "S" {
		return errors.New("cannot add users to singleton groups")
	}

	_, err = s.db.ExecContext(ctx, "INSERT INTO wf_group_users(group_id, user_id) VALUES(?, ?)", gid, uid)
	return err
}

func (s *MySQLStore) RemoveGroupUser(ctx context.Context, gid model.GroupID, uid model.UserID) error {
	if gid <= 0 || uid <= 0 {
		return errors.New("group ID and user ID must be positive integers")
	}

	var gtype string
	row := s.db.QueryRowContext(ctx, "SELECT group_type FROM wf_groups_master WHERE id = ?", gid)
	err := row.Scan(&gtype)
	if err != nil {
		return err
	}
	if gtype == "S" {
		return errors.New("cannot remove users from singleton groups")
	}

	res, err := s.db.ExecContext(ctx, "DELETE FROM wf_group_users WHERE group_id = ? AND user_id = ?", gid, uid)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("expected number of affected rows : 1; actual affected : %d", n)
	}

	return nil
}
