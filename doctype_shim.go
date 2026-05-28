package flow

import (
	"context"
	"database/sql"

	mysqlstore "github.com/viveksharavanan/workflow-as-service/internal/store/mysql"
)

type _DocTypes struct{}

// DocTypes is the singleton entry point for document type operations.
var DocTypes _DocTypes

func (_DocTypes) New(tx *sql.Tx, name string) (DocTypeID, error) {
	txs := mysqlstore.NewMySQLStoreFromTx(tx)
	return txs.CreateDocType(context.Background(), name)
}

func (_DocTypes) List(offset, limit int64) ([]*DocType, error) {
	return s.ListDocTypes(context.Background(), offset, limit)
}

func (_DocTypes) Get(id DocTypeID) (*DocType, error) {
	return s.GetDocType(context.Background(), id)
}

func (_DocTypes) GetByName(name string) (*DocType, error) {
	return s.GetDocTypeByName(context.Background(), name)
}

func (_DocTypes) Rename(tx *sql.Tx, id DocTypeID, name string) error {
	txs := mysqlstore.NewMySQLStoreFromTx(tx)
	return txs.RenameDocType(context.Background(), id, name)
}

func (_DocTypes) Transitions(dtype DocTypeID, from DocStateID) (map[DocStateID]*TransitionMap, error) {
	return s.ListTransitions(context.Background(), dtype, from)
}

func (_DocTypes) AddTransition(tx *sql.Tx, dtype DocTypeID, state DocStateID, action DocActionID, toState DocStateID) error {
	txs := mysqlstore.NewMySQLStoreFromTx(tx)
	return txs.AddTransition(context.Background(), dtype, state, action, toState)
}

func (_DocTypes) RemoveTransition(tx *sql.Tx, dtype DocTypeID, state DocStateID, action DocActionID) error {
	txs := mysqlstore.NewMySQLStoreFromTx(tx)
	return txs.RemoveTransition(context.Background(), dtype, state, action)
}
