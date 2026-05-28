package mysql

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"strings"

	"github.com/viveksharavanan/workflow-as-service/internal/model"
)

func (s *MySQLStore) CreateAccessContext(ctx context.Context, name string) (model.AccessContextID, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, errors.New("access context name should be non-empty")
	}

	q := `INSERT INTO wf_access_contexts(name, active) VALUES(?, 1)`
	res, err := s.db.ExecContext(ctx, q, name)
	if err != nil {
		return 0, err
	}
	acID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return model.AccessContextID(acID), nil
}

func (s *MySQLStore) ListAccessContexts(ctx context.Context, prefix string, offset, limit int64) ([]*model.AccessContext, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit should be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	var q string
	var args []interface{}

	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		q = `
		SELECT id, name, active
		FROM wf_access_contexts
		ORDER BY id
		LIMIT ? OFFSET ?
		`
		args = []interface{}{limit, offset}
	} else {
		q = `
		SELECT id, name, active
		FROM wf_access_contexts
		WHERE name LIKE ?
		ORDER BY id
		LIMIT ? OFFSET ?
		`
		args = []interface{}{prefix + "%", limit, offset}
	}

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*model.AccessContext, 0, 10)
	for rows.Next() {
		var elem model.AccessContext
		err = rows.Scan(&elem.ID, &elem.Name, &elem.Active)
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

func (s *MySQLStore) ListAccessContextsByGroup(ctx context.Context, gid model.GroupID, offset, limit int64) ([]*model.AccessContext, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit should be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	q := `
	SELECT ac.id, ac.name, ac.active
	FROM wf_access_contexts ac
	JOIN wf_ac_group_hierarchy agh ON agh.ac_id = ac.id
	WHERE agh.group_id = ?
	ORDER BY agh.ac_id
	LIMIT ? OFFSET ?
	`
	rows, err := s.db.QueryContext(ctx, q, gid, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*model.AccessContext, 0, 10)
	for rows.Next() {
		var elem model.AccessContext
		err = rows.Scan(&elem.ID, &elem.Name, &elem.Active)
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

func (s *MySQLStore) ListAccessContextsByUser(ctx context.Context, uid model.UserID, offset, limit int64) ([]*model.AccessContext, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit should be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	q := `
	SELECT ac.id, ac.name, ac.active
	FROM wf_access_contexts ac
	JOIN wf_ac_group_hierarchy agh ON agh.ac_id = ac.id
	WHERE agh.group_id = (
		SELECT gm.id
		FROM wf_groups_master gm
		JOIN wf_group_users gu ON gu.group_id = gm.id
		WHERE gu.user_id = ?
		AND gm.group_type = 'S'
	)
	ORDER BY agh.ac_id
	LIMIT ? OFFSET ?
	`
	rows, err := s.db.QueryContext(ctx, q, uid, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*model.AccessContext, 0, 10)
	for rows.Next() {
		var elem model.AccessContext
		err = rows.Scan(&elem.ID, &elem.Name, &elem.Active)
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

func (s *MySQLStore) GetAccessContext(ctx context.Context, id model.AccessContextID) (*model.AccessContext, error) {
	q := `
	SELECT id, name, active
	FROM wf_access_contexts
	WHERE id = ?
	`
	res := s.db.QueryRowContext(ctx, q, id)
	var elem model.AccessContext
	err := res.Scan(&elem.ID, &elem.Name, &elem.Active)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

func (s *MySQLStore) RenameAccessContext(ctx context.Context, id model.AccessContextID, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("access context name should be non-empty")
	}

	q := `
	UPDATE wf_access_contexts
	SET name = ?
	WHERE id = ?
	`
	_, err := s.db.ExecContext(ctx, q, name, id)
	return err
}

func (s *MySQLStore) SetAccessContextActive(ctx context.Context, id model.AccessContextID, active bool) error {
	act := 0
	if active {
		act = 1
	}

	q := `
	UPDATE wf_access_contexts
	SET active = ?
	WHERE id = ?
	`
	_, err := s.db.ExecContext(ctx, q, act, id)
	return err
}

func (s *MySQLStore) GetAccessContextGroupRoles(ctx context.Context, id model.AccessContextID, gids []model.GroupID, offset, limit int64) (map[model.GroupID]*model.AcGroupRoles, error) {
	if id <= 0 {
		return nil, errors.New("access context ID should be a positive integer")
	}
	if len(gids) == 0 {
		return nil, errors.New("list of group IDs should be non-empty")
	}
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit should be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	args := make([]interface{}, 0, len(gids)+3)
	args = append(args, id)
	for _, gid := range gids {
		args = append(args, gid)
	}
	args = append(args, limit)
	args = append(args, offset)

	q := `
	SELECT agrs.group_id, gm.name, agrs.role_id, rm.name
	FROM wf_ac_group_roles agrs
	JOIN wf_groups_master gm ON gm.id = agrs.group_id
	JOIN wf_roles_master rm ON rm.id = agrs.role_id
	WHERE agrs.ac_id = ?
	AND agrs.group_id IN (?` + strings.Repeat(",?", len(gids)-1) + `)
	ORDER BY agrs.group_id
	LIMIT ? OFFSET ?
	`
	stmt, err := s.db.PrepareContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	grs := make(map[model.GroupID]*model.AcGroupRoles)
	var curGid int64
	for rows.Next() {
		var gid int64
		var gname string
		var role model.Role
		err = rows.Scan(&gid, &gname, &role.ID, &role.Name)
		if err != nil {
			return nil, err
		}

		var gr *model.AcGroupRoles
		if curGid == gid {
			gr = grs[model.GroupID(gid)]
		} else {
			gr = &model.AcGroupRoles{Group: gname, Roles: make([]model.Role, 0, 4)}
			curGid = gid
		}
		gr.Roles = append(gr.Roles, role)
		grs[model.GroupID(gid)] = gr
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return grs, nil
}

func (s *MySQLStore) AddAccessContextGroupRole(ctx context.Context, id model.AccessContextID, gid model.GroupID, rid model.RoleID) error {
	if gid <= 0 || rid <= 0 {
		return errors.New("group ID and role ID should be positive integers")
	}

	_, err := s.db.ExecContext(ctx, `INSERT INTO wf_ac_group_roles(ac_id, group_id, role_id) VALUES(?, ?, ?)`, id, gid, rid)
	return err
}

func (s *MySQLStore) RemoveAccessContextGroupRole(ctx context.Context, id model.AccessContextID, gid model.GroupID, rid model.RoleID) error {
	if gid <= 0 || rid <= 0 {
		return errors.New("group ID and role ID should be positive integers")
	}

	_, err := s.db.ExecContext(ctx, `DELETE FROM wf_ac_group_roles WHERE ac_id = ? AND group_id = ? AND role_id = ?`, id, gid, rid)
	return err
}

func (s *MySQLStore) ListAccessContextGroups(ctx context.Context, id model.AccessContextID, offset, limit int64) (map[model.GroupID]*model.AcGroup, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit should be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	q := `
	SELECT gm.id, gm.name, gm.group_type, rep_to.id, rep_to.name, rep_to.group_type
	FROM wf_groups_master gm
	JOIN wf_ac_group_hierarchy auh ON auh.group_id = gm.id
	JOIN wf_groups_master rep_to ON rep_to.id = auh.reports_to
	WHERE auh.ac_id = ?
	ORDER BY auh.group_id
	LIMIT ? OFFSET ?
	`
	rows, err := s.db.QueryContext(ctx, q, id, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	gh := make(map[model.GroupID]*model.AcGroup)
	for rows.Next() {
		var g model.AcGroup
		err = rows.Scan(&g.ID, &g.Name, &g.GroupType, &g.ReportsTo.ID, &g.ReportsTo.Name, &g.ReportsTo.GroupType)
		if err != nil {
			return nil, err
		}

		gh[model.GroupID(g.ID)] = &g
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return gh, nil
}

func (s *MySQLStore) AddAccessContextGroup(ctx context.Context, id model.AccessContextID, gid, reportsTo model.GroupID) error {
	if gid <= 0 || reportsTo < 0 {
		return errors.New("group ID should be a positive integer; reporting authority ID should be a non-negative integer")
	}

	q := `INSERT INTO wf_ac_group_hierarchy(ac_id, group_id, reports_to) VALUES (?, ?, ?)`
	_, err := s.db.ExecContext(ctx, q, id, gid, reportsTo)
	return err
}

func (s *MySQLStore) DeleteAccessContextGroup(ctx context.Context, id model.AccessContextID, gid model.GroupID) error {
	if gid <= 0 {
		return errors.New("user ID should be positive integer")
	}

	q := `DELETE FROM wf_ac_group_hierarchy WHERE ac_id = ? AND group_id = ?`
	_, err := s.db.ExecContext(ctx, q, id, gid)
	return err
}

func (s *MySQLStore) GetAccessContextGroupReportsTo(ctx context.Context, id model.AccessContextID, gid model.GroupID) (model.GroupID, error) {
	q := `
	SELECT reports_to
	FROM wf_ac_group_hierarchy
	WHERE ac_id = ?
	AND group_id = ?
	`
	row := s.db.QueryRowContext(ctx, q, id, gid)
	var repID int64
	err := row.Scan(&repID)
	if err != nil {
		return 0, err
	}

	return model.GroupID(repID), nil
}

func (s *MySQLStore) ListAccessContextGroupReportees(ctx context.Context, id model.AccessContextID, gid model.GroupID) ([]model.GroupID, error) {
	q := `
	SELECT group_id
	FROM wf_ac_group_hierarchy
	WHERE ac_id = ?
	AND reports_to = ?
	`
	rows, err := s.db.QueryContext(ctx, q, id, gid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]model.GroupID, 0, 4)
	for rows.Next() {
		var repID int64
		err = rows.Scan(&repID)
		if err != nil {
			return nil, err
		}
		ary = append(ary, model.GroupID(repID))
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ary, nil
}

func (s *MySQLStore) ChangeAccessContextReporting(ctx context.Context, id model.AccessContextID, gid, reportsTo model.GroupID) error {
	if gid <= 0 || reportsTo < 0 {
		return errors.New("group ID should be positive integer; reporting authority ID should be a non-negative integer")
	}

	q := `
	UPDATE wf_ac_group_hierarchy
	SET reports_to = ?
	WHERE ac_id = ?
	AND group_id = ?
	`
	_, err := s.db.ExecContext(ctx, q, reportsTo, id, gid)
	return err
}

func (s *MySQLStore) AccessContextIncludesGroup(ctx context.Context, id model.AccessContextID, gid model.GroupID) (bool, error) {
	if gid <= 0 {
		return false, errors.New("group ID should be a positive integer")
	}

	q := `
	SELECT reports_to
	FROM wf_ac_group_hierarchy
	WHERE ac_id = ?
	AND group_id = ?
	`
	var repTo int64
	row := s.db.QueryRowContext(ctx, q, id, gid)
	err := row.Scan(&repTo)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (s *MySQLStore) AccessContextIncludesUser(ctx context.Context, id model.AccessContextID, uid model.UserID) (bool, error) {
	if uid <= 0 {
		return false, errors.New("user ID should be a positive integer")
	}

	q := `
	SELECT COUNT(agh.reports_to)
	FROM wf_ac_group_hierarchy agh
	WHERE agh.ac_id = ?
	AND agh.group_id IN (
		SELECT gm.id
		FROM wf_groups_master gm
		JOIN wf_group_users gu ON gu.group_id = gm.id
		WHERE gu.user_id = ?
	)
	`
	var count int64
	row := s.db.QueryRowContext(ctx, q, id, uid)
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}

	if count == 0 {
		return false, nil
	}

	return true, nil
}

func (s *MySQLStore) GetUserPermissions(ctx context.Context, id model.AccessContextID, uid model.UserID) (map[model.DocTypeID][]model.DocAction, error) {
	if uid <= 0 {
		return nil, errors.New("user ID should be a positive integer")
	}

	q := `
	SELECT acpv.doctype_id, acpv.docaction_id, dam.name, dam.reconfirm
	FROM wf_ac_perms_v acpv
	JOIN wf_docactions_master dam ON dam.id = acpv.docaction_id
	WHERE acpv.ac_id = ?
	AND acpv.user_id = ?
	`
	rows, err := s.db.QueryContext(ctx, q, id, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := map[model.DocTypeID][]model.DocAction{}
	for rows.Next() {
		var dtid int64
		var da model.DocAction
		err = rows.Scan(&dtid, &da.ID, &da.Name, &da.Reconfirm)
		if err != nil {
			return nil, err
		}

		ary, ok := res[model.DocTypeID(dtid)]
		if !ok {
			ary = []model.DocAction{}
		}
		ary = append(ary, da)
		res[model.DocTypeID(dtid)] = ary
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return res, nil
}

func (s *MySQLStore) GetUserPermissionsByDocType(ctx context.Context, id model.AccessContextID, dtype model.DocTypeID, uid model.UserID) ([]model.DocAction, error) {
	if id <= 0 || dtype <= 0 || uid <= 0 {
		return nil, errors.New("all identifiers should be positive integers")
	}

	q := `
	SELECT acpv.docaction_id, dam.name, dam.reconfirm
	FROM wf_ac_perms_v acpv
	JOIN wf_docactions_master dam ON dam.id = acpv.docaction_id
	WHERE acpv.ac_id = ?
	AND acpv.doctype_id = ?
	AND acpv.user_id = ?
	`
	rows, err := s.db.QueryContext(ctx, q, id, dtype, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []model.DocAction{}
	for rows.Next() {
		var da model.DocAction
		err = rows.Scan(&da.ID, &da.Name, &da.Reconfirm)
		if err != nil {
			return nil, err
		}

		res = append(res, da)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return res, nil
}

func (s *MySQLStore) GetGroupPermissions(ctx context.Context, id model.AccessContextID, gid model.GroupID) (map[model.DocTypeID][]model.DocAction, error) {
	if gid <= 0 {
		return nil, errors.New("group ID should be a positive integer")
	}

	q := `
	SELECT acpv.doctype_id, acpv.docaction_id, dam.name, dam.reconfirm
	FROM wf_ac_perms_v acpv
	JOIN wf_docactions_master dam ON dam.id = acpv.docaction_id
	WHERE acpv.ac_id = ?
	AND acpv.group_id = ?
	`
	rows, err := s.db.QueryContext(ctx, q, id, gid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := map[model.DocTypeID][]model.DocAction{}
	for rows.Next() {
		var dtid int64
		var da model.DocAction
		err = rows.Scan(&dtid, &da.ID, &da.Name, &da.Reconfirm)
		if err != nil {
			return nil, err
		}

		ary, ok := res[model.DocTypeID(dtid)]
		if !ok {
			ary = []model.DocAction{}
		}
		ary = append(ary, da)
		res[model.DocTypeID(dtid)] = ary
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return res, nil
}

func (s *MySQLStore) GetGroupPermissionsByDocType(ctx context.Context, id model.AccessContextID, dtype model.DocTypeID, gid model.GroupID) ([]model.DocAction, error) {
	if id <= 0 || dtype <= 0 || gid <= 0 {
		return nil, errors.New("all identifiers should be positive integers")
	}

	q := `
	SELECT acpv.docaction_id, dam.name, dam.reconfirm
	FROM wf_ac_perms_v acpv
	JOIN wf_docactions_master dam ON dam.id = acpv.docaction_id
	WHERE acpv.ac_id = ?
	AND acpv.doctype_id = ?
	AND acpv.group_id = ?
	`
	rows, err := s.db.QueryContext(ctx, q, id, dtype, gid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []model.DocAction{}
	for rows.Next() {
		var da model.DocAction
		err = rows.Scan(&da.ID, &da.Name, &da.Reconfirm)
		if err != nil {
			return nil, err
		}

		res = append(res, da)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return res, nil
}

func (s *MySQLStore) UserHasPermission(ctx context.Context, id model.AccessContextID, uid model.UserID, dtype model.DocTypeID, action model.DocActionID) (bool, error) {
	if uid <= 0 || dtype <= 0 || action <= 0 {
		return false, errors.New("invalid user ID or document type or document action")
	}

	q := `
	SELECT role_id FROM wf_ac_perms_v
	WHERE ac_id = ?
	AND user_id = ?
	AND doctype_id = ?
	AND docaction_id = ?
	LIMIT 1
	`
	row := s.db.QueryRowContext(ctx, q, id, uid, dtype, action)
	var roleID int64
	err := row.Scan(&roleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *MySQLStore) GroupHasPermission(ctx context.Context, id model.AccessContextID, gid model.GroupID, dtype model.DocTypeID, action model.DocActionID) (bool, error) {
	if gid <= 0 || dtype <= 0 || action <= 0 {
		return false, errors.New("invalid group ID or document type or document action")
	}

	q := `
	SELECT role_id FROM wf_ac_perms_v
	WHERE ac_id = ?
	AND group_id = ?
	AND doctype_id = ?
	AND docaction_id = ?
	LIMIT 1
	`
	row := s.db.QueryRowContext(ctx, q, id, gid, dtype, action)
	var roleID int64
	err := row.Scan(&roleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
