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

func (s *MySQLStore) CreateDocEvent(ctx context.Context, input *model.DocEventsNewInput) (model.DocEventID, error) {
	if input.DocTypeID <= 0 || input.DocumentID <= 0 || input.DocStateID <= 0 || input.DocActionID <= 0 || input.GroupID <= 0 {
		return 0, errors.New("all identifiers should be positive integers")
	}
	if input.Text == "" {
		return 0, errors.New("please add comments or notes")
	}

	q := `
	INSERT INTO wf_docevents(doctype_id, doc_id, docstate_id, docaction_id, group_id, data, ctime, status)
	VALUES(?, ?, ?, ?, ?, ?, NOW(), 'P')
	`
	res, err := s.db.ExecContext(ctx, q, input.DocTypeID, input.DocumentID, input.DocStateID, input.DocActionID, input.GroupID, input.Text)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return model.DocEventID(id), nil
}

func (s *MySQLStore) ListDocEvents(ctx context.Context, input *model.DocEventsListInput, offset, limit int64) ([]*model.DocEvent, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit must be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	// Base query.
	q := `
	SELECT de.id, de.doctype_id, de.doc_id, de.docstate_id, de.docaction_id, de.group_id, de.data, de.ctime, de.status
	FROM wf_docevents de
	`

	// Process input specification.
	where := []string{}
	args := []interface{}{}

	if input.DocTypeID > 0 {
		where = append(where, `de.doctype_id = ?`)
		args = append(args, input.DocTypeID)
	}

	if input.AccessContextID > 0 {
		q += `JOIN wf_documents docs ON docs.doctype_id = de.doctype_id AND docs.id = de.doc_id
		`
		where = append(where, `docs.ac_id = ?`)
		args = append(args, input.AccessContextID)
	}

	switch input.Status {
	case model.EventStatusAll:
		// Intentionally left blank

	case model.EventStatusApplied:
		where = append(where, `status = 'A'`)

	case model.EventStatusPending:
		where = append(where, `status = 'P'`)

	default:
		return nil, fmt.Errorf("unknown event status specified in filter : %d", input.Status)
	}

	if input.GroupID > 0 {
		where = append(where, `de.group_id = ?`)
		args = append(args, input.GroupID)
	}

	if input.DocStateID > 0 {
		where = append(where, `de.docstate_id = ?`)
		args = append(args, input.DocStateID)
	}

	if !input.CtimeStarting.IsZero() {
		where = append(where, `de.ctime >= ?`)
		args = append(args, input.CtimeStarting)
	}

	if !input.CtimeBefore.IsZero() {
		where = append(where, `de.ctime < ?`)
		args = append(args, input.CtimeBefore)
	}

	if len(where) > 0 {
		q += ` WHERE ` + strings.Join(where, ` AND `)
	}

	q += `
	ORDER BY de.id
	LIMIT ? OFFSET ?
	`
	args = append(args, limit, offset)
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var text sql.NullString
	var dstatus string
	ary := make([]*model.DocEvent, 0, 10)
	for rows.Next() {
		var elem model.DocEvent
		err = rows.Scan(&elem.ID, &elem.DocType, &elem.DocID, &elem.State, &elem.Action, &elem.Group, &text, &elem.Ctime, &dstatus)
		if err != nil {
			return nil, err
		}
		if text.Valid {
			elem.Text = text.String
		}
		switch dstatus {
		case "A":
			elem.Status = model.EventStatusApplied
		case "P":
			elem.Status = model.EventStatusPending
		default:
			return nil, fmt.Errorf("unknown event status : %s", dstatus)
		}
		ary = append(ary, &elem)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ary, nil
}

func (s *MySQLStore) GetDocEvent(ctx context.Context, eid model.DocEventID) (*model.DocEvent, error) {
	if eid <= 0 {
		return nil, errors.New("event ID should be a positive integer")
	}

	var text sql.NullString
	var dstatus string
	var elem model.DocEvent
	q := `
	SELECT id, doctype_id, doc_id, docstate_id, docaction_id, group_id, data, ctime, status
	FROM wf_docevents
	WHERE id = ?
	`
	row := s.db.QueryRowContext(ctx, q, eid)
	err := row.Scan(&elem.ID, &elem.DocType, &elem.DocID, &elem.State, &elem.Action, &elem.Group, &text, &elem.Ctime, &dstatus)
	if err != nil {
		return nil, err
	}
	if text.Valid {
		elem.Text = text.String
	}
	switch dstatus {
	case "A":
		elem.Status = model.EventStatusApplied
	case "P":
		elem.Status = model.EventStatusPending
	default:
		return nil, fmt.Errorf("unknown event status : %s", dstatus)
	}

	return &elem, nil
}

func (s *MySQLStore) GetDocEventStatus(ctx context.Context, id model.DocEventID) (model.EventStatus, error) {
	var dstatus string
	row := s.db.QueryRowContext(ctx, "SELECT status FROM wf_docevents WHERE id = ?", id)
	err := row.Scan(&dstatus)
	if err != nil {
		return 0, err
	}
	switch dstatus {
	case "A":
		return model.EventStatusApplied, nil
	case "P":
		return model.EventStatusPending, nil
	default:
		return 0, fmt.Errorf("unknown event status : %s", dstatus)
	}
}

func (s *MySQLStore) SetDocEventStatus(ctx context.Context, id model.DocEventID, status string) error {
	q := `UPDATE wf_docevents SET status = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, q, status, id)
	return err
}

func (s *MySQLStore) RecordEventApplication(ctx context.Context, dtype model.DocTypeID, docID model.DocumentID, fromState model.DocStateID, eventID model.DocEventID, toState model.DocStateID, statusOnly bool) error {
	if !statusOnly {
		q := `
		INSERT INTO wf_docevent_application(doctype_id, doc_id, from_state_id, docevent_id, to_state_id)
		VALUES(?, ?, ?, ?, ?)
		`
		_, err := s.db.ExecContext(ctx, q, dtype, docID, fromState, eventID, toState)
		if err != nil {
			return err
		}
	}

	q := `UPDATE wf_docevents SET status = 'A' WHERE id = ?`
	_, err := s.db.ExecContext(ctx, q, eventID)
	return err
}
