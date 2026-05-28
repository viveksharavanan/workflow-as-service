package workflow

import (
	"context"
	"errors"
	"log"

	"github.com/viveksharavanan/workflow-as-service/internal/model"
	"github.com/viveksharavanan/workflow-as-service/internal/store"
)

// Engine executes workflow state transitions. It depends on a Store
// interface for all database operations.
type Engine struct {
	store store.Store
}

// NewEngine creates a new workflow engine backed by the given store.
func NewEngine(s store.Store) *Engine {
	return &Engine{store: s}
}

// ApplyEvent takes an input user action or a system event, and applies
// its document action to the given document. This results in a possibly
// new document state. This method also prepares a message that is posted
// to applicable mailboxes.
func (e *Engine) ApplyEvent(ctx context.Context, wf *model.Workflow, event *model.DocEvent, recipients []model.GroupID) (model.DocStateID, error) {
	if !wf.Active {
		return 0, model.ErrWorkflowInactive
	}
	if event.Status == model.EventStatusApplied {
		return 0, model.ErrDocEventAlreadyApplied
	}
	if wf.DocType.ID != event.DocType {
		return 0, model.ErrDocEventDocTypeMismatch
	}

	n, err := e.store.GetNodeByState(ctx, wf.DocType.ID, event.State)
	if err != nil {
		return 0, err
	}

	gt, err := e.store.GetGroupType(ctx, event.Group)
	if err != nil {
		return 0, err
	}
	if gt != "S" {
		return 0, errors.New("group must be singleton")
	}

	var nstate model.DocStateID
	err = e.store.WithTx(ctx, func(txStore store.Store) error {
		var txErr error
		nstate, txErr = e.applyEvent(ctx, txStore, n, event, recipients)
		return txErr
	})
	if err != nil {
		return 0, err
	}

	return nstate, nil
}

// applyEvent checks to see if the given event can be applied
// successfully. Accordingly, it prepares a message by utilising the
// registered node function, and posts it to applicable mailboxes.
func (e *Engine) applyEvent(ctx context.Context, s store.Store, n *model.Node, event *model.DocEvent, recipients []model.GroupID) (model.DocStateID, error) {
	ts, err := s.GetTransitionTargets(ctx, n.DocType, n.State)
	if err != nil {
		return 0, err
	}
	tstate, ok := ts[event.Action]
	if !ok {
		return 0, model.ErrWorkflowInvalidAction
	}

	doc, err := s.GetDocument(ctx, event.DocType, event.DocID)
	if err != nil {
		return 0, err
	}
	if doc.State.ID != event.State {
		return 0, model.ErrDocEventStateMismatch
	}

	if doc.State.ID == tstate {
		err = s.RecordEventApplication(ctx, event.DocType, event.DocID, event.State, event.ID, tstate, true)
		if err != nil {
			return 0, err
		}
		return tstate, model.ErrDocEventRedundant
	}

	tnode, err := s.GetNodeByState(ctx, n.DocType, tstate)
	if err != nil {
		return 0, err
	}

	switch tnode.NodeType {
	case model.NodeTypeJoinAny:
		fallthrough

	case model.NodeTypeBegin, model.NodeTypeEnd, model.NodeTypeLinear, model.NodeTypeBranch:
		tacid := tnode.AccCtx
		if tacid == 0 {
			tacid = doc.AccCtx.ID
		}
		err = s.SetDocumentState(ctx, event.DocType, event.DocID, tstate, tacid)
		if err != nil {
			return 0, err
		}

		err = s.RecordEventApplication(ctx, event.DocType, event.DocID, event.State, event.ID, tstate, false)
		if err != nil {
			return 0, err
		}

		recv := make(map[model.GroupID]struct{})
		for _, gid := range recipients {
			recv[gid] = struct{}{}
		}

		nfunc := n.Func()
		if nfunc == nil {
			nfunc = model.DefNodeFunc
		}
		msg := nfunc(doc, event)

		reportingGroups, err := s.GetReportingGroups(ctx, tacid, event.Group)
		if err != nil {
			return 0, err
		}
		for _, gid := range reportingGroups {
			recv[gid] = struct{}{}
		}

		participants, err := s.GetEventParticipants(ctx, doc.DocType.ID, doc.ID)
		if err != nil {
			return 0, err
		}
		for _, gid := range participants {
			recv[gid] = struct{}{}
		}

		if len(recv) > 0 {
			msgID, err := s.CreateMessage(ctx, msg)
			if err != nil {
				return 0, err
			}
			recvSlice := make([]model.GroupID, 0, len(recv))
			for gid := range recv {
				recvSlice = append(recvSlice, gid)
			}
			err = s.DeliverMessage(ctx, msgID, recvSlice)
			if err != nil {
				return 0, err
			}
		}

	case model.NodeTypeJoinAll:
		// TODO(js)

	default:
		log.Panicf("unknown node type encountered : %s\n", tnode.NodeType)
	}

	return tstate, nil
}
