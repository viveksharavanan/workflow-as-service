package store

import (
	"context"

	"github.com/viveksharavanan/workflow-as-service/internal/model"
)

// Store defines all database operations for the workflow engine.
type Store interface {
	// WithTx executes fn within a database transaction. If the store
	// already operates inside a transaction, fn runs in that same
	// transaction without nesting.
	WithTx(ctx context.Context, fn func(Store) error) error

	// ----------------------------------------------------------------
	// DocType
	// ----------------------------------------------------------------

	CreateDocType(ctx context.Context, name string) (model.DocTypeID, error)
	ListDocTypes(ctx context.Context, offset, limit int64) ([]*model.DocType, error)
	GetDocType(ctx context.Context, id model.DocTypeID) (*model.DocType, error)
	GetDocTypeByName(ctx context.Context, name string) (*model.DocType, error)
	RenameDocType(ctx context.Context, id model.DocTypeID, name string) error

	// Transitions answers the full transition map for a document type.
	// If from > 0, only transitions from that state are returned.
	ListTransitions(ctx context.Context, dtype model.DocTypeID, from model.DocStateID) (map[model.DocStateID]*model.TransitionMap, error)
	// GetTransitionTargets answers action->target-state for a specific
	// (doctype, state) pair.
	GetTransitionTargets(ctx context.Context, dtype model.DocTypeID, state model.DocStateID) (map[model.DocActionID]model.DocStateID, error)
	AddTransition(ctx context.Context, dtype model.DocTypeID, state model.DocStateID, action model.DocActionID, toState model.DocStateID) error
	RemoveTransition(ctx context.Context, dtype model.DocTypeID, state model.DocStateID, action model.DocActionID) error

	// ----------------------------------------------------------------
	// DocState
	// ----------------------------------------------------------------

	CreateDocState(ctx context.Context, name string) (model.DocStateID, error)
	ListDocStates(ctx context.Context, offset, limit int64) ([]*model.DocState, error)
	GetDocState(ctx context.Context, id model.DocStateID) (*model.DocState, error)
	GetDocStateByName(ctx context.Context, name string) (*model.DocState, error)
	RenameDocState(ctx context.Context, id model.DocStateID, name string) error

	// ----------------------------------------------------------------
	// DocAction
	// ----------------------------------------------------------------

	CreateDocAction(ctx context.Context, name string, reconfirm bool) (model.DocActionID, error)
	ListDocActions(ctx context.Context, offset, limit int64) ([]*model.DocAction, error)
	GetDocAction(ctx context.Context, id model.DocActionID) (*model.DocAction, error)
	GetDocActionByName(ctx context.Context, name string) (*model.DocAction, error)
	RenameDocAction(ctx context.Context, id model.DocActionID, name string) error

	// ----------------------------------------------------------------
	// DocEvent
	// ----------------------------------------------------------------

	CreateDocEvent(ctx context.Context, input *model.DocEventsNewInput) (model.DocEventID, error)
	ListDocEvents(ctx context.Context, input *model.DocEventsListInput, offset, limit int64) ([]*model.DocEvent, error)
	GetDocEvent(ctx context.Context, id model.DocEventID) (*model.DocEvent, error)
	GetDocEventStatus(ctx context.Context, id model.DocEventID) (model.EventStatus, error)
	SetDocEventStatus(ctx context.Context, id model.DocEventID, status string) error

	// RecordEventApplication writes a record of a successful event
	// application and marks the event as applied.
	RecordEventApplication(ctx context.Context, dtype model.DocTypeID, docID model.DocumentID, fromState model.DocStateID, eventID model.DocEventID, toState model.DocStateID, statusOnly bool) error

	// ----------------------------------------------------------------
	// Document
	// ----------------------------------------------------------------

	CreateDocument(ctx context.Context, input *model.DocumentsNewInput) (model.DocumentID, error)
	ListDocuments(ctx context.Context, input *model.DocumentsListInput, offset, limit int64) ([]*model.Document, error)
	GetDocument(ctx context.Context, dtype model.DocTypeID, id model.DocumentID) (*model.Document, error)
	GetDocumentParent(ctx context.Context, dtype model.DocTypeID, id model.DocumentID) (*model.Document, error)
	SetDocumentState(ctx context.Context, dtype model.DocTypeID, id model.DocumentID, state model.DocStateID, ac model.AccessContextID) error
	SetDocumentTitle(ctx context.Context, dtype model.DocTypeID, id model.DocumentID, title string) error
	SetDocumentData(ctx context.Context, dtype model.DocTypeID, id model.DocumentID, data string) error

	ListDocumentBlobs(ctx context.Context, dtype model.DocTypeID, id model.DocumentID) ([]*model.Blob, error)
	GetDocumentBlob(ctx context.Context, dtype model.DocTypeID, id model.DocumentID, blob *model.Blob) error
	AddDocumentBlob(ctx context.Context, dtype model.DocTypeID, id model.DocumentID, blob *model.Blob) error
	DeleteDocumentBlob(ctx context.Context, dtype model.DocTypeID, id model.DocumentID, sha1 string) error

	ListDocumentTags(ctx context.Context, dtype model.DocTypeID, id model.DocumentID) ([]string, error)
	AddDocumentTags(ctx context.Context, dtype model.DocTypeID, id model.DocumentID, tags ...string) error
	RemoveDocumentTag(ctx context.Context, dtype model.DocTypeID, id model.DocumentID, tag string) error

	ListDocumentChildren(ctx context.Context, dtype model.DocTypeID, id model.DocumentID) ([]struct {
		DocTypeID  model.DocTypeID
		DocumentID model.DocumentID
	}, error)

	// GetWorkflowInitialState answers the initial document state for
	// the active workflow of the given document type.
	GetWorkflowInitialState(ctx context.Context, dtype model.DocTypeID) (model.DocStateID, error)

	// ----------------------------------------------------------------
	// Workflow
	// ----------------------------------------------------------------

	CreateWorkflow(ctx context.Context, name string, dtype model.DocTypeID, state model.DocStateID) (model.WorkflowID, error)
	ListWorkflows(ctx context.Context, offset, limit int64) ([]*model.Workflow, error)
	GetWorkflow(ctx context.Context, id model.WorkflowID) (*model.Workflow, error)
	GetWorkflowByDocType(ctx context.Context, dtid model.DocTypeID) (*model.Workflow, error)
	GetWorkflowByName(ctx context.Context, name string) (*model.Workflow, error)
	RenameWorkflow(ctx context.Context, id model.WorkflowID, name string) error
	SetWorkflowActive(ctx context.Context, id model.WorkflowID, active bool) error
	AddWorkflowNode(ctx context.Context, dtype model.DocTypeID, state model.DocStateID, ac model.AccessContextID, wid model.WorkflowID, name string, ntype model.NodeType) (model.NodeID, error)
	RemoveWorkflowNode(ctx context.Context, wid model.WorkflowID, nid model.NodeID) error

	// ----------------------------------------------------------------
	// Node
	// ----------------------------------------------------------------

	ListNodes(ctx context.Context, wid model.WorkflowID) ([]*model.Node, error)
	GetNode(ctx context.Context, id model.NodeID) (*model.Node, error)
	GetNodeByState(ctx context.Context, dtype model.DocTypeID, state model.DocStateID) (*model.Node, error)

	// GetReportingGroups answers the reporting-authority group IDs for
	// the given group in the given access context.
	GetReportingGroups(ctx context.Context, acid model.AccessContextID, gid model.GroupID) ([]model.GroupID, error)
	// GetEventParticipants answers the distinct group IDs of all events
	// on the given document.
	GetEventParticipants(ctx context.Context, dtype model.DocTypeID, docID model.DocumentID) ([]model.GroupID, error)

	// CreateMessage inserts a message record and returns its ID.
	CreateMessage(ctx context.Context, msg *model.Message) (model.MessageID, error)
	// DeliverMessage inserts mailbox entries for each recipient group.
	DeliverMessage(ctx context.Context, msgID model.MessageID, recipients []model.GroupID) error

	// ----------------------------------------------------------------
	// User
	// ----------------------------------------------------------------

	ListUsers(ctx context.Context, prefix string, offset, limit int64) ([]*model.User, error)
	GetUser(ctx context.Context, id model.UserID) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	IsUserActive(ctx context.Context, id model.UserID) (bool, error)
	GetUserGroups(ctx context.Context, uid model.UserID) ([]*model.Group, error)
	GetUserSingletonGroup(ctx context.Context, uid model.UserID) (*model.Group, error)

	// ----------------------------------------------------------------
	// Group
	// ----------------------------------------------------------------

	CreateGroup(ctx context.Context, name string, gtype string) (model.GroupID, error)
	CreateSingletonGroup(ctx context.Context, uid model.UserID) (model.GroupID, error)
	ListGroups(ctx context.Context, offset, limit int64) ([]*model.Group, error)
	GetGroup(ctx context.Context, id model.GroupID) (*model.Group, error)
	GetGroupType(ctx context.Context, id model.GroupID) (string, error)
	RenameGroup(ctx context.Context, id model.GroupID, name string) error
	DeleteGroup(ctx context.Context, id model.GroupID) error
	ListGroupUsers(ctx context.Context, gid model.GroupID) ([]*model.User, error)
	GroupHasUser(ctx context.Context, gid model.GroupID, uid model.UserID) (bool, error)
	GetSingletonUser(ctx context.Context, gid model.GroupID) (*model.User, error)
	AddGroupUser(ctx context.Context, gid model.GroupID, uid model.UserID) error
	RemoveGroupUser(ctx context.Context, gid model.GroupID, uid model.UserID) error

	// ----------------------------------------------------------------
	// Role
	// ----------------------------------------------------------------

	CreateRole(ctx context.Context, name string) (model.RoleID, error)
	ListRoles(ctx context.Context, offset, limit int64) ([]*model.Role, error)
	GetRole(ctx context.Context, id model.RoleID) (*model.Role, error)
	GetRoleByName(ctx context.Context, name string) (*model.Role, error)
	RenameRole(ctx context.Context, id model.RoleID, name string) error
	DeleteRole(ctx context.Context, id model.RoleID) error
	AddRolePermissions(ctx context.Context, rid model.RoleID, dtype model.DocTypeID, actions []model.DocActionID) error
	RemoveRolePermissions(ctx context.Context, rid model.RoleID, dtype model.DocTypeID, actions []model.DocActionID) error
	ListRolePermissions(ctx context.Context, rid model.RoleID) (map[string]struct {
		DocTypeID model.DocTypeID
		Actions   []*model.DocAction
	}, error)
	RoleHasPermission(ctx context.Context, rid model.RoleID, dtype model.DocTypeID, action model.DocActionID) (bool, error)

	// ----------------------------------------------------------------
	// AccessContext
	// ----------------------------------------------------------------

	CreateAccessContext(ctx context.Context, name string) (model.AccessContextID, error)
	ListAccessContexts(ctx context.Context, prefix string, offset, limit int64) ([]*model.AccessContext, error)
	ListAccessContextsByGroup(ctx context.Context, gid model.GroupID, offset, limit int64) ([]*model.AccessContext, error)
	ListAccessContextsByUser(ctx context.Context, uid model.UserID, offset, limit int64) ([]*model.AccessContext, error)
	GetAccessContext(ctx context.Context, id model.AccessContextID) (*model.AccessContext, error)
	RenameAccessContext(ctx context.Context, id model.AccessContextID, name string) error
	SetAccessContextActive(ctx context.Context, id model.AccessContextID, active bool) error

	GetAccessContextGroupRoles(ctx context.Context, id model.AccessContextID, gids []model.GroupID, offset, limit int64) (map[model.GroupID]*model.AcGroupRoles, error)
	AddAccessContextGroupRole(ctx context.Context, id model.AccessContextID, gid model.GroupID, rid model.RoleID) error
	RemoveAccessContextGroupRole(ctx context.Context, id model.AccessContextID, gid model.GroupID, rid model.RoleID) error

	ListAccessContextGroups(ctx context.Context, id model.AccessContextID, offset, limit int64) (map[model.GroupID]*model.AcGroup, error)
	AddAccessContextGroup(ctx context.Context, id model.AccessContextID, gid, reportsTo model.GroupID) error
	DeleteAccessContextGroup(ctx context.Context, id model.AccessContextID, gid model.GroupID) error
	GetAccessContextGroupReportsTo(ctx context.Context, id model.AccessContextID, gid model.GroupID) (model.GroupID, error)
	ListAccessContextGroupReportees(ctx context.Context, id model.AccessContextID, gid model.GroupID) ([]model.GroupID, error)
	ChangeAccessContextReporting(ctx context.Context, id model.AccessContextID, gid, reportsTo model.GroupID) error
	AccessContextIncludesGroup(ctx context.Context, id model.AccessContextID, gid model.GroupID) (bool, error)
	AccessContextIncludesUser(ctx context.Context, id model.AccessContextID, uid model.UserID) (bool, error)

	GetUserPermissions(ctx context.Context, id model.AccessContextID, uid model.UserID) (map[model.DocTypeID][]model.DocAction, error)
	GetUserPermissionsByDocType(ctx context.Context, id model.AccessContextID, dtype model.DocTypeID, uid model.UserID) ([]model.DocAction, error)
	GetGroupPermissions(ctx context.Context, id model.AccessContextID, gid model.GroupID) (map[model.DocTypeID][]model.DocAction, error)
	GetGroupPermissionsByDocType(ctx context.Context, id model.AccessContextID, dtype model.DocTypeID, gid model.GroupID) ([]model.DocAction, error)
	UserHasPermission(ctx context.Context, id model.AccessContextID, uid model.UserID, dtype model.DocTypeID, action model.DocActionID) (bool, error)
	GroupHasPermission(ctx context.Context, id model.AccessContextID, gid model.GroupID, dtype model.DocTypeID, action model.DocActionID) (bool, error)

	// ----------------------------------------------------------------
	// Mailbox
	// ----------------------------------------------------------------

	CountMailboxByUser(ctx context.Context, uid model.UserID, unread bool) (int64, error)
	CountMailboxByGroup(ctx context.Context, gid model.GroupID, unread bool) (int64, error)
	ListMailboxByUser(ctx context.Context, uid model.UserID, offset, limit int64, unread bool) ([]*model.Notification, error)
	ListMailboxByGroup(ctx context.Context, gid model.GroupID, offset, limit int64, unread bool) ([]*model.Notification, error)
	GetMailboxMessage(ctx context.Context, msgID model.MessageID) (*model.Notification, error)
	ReassignMessage(ctx context.Context, fromGID, toGID model.GroupID, msgID model.MessageID) error
	SetMailboxStatusByUser(ctx context.Context, uid model.UserID, msgID model.MessageID, unread bool) error
	SetMailboxStatusByGroup(ctx context.Context, gid model.GroupID, msgID model.MessageID, unread bool) error
}
