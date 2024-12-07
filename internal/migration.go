package internal

type Version struct {
	Id       int
	Up, Down *Migration
}

type Migration struct {
	UseTx bool
	SQL   string
}
