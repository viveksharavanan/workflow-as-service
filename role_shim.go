package flow

import (
	"context"
	"database/sql"

	mysqlstore "github.com/viveksharavanan/workflow-as-service/internal/store/mysql"
)

type _Roles struct{}

// Roles is the singleton entry point for role operations.
var Roles _Roles

func (_Roles) New(tx *sql.Tx, name string) (RoleID, error) {
	txs := mysqlstore.NewMySQLStoreFromTx(tx)
	return txs.CreateRole(context.Background(), name)
}

func (_Roles) List(offset, limit int64) ([]*Role, error) {
	return s.ListRoles(context.Background(), offset, limit)
}

func (_Roles) Get(id RoleID) (*Role, error) {
	return s.GetRole(context.Background(), id)
}

func (_Roles) GetByName(name string) (*Role, error) {
	return s.GetRoleByName(context.Background(), name)
}

func (_Roles) Rename(tx *sql.Tx, id RoleID, name string) error {
	txs := mysqlstore.NewMySQLStoreFromTx(tx)
	return txs.RenameRole(context.Background(), id, name)
}

func (_Roles) Delete(tx *sql.Tx, id RoleID) error {
	txs := mysqlstore.NewMySQLStoreFromTx(tx)
	return txs.DeleteRole(context.Background(), id)
}

func (_Roles) AddPermissions(tx *sql.Tx, rid RoleID, dtype DocTypeID, actions []DocActionID) error {
	txs := mysqlstore.NewMySQLStoreFromTx(tx)
	return txs.AddRolePermissions(context.Background(), rid, dtype, actions)
}

func (_Roles) RemovePermissions(tx *sql.Tx, rid RoleID, dtype DocTypeID, actions []DocActionID) error {
	txs := mysqlstore.NewMySQLStoreFromTx(tx)
	return txs.RemoveRolePermissions(context.Background(), rid, dtype, actions)
}

func (_Roles) Permissions(rid RoleID) (map[string]struct {
	DocTypeID DocTypeID
	Actions   []*DocAction
}, error) {
	return s.ListRolePermissions(context.Background(), rid)
}

func (_Roles) HasPermission(rid RoleID, dtype DocTypeID, action DocActionID) (bool, error) {
	return s.RoleHasPermission(context.Background(), rid, dtype, action)
}
