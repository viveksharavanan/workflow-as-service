package flow

import (
	"context"
)

type _Users struct{}

// Users is the singleton entry point for user operations.
var Users _Users

func (_Users) List(prefix string, offset, limit int64) ([]*User, error) {
	return s.ListUsers(context.Background(), prefix, offset, limit)
}

func (_Users) Get(id UserID) (*User, error) {
	return s.GetUser(context.Background(), id)
}
