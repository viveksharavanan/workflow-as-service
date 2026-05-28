package mysql

import (
	"context"
	"errors"
	"math"

	"github.com/viveksharavanan/workflow-as-service/internal/model"
)

func (s *MySQLStore) CountMailboxByUser(ctx context.Context, uid model.UserID, unread bool) (int64, error) {
	if uid <= 0 {
		return 0, errors.New("user ID should be a positive integer")
	}

	q := `
	SELECT COUNT(id)
	FROM wf_mailboxes
	WHERE group_id = (
		SELECT gm.id
		FROM wf_groups_master gm
		JOIN wf_group_users gu ON gu.group_id = gm.id
		WHERE gu.user_id = ?
		AND gm.group_type = 'S'
	)
	`
	if unread {
		q += `AND unread = 1`
	}

	row := s.db.QueryRowContext(ctx, q, uid)
	var n int64
	err := row.Scan(&n)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (s *MySQLStore) CountMailboxByGroup(ctx context.Context, gid model.GroupID, unread bool) (int64, error) {
	if gid <= 0 {
		return 0, errors.New("group ID should be a positive integer")
	}

	q := `
	SELECT COUNT(id)
	FROM wf_mailboxes
	WHERE group_id = ?
	`
	if unread {
		q += `AND unread = 1`
	}

	row := s.db.QueryRowContext(ctx, q, gid)
	var n int64
	err := row.Scan(&n)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (s *MySQLStore) ListMailboxByUser(ctx context.Context, uid model.UserID, offset, limit int64, unread bool) ([]*model.Notification, error) {
	if uid <= 0 {
		return nil, errors.New("user ID should be a positive integer")
	}
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit must be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	q := `
	SELECT mbs.group_id, msgs.id, msgs.doctype_id, dtm.name, msgs.doc_id, msgs.docevent_id, msgs.title, msgs.data, mbs.unread, mbs.ctime
	FROM wf_messages msgs
	JOIN wf_mailboxes mbs ON mbs.message_id = msgs.id
	JOIN wf_doctypes_master dtm ON dtm.id = msgs.doctype_id
	WHERE mbs.group_id = (
		SELECT gm.id
		FROM wf_groups_master gm
		JOIN wf_group_users gu ON gu.group_id = gm.id
		WHERE gu.user_id = ?
		AND gm.group_type = 'S'
	)
	`
	if unread {
		q += `AND mbs.unread = 1`
	}
	q += `
	ORDER BY msgs.id
	LIMIT ? OFFSET ?
	`

	rows, err := s.db.QueryContext(ctx, q, uid, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*model.Notification, 0, 10)
	for rows.Next() {
		var elem model.Notification
		err = rows.Scan(&elem.GroupID, &elem.Message.ID, &elem.Message.DocType.ID,
			&elem.Message.DocType.Name, &elem.Message.DocID, &elem.Message.Event,
			&elem.Message.Title, &elem.Message.Data, &elem.Unread, &elem.Ctime)
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

func (s *MySQLStore) ListMailboxByGroup(ctx context.Context, gid model.GroupID, offset, limit int64, unread bool) ([]*model.Notification, error) {
	if gid <= 0 {
		return nil, errors.New("group ID should be a positive integer")
	}
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit must be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	q := `
	SELECT mbs.group_id, msgs.id, msgs.doctype_id, dtm.name, msgs.doc_id, msgs.docevent_id, msgs.title, msgs.data, mbs.unread, mbs.ctime
	FROM wf_messages msgs
	JOIN wf_mailboxes mbs ON mbs.message_id = msgs.id
	JOIN wf_doctypes_master dtm ON dtm.id = msgs.doctype_id
	WHERE mbs.group_id = ?
	`
	if unread {
		q += `AND mbs.unread = 1`
	}
	q += `
	ORDER BY msgs.id
	LIMIT ? OFFSET ?
	`

	rows, err := s.db.QueryContext(ctx, q, gid, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*model.Notification, 0, 10)
	for rows.Next() {
		var elem model.Notification
		err = rows.Scan(&elem.GroupID, &elem.Message.ID, &elem.Message.DocType.ID,
			&elem.Message.DocType.Name, &elem.Message.DocID, &elem.Message.Event,
			&elem.Message.Title, &elem.Message.Data, &elem.Unread, &elem.Ctime)
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

func (s *MySQLStore) GetMailboxMessage(ctx context.Context, msgID model.MessageID) (*model.Notification, error) {
	if msgID <= 0 {
		return nil, errors.New("message ID should be positive integers")
	}

	q := `
	SELECT mbs.group_id, msgs.id, msgs.doctype_id, dtm.name, msgs.doc_id, msgs.docevent_id, msgs.title, msgs.data, mbs.unread, mbs.ctime
	FROM wf_messages msgs
	JOIN wf_mailboxes mbs ON mbs.message_id = msgs.id
	JOIN wf_doctypes_master dtm ON dtm.id = msgs.doctype_id
	WHERE mbs.id = ?
	`
	row := s.db.QueryRowContext(ctx, q, msgID)
	var elem model.Notification
	err := row.Scan(&elem.GroupID, &elem.Message.ID, &elem.Message.DocType.ID,
		&elem.Message.DocType.Name, &elem.Message.DocID, &elem.Message.Event,
		&elem.Message.Title, &elem.Message.Data, &elem.Unread, &elem.Ctime)
	if err != nil {
		return nil, err
	}

	return &elem, nil
}

func (s *MySQLStore) ReassignMessage(ctx context.Context, fromGID, toGID model.GroupID, msgID model.MessageID) error {
	if fromGID <= 0 || toGID <= 0 || msgID <= 0 {
		return errors.New("all identifiers should be positive integers")
	}
	if fromGID == toGID {
		return nil
	}

	q := `
	UPDATE wf_mailboxes SET group_id = ?, unread = 1
	WHERE group_id = ?
	AND message_id = ?
	`
	_, err := s.db.ExecContext(ctx, q, toGID, fromGID, msgID)
	return err
}

func (s *MySQLStore) SetMailboxStatusByUser(ctx context.Context, uid model.UserID, msgID model.MessageID, unread bool) error {
	if uid <= 0 || msgID <= 0 {
		return errors.New("all identifiers should be positive integers")
	}

	q := `
	UPDATE wf_mailboxes SET unread = ?
	WHERE group_id = (
		SELECT gm.id
		FROM wf_groups_master gm
		JOIN wf_group_users gu ON gu.group_id = gm.id
		WHERE gu.user_id = ?
		AND gm.group_type = 'S'
	)
	AND message_id = ?
	`
	_, err := s.db.ExecContext(ctx, q, unread, uid, msgID)
	return err
}

func (s *MySQLStore) SetMailboxStatusByGroup(ctx context.Context, gid model.GroupID, msgID model.MessageID, unread bool) error {
	if gid <= 0 || msgID <= 0 {
		return errors.New("all identifiers should be positive integers")
	}

	q := `
	UPDATE wf_mailboxes SET unread = ?
	WHERE group_id = ?
	AND message_id = ?
	`
	_, err := s.db.ExecContext(ctx, q, unread, gid, msgID)
	return err
}
