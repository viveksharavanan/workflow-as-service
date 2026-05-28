package flow

import (
	"context"
	"database/sql"

	mysqlstore "github.com/viveksharavanan/workflow-as-service/internal/store/mysql"
)

type _DocStates struct{}

// DocStates is the singleton entry point for document state operations.
var DocStates _DocStates

func (_DocStates) New(tx *sql.Tx, name string) (DocStateID, error) {
	txs := mysqlstore.NewMySQLStoreFromTx(tx)
	return txs.CreateDocState(context.Background(), name)
}

func (_DocStates) List(offset, limit int64) ([]*DocState, error) {
	return s.ListDocStates(context.Background(), offset, limit)
}

func (_DocStates) Get(id DocStateID) (*DocState, error) {
	return s.GetDocState(context.Background(), id)
}

func (_DocStates) GetByName(name string) (*DocState, error) {
	return s.GetDocStateByName(context.Background(), name)
}

func (_DocStates) Rename(tx *sql.Tx, id DocStateID, name string) error {
	txs := mysqlstore.NewMySQLStoreFromTx(tx)
	return txs.RenameDocState(context.Background(), id, name)
}
