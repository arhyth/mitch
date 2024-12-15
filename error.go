package mitch

import "errors"

var (
	ErrUnsetDBURL         = errors.New("no database url or host set")
	ErrUnsetDBName        = errors.New("database name not set")
	ErrUnsetDBPassword    = errors.New("database password not set")
	ErrFileVersionPrefix  = errors.New("migration filename has no version prefix")
	ErrVersionZero        = errors.New("migration version must be greater than zero")
	ErrVersionDiscrepancy = errors.New("migration applied changed to a different version number")
	ErrMultiStatementLine = errors.New("line has multiple SQL statements")
)
