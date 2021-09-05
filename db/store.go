package db

import (
	"github.com/jmoiron/sqlx"
)

type (
	Store struct {
		Db *sqlx.DB
	}
)
