package models

import (
	"database/sql"
	"fmt"
	"os"

	db2 "github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/mysql"
	"github.com/upper/db/v4/adapter/postgresql"
)

var db *sql.DB
var upper db2.Session

// Models holds references to all our models for the entire application. Add new
// models here when they are created and include them in New.
type Models struct {
	Users         User
	Tokens        Token
	RememberToken RememberToken
}

// New creates a new database pool based on our .env DATABASE_TYPE and returns
// a Model struct that references our Models throughout the application.
func New(databasePool *sql.DB) Models {
	db = databasePool
	switch os.Getenv("DATABASE_TYPE") {
	case "mysql", "mariadb":
		upper, _ = mysql.New(databasePool)

	case "postgres", "postgresql":
		upper, _ = postgresql.New(databasePool)
	default:
		// load no DBs
	}
	return Models{
		Users:         User{},
		Tokens:        Token{},
		RememberToken: RememberToken{},
	}
}

// getInsertID handles how IDs are returned from mysql or postgres type databases with different
// types for the ID
func getInsertID(i db2.ID) int {
	idType := fmt.Sprintf("%T", i)
	if idType == "int64" {
		return int(i.(int64))
	}
	return i.(int)
}
