package mitch

import "errors"

var (
	ErrUnsetDBURL      = errors.New("no database url or host set")
	ErrUnsetDBName     = errors.New("database name not set")
	ErrUnsetDBPassword = errors.New("database password not set")
)
