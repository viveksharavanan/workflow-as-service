package flow

import (
	"context"
	"database/sql"

	mysqlstore "github.com/viveksharavanan/workflow-as-service/internal/store/mysql"
)

type _DocActions struct{}

// DocActions is the singleton entry point for document action operations.
var DocActions _DocActions

func (_DocActions) New(tx *sql.Tx, name string, reconfirm bool) (DocActionID, error) {
	txs := mysqlstore.NewMySQLStoreFromTx(tx)
	return txs.CreateDocAction(context.Background(), name, reconfirm)
}

func (_DocActions) List(offset, limit int64) ([]*DocAction, error) {
	return s.ListDocActions(context.Background(), offset, limit)
}

func (_DocActions) Get(id DocActionID) (*DocAction, error) {
	return s.GetDocAction(context.Background(), id)
}

func (_DocActions) GetByName(name string) (*DocAction, error) {
	return s.GetDocActionByName(context.Background(), name)
}

func (_DocActions) Rename(tx *sql.Tx, id DocActionID, name string) error {
	txs := mysqlstore.NewMySQLStoreFromTx(tx)
	return txs.RenameDocAction(context.Background(), id, name)
}
