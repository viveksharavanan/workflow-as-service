package mysql

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"strings"

	"github.com/viveksharavanan/workflow-as-service/internal/model"
)

func (s *MySQLStore) CreateDocument(ctx context.Context, input *model.DocumentsNewInput) (model.DocumentID, error) {
	if input.DocTypeID <= 0 || input.AccessContextID <= 0 || input.GroupID <= 0 {
		return 0, errors.New("all identifiers should be positive integers")
	}
	if len(input.Data) == 0 {
		return 0, errors.New("document's body should be non-empty")
	}

	var dsid int64
	var docPath model.DocPath
	if input.ParentID > 0 {
		pdoc, err := s.GetDocument(ctx, input.ParentType, input.ParentID)
		if err != nil {
			return 0, err
		}
		docPath = pdoc.Path
		docPath.Append(input.ParentType, input.ParentID)

		// Child document does not have its own state.
		dsid = 1 // `__RESERVED_CHILD_STATE__`
	} else {
		wfState, err := s.GetWorkflowInitialState(ctx, input.DocTypeID)
		if err != nil {
			return 0, err
		}
		dsid = int64(wfState)
	}

	q := `INSERT INTO wf_documents(doctype_id, path, ac_id, docstate_id, group_id, ctime, title, data)
	VALUES (?, ?, ?, ?, ?, NOW(), ?, ?)
	`
	res, err := s.db.ExecContext(ctx, q, input.DocTypeID, string(docPath), input.AccessContextID, dsid, input.GroupID, input.Title, input.Data)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if input.ParentID > 0 {
		q2 := `
		INSERT INTO wf_document_children(parent_doctype_id, parent_id, child_doctype_id, child_id)
		VALUES (?, ?, ?, ?)
		`
		_, err = s.db.ExecContext(ctx, q2, input.ParentType, input.ParentID, input.DocTypeID, id)
		if err != nil {
			return 0, err
		}
	}

	return model.DocumentID(id), nil
}

func (s *MySQLStore) GetWorkflowInitialState(ctx context.Context, dtype model.DocTypeID) (model.DocStateID, error) {
	q := `
	SELECT docstate_id
	FROM wf_workflows
	WHERE doctype_id = ?
	AND active = 1
	`
	row := s.db.QueryRowContext(ctx, q, dtype)
	var dsid int64
	err := row.Scan(&dsid)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("no active workflow is defined for the given document type")
		}
		return 0, err
	}

	return model.DocStateID(dsid), nil
}

func (s *MySQLStore) ListDocuments(ctx context.Context, input *model.DocumentsListInput, offset, limit int64) ([]*model.Document, error) {
	if offset < 0 || limit < 0 {
		return nil, errors.New("offset and limit must be non-negative integers")
	}
	if limit == 0 {
		limit = math.MaxInt64
	}

	// Base query.
	q := `
	SELECT docs.id, docs.path, docs.ac_id, docs.group_id, gm.name, docs.docstate_id, dsm.name, docs.ctime, docs.title
	FROM wf_documents docs
	JOIN wf_groups_master gm ON gm.id = docs.group_id
	JOIN wf_docstates_master dsm ON dsm.id = docs.docstate_id
	`

	// Process input specification.
	where := []string{}
	args := []interface{}{input.DocTypeID, input.AccessContextID}
	q += `WHERE docs.doctype_id = ?
	AND docs.ac_id = ?
	`

	if input.GroupID > 0 {
		where = append(where, `docs.group_id = ?`)
		args = append(args, input.GroupID)
	}

	if input.DocStateID > 0 {
		where = append(where, `docs.docstate_id = ?`)
		args = append(args, input.DocStateID)
	}

	if !input.CtimeStarting.IsZero() {
		where = append(where, `docs.ctime >= ?`)
		args = append(args, input.CtimeStarting)
	}

	if !input.CtimeBefore.IsZero() {
		where = append(where, `docs.ctime < ?`)
		args = append(args, input.CtimeBefore)
	}

	if input.TitleContains != "" {
		where = append(where, `docs.title LIKE ?`)
		args = append(args, "%"+input.TitleContains+"%")
	}

	if input.RootOnly {
		where = append(where, `docs.path = ''`)
	}

	if len(where) > 0 {
		q += ` AND ` + strings.Join(where, ` AND `)
	}

	q += `
	ORDER BY docs.id
	LIMIT ? OFFSET ?
	`
	args = append(args, limit, offset)

	// Fetch document data.
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ary := make([]*model.Document, 0, 10)
	for rows.Next() {
		var elem model.Document
		var title sql.NullString
		err = rows.Scan(&elem.ID, &elem.Path, &elem.AccCtx.ID, &elem.Group.ID, &elem.Group.Name, &elem.State.ID, &elem.State.Name, &elem.Ctime, &title)
		if err != nil {
			return nil, err
		}

		elem.DocType.ID = input.DocTypeID
		q2 := `SELECT name FROM wf_doctypes_master WHERE id = ?`
		row2 := s.db.QueryRowContext(ctx, q2, input.DocTypeID)
		err = row2.Scan(&elem.DocType.Name)
		if err != nil {
			return nil, err
		}

		if title.Valid {
			elem.Title = title.String
		}
		ary = append(ary, &elem)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ary, nil
}

func (s *MySQLStore) GetDocument(ctx context.Context, dtype model.DocTypeID, id model.DocumentID) (*model.Document, error) {
	var elem model.Document
	q := `
	SELECT docs.path, docs.ac_id, docs.group_id, gm.name, docs.ctime, docs.title, docs.data, docs.docstate_id, dsm.name
	FROM wf_documents AS docs
	JOIN wf_groups_master gm ON gm.id = docs.group_id
	JOIN wf_docstates_master dsm ON docs.docstate_id = dsm.id
	WHERE docs.doctype_id = ?
	AND docs.id = ?
	`
	row := s.db.QueryRowContext(ctx, q, dtype, id)
	err := row.Scan(&elem.Path, &elem.AccCtx.ID, &elem.Group.ID, &elem.Group.Name, &elem.Ctime, &elem.Title, &elem.Data, &elem.State.ID, &elem.State.Name)
	if err != nil {
		return nil, err
	}

	q = `SELECT name FROM wf_doctypes_master WHERE id = ?`
	row = s.db.QueryRowContext(ctx, q, dtype)
	err = row.Scan(&elem.DocType.Name)
	if err != nil {
		return nil, err
	}

	elem.ID = id
	elem.DocType.ID = dtype
	return &elem, nil
}

func (s *MySQLStore) GetDocumentParent(ctx context.Context, dtype model.DocTypeID, id model.DocumentID) (*model.Document, error) {
	q := `
	SELECT parent_doctype_id, parent_id
	FROM wf_document_children
	WHERE child_doctype_id = ?
	AND child_id = ?
	LIMIT 1
	`
	row := s.db.QueryRowContext(ctx, q, dtype, id)
	var ptid, pid int64
	err := row.Scan(&ptid, &pid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrDocumentNoParent
		}
		return nil, err
	}

	return s.GetDocument(ctx, model.DocTypeID(ptid), model.DocumentID(pid))
}

func (s *MySQLStore) SetDocumentState(ctx context.Context, dtype model.DocTypeID, id model.DocumentID, state model.DocStateID, ac model.AccessContextID) error {
	var q string
	var err error
	if ac > 0 {
		q = `UPDATE wf_documents SET docstate_id = ?, ac_id = ? WHERE doctype_id = ? AND id = ?`
		_, err = s.db.ExecContext(ctx, q, state, ac, dtype, id)
	} else {
		q = `UPDATE wf_documents SET docstate_id = ? WHERE doctype_id = ? AND id = ?`
		_, err = s.db.ExecContext(ctx, q, state, dtype, id)
	}
	return err
}

func (s *MySQLStore) SetDocumentTitle(ctx context.Context, dtype model.DocTypeID, id model.DocumentID, title string) error {
	if title == "" {
		return errors.New("document title should not be empty")
	}

	// A child document does not have its own title.
	var docPath model.DocPath
	var dgroup model.GroupID
	q := `SELECT path, group_id FROM wf_documents WHERE doctype_id = ? AND id = ?`
	row := s.db.QueryRowContext(ctx, q, dtype, id)
	err := row.Scan(&docPath, &dgroup)
	if err != nil {
		return err
	}
	if docPath != "" {
		return errors.New("a child document cannot have its own title")
	}

	q = `UPDATE wf_documents SET title = ?, ctime = NOW() WHERE doctype_id = ? AND id = ?`
	_, err = s.db.ExecContext(ctx, q, title, dtype, id)
	return err
}

func (s *MySQLStore) SetDocumentData(ctx context.Context, dtype model.DocTypeID, id model.DocumentID, data string) error {
	if data == "" {
		return errors.New("document data should not be empty")
	}

	q := `UPDATE wf_documents SET data = ?, ctime = NOW() WHERE doctype_id = ? AND id = ?`
	_, err := s.db.ExecContext(ctx, q, data, dtype, id)
	return err
}

func (s *MySQLStore) ListDocumentBlobs(ctx context.Context, dtype model.DocTypeID, id model.DocumentID) ([]*model.Blob, error) {
	bs := make([]*model.Blob, 0, 1)
	q := `
	SELECT name, sha1sum
	FROM wf_document_blobs
	WHERE doctype_id = ?
	AND doc_id = ?
	`
	rows, err := s.db.QueryContext(ctx, q, dtype, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var b model.Blob
		err = rows.Scan(&b.Name, &b.SHA1Sum)
		if err != nil {
			return nil, err
		}
		bs = append(bs, &b)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return bs, nil
}

func (s *MySQLStore) GetDocumentBlob(ctx context.Context, dtype model.DocTypeID, id model.DocumentID, blob *model.Blob) error {
	if blob == nil {
		return errors.New("blob should be non-nil")
	}

	q := `
	SELECT name, path
	FROM wf_document_blobs
	WHERE doctype_id = ?
	AND doc_id = ?
	AND sha1sum = ?
	`
	row := s.db.QueryRowContext(ctx, q, dtype, id, blob.SHA1Sum)
	var b model.Blob
	err := row.Scan(&b.Name, &b.Path)
	if err != nil {
		return err
	}
	b.SHA1Sum = blob.SHA1Sum

	// Copy the blob into the destination path given.
	inf, err := os.Open(b.Path)
	if err != nil {
		return err
	}
	defer inf.Close()
	outf, err := os.Create(blob.Path)
	if err != nil {
		return err
	}
	defer outf.Close()
	_, err = io.Copy(outf, inf)
	if err != nil {
		return err
	}

	return nil
}

func (s *MySQLStore) AddDocumentBlob(ctx context.Context, dtype model.DocTypeID, id model.DocumentID, blob *model.Blob) error {
	if blob == nil {
		return errors.New("blob should be non-nil")
	}

	// Verify the given checksum.
	f, err := os.Open(blob.Path)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha1.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return err
	}
	csum := fmt.Sprintf("%x", h.Sum(nil))
	if blob.SHA1Sum != csum {
		return fmt.Errorf("checksum mismatch -- given SHA1 sum : %s, computed SHA1 sum : %s", blob.SHA1Sum, csum)
	}

	// Store the blob in the appropriate path.
	success := false
	bpath := path.Join(s.blobsDir, csum[0:2], csum)
	err = os.Rename(blob.Path, bpath)
	if err != nil {
		return err
	}
	defer func() {
		if !success {
			os.Remove(bpath)
		}
	}()

	// Now write the database entry.
	q := `
	INSERT INTO wf_document_blobs(doctype_id, doc_id, name, path, sha1sum)
	VALUES(?, ?, ?, ?, ?)
	`
	_, err = s.db.ExecContext(ctx, q, dtype, id, blob.Name, bpath, csum)
	if err != nil {
		return err
	}

	success = true
	return nil
}

func (s *MySQLStore) DeleteDocumentBlob(ctx context.Context, dtype model.DocTypeID, id model.DocumentID, sha1sum string) error {
	if sha1sum == "" {
		return errors.New("SHA1 sum should be non-empty")
	}

	q := `
	SELECT COUNT(*)
	FROM wf_document_blobs
	WHERE sha1sum = ?
	`
	var count int64
	row := s.db.QueryRowContext(ctx, q, sha1sum)
	err := row.Scan(&count)
	if err != nil {
		return err
	}
	if count == 1 {
		q = `
		SELECT path
		FROM wf_document_blobs
		WHERE doctype_id = ?
		AND doc_id = ?
		AND sha1sum = ?
		`
		var blobPath string
		row = s.db.QueryRowContext(ctx, q, dtype, id, sha1sum)
		err = row.Scan(&blobPath)
		if err != nil {
			return err
		}

		err = os.Remove(blobPath)
		if err != nil {
			return err
		}
	}

	q = `
	DELETE FROM wf_document_blobs
	WHERE doctype_id = ?
	AND doc_id = ?
	AND sha1sum = ?
	`
	_, err = s.db.ExecContext(ctx, q, dtype, id, sha1sum)
	return err
}

func (s *MySQLStore) ListDocumentTags(ctx context.Context, dtype model.DocTypeID, id model.DocumentID) ([]string, error) {
	ts := make([]string, 0, 1)
	q := `
	SELECT tag
	FROM wf_document_tags
	WHERE doctype_id = ?
	AND doc_id = ?
	`
	rows, err := s.db.QueryContext(ctx, q, dtype, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t string
		err = rows.Scan(&t)
		if err != nil {
			return nil, err
		}
		ts = append(ts, t)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return ts, nil
}

func (s *MySQLStore) AddDocumentTags(ctx context.Context, dtype model.DocTypeID, id model.DocumentID, tags ...string) error {
	// A child document does not have its own tags.
	q := `
	SELECT parent_id
	FROM wf_document_children
	WHERE child_doctype_id = ?
	AND child_id = ?
	ORDER BY child_id
	LIMIT 1
	`
	var tid int64
	row := s.db.QueryRowContext(ctx, q, dtype, id)
	err := row.Scan(&tid)
	if err == nil {
		return model.ErrDocumentIsChild
	}
	if err != sql.ErrNoRows {
		return err
	}

	// Now write the database entries.
	q = `
	INSERT INTO wf_document_tags(doctype_id, doc_id, tag)
	VALUES(?, ?, ?)
	`
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		tag = strings.ToLower(tag)
		_, err = s.db.ExecContext(ctx, q, dtype, id, tag)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *MySQLStore) RemoveDocumentTag(ctx context.Context, dtype model.DocTypeID, id model.DocumentID, tag string) error {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return errors.New("tag should not be empty")
	}
	tag = strings.ToLower(tag)

	q := `
	DELETE FROM wf_document_tags
	WHERE doctype_id = ?
	AND doc_id = ?
	AND tag = ?
	`
	_, err := s.db.ExecContext(ctx, q, dtype, id, tag)
	return err
}

func (s *MySQLStore) ListDocumentChildren(ctx context.Context, dtype model.DocTypeID, id model.DocumentID) ([]struct {
	DocTypeID  model.DocTypeID
	DocumentID model.DocumentID
}, error) {
	cids := make([]struct {
		DocTypeID  model.DocTypeID
		DocumentID model.DocumentID
	}, 0, 1)

	q := `
	SELECT child_doctype_id, child_id
	FROM wf_document_children
	WHERE parent_doctype_id = ?
	AND parent_id = ?
	`
	rows, err := s.db.QueryContext(ctx, q, dtype, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var elem struct {
			DocTypeID  model.DocTypeID
			DocumentID model.DocumentID
		}
		err = rows.Scan(&elem.DocTypeID, &elem.DocumentID)
		if err != nil {
			return nil, err
		}
		cids = append(cids, elem)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return cids, nil
}
