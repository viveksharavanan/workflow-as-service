package flow

import (
	"context"
	"database/sql"

	mysqlstore "github.com/viveksharavanan/workflow-as-service/internal/store/mysql"
)

type _Workflows struct{}

// Workflows is the singleton entry point for workflow operations.
var Workflows _Workflows

func (_Workflows) New(tx *sql.Tx, name string, dtype DocTypeID, state DocStateID) (WorkflowID, error) {
	txs := mysqlstore.NewMySQLStoreFromTx(tx)
	return txs.CreateWorkflow(context.Background(), name, dtype, state)
}

func (_Workflows) List(offset, limit int64) ([]*Workflow, error) {
	return s.ListWorkflows(context.Background(), offset, limit)
}

func (_Workflows) Get(id WorkflowID) (*Workflow, error) {
	return s.GetWorkflow(context.Background(), id)
}

func (_Workflows) GetByDocType(dtid DocTypeID) (*Workflow, error) {
	return s.GetWorkflowByDocType(context.Background(), dtid)
}

func (_Workflows) GetByName(name string) (*Workflow, error) {
	return s.GetWorkflowByName(context.Background(), name)
}

func (_Workflows) Rename(tx *sql.Tx, id WorkflowID, name string) error {
	txs := mysqlstore.NewMySQLStoreFromTx(tx)
	return txs.RenameWorkflow(context.Background(), id, name)
}

func (_Workflows) SetActive(tx *sql.Tx, id WorkflowID, active bool) error {
	txs := mysqlstore.NewMySQLStoreFromTx(tx)
	return txs.SetWorkflowActive(context.Background(), id, active)
}

func (_Workflows) AddNode(tx *sql.Tx, dtype DocTypeID, state DocStateID, ac AccessContextID, wid WorkflowID, name string, ntype NodeType) (NodeID, error) {
	txs := mysqlstore.NewMySQLStoreFromTx(tx)
	return txs.AddWorkflowNode(context.Background(), dtype, state, ac, wid, name, ntype)
}

func (_Workflows) RemoveNode(tx *sql.Tx, wid WorkflowID, nid NodeID) error {
	txs := mysqlstore.NewMySQLStoreFromTx(tx)
	return txs.RemoveWorkflowNode(context.Background(), wid, nid)
}
