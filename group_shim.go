package flow

import (
	"context"
	"database/sql"

	mysqlstore "github.com/viveksharavanan/workflow-as-service/internal/store/mysql"
)

type _Groups struct{}

// Groups is the singleton entry point for group operations.
var Groups _Groups

func (_Groups) NewSingleton(tx *sql.Tx, uid UserID) (GroupID, error) {
	txs := mysqlstore.NewMySQLStoreFromTx(tx)
	return txs.CreateSingletonGroup(context.Background(), uid)
}

func (_Groups) New(tx *sql.Tx, name string, gtype string) (GroupID, error) {
	txs := mysqlstore.NewMySQLStoreFromTx(tx)
	return txs.CreateGroup(context.Background(), name, gtype)
}

func (_Groups) List(offset, limit int64) ([]*Group, error) {
	return s.ListGroups(context.Background(), offset, limit)
}

func (_Groups) Get(id GroupID) (*Group, error) {
	return s.GetGroup(context.Background(), id)
}

func (_Groups) Rename(tx *sql.Tx, id GroupID, name string) error {
	txs := mysqlstore.NewMySQLStoreFromTx(tx)
	return txs.RenameGroup(context.Background(), id, name)
}

func (_Groups) Delete(tx *sql.Tx, id GroupID) error {
	txs := mysqlstore.NewMySQLStoreFromTx(tx)
	return txs.DeleteGroup(context.Background(), id)
}

func (_Groups) Users(gid GroupID) ([]*User, error) {
	return s.ListGroupUsers(context.Background(), gid)
}

func (_Groups) HasUser(gid GroupID, uid UserID) (bool, error) {
	return s.GroupHasUser(context.Background(), gid, uid)
}

func (_Groups) SingletonUser(gid GroupID) (*User, error) {
	return s.GetSingletonUser(context.Background(), gid)
}

func (_Groups) AddUser(tx *sql.Tx, gid GroupID, uid UserID) error {
	txs := mysqlstore.NewMySQLStoreFromTx(tx)
	return txs.AddGroupUser(context.Background(), gid, uid)
}

func (_Groups) RemoveUser(tx *sql.Tx, gid GroupID, uid UserID) error {
	txs := mysqlstore.NewMySQLStoreFromTx(tx)
	return txs.RemoveGroupUser(context.Background(), gid, uid)
}
