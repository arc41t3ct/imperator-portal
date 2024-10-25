//go:build unit

package models

import (
	"fmt"
	"os"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	db2 "github.com/upper/db/v4"
)

func TestNew(t *testing.T) {
	fakeDB, _, err := sqlmock.New()
	if err != nil {
		t.Error(err)
	}
	defer fakeDB.Close()

	_ = os.Setenv("DATABASE_TYPE", "postgres")
	m := New(fakeDB)
	if fmt.Sprintf("%T", m) != "models.Models" {
		t.Error("wrong type", fmt.Sprintf("%T", m))
	}

	_ = os.Setenv("DATABASE_TYPE", "mysql")
	m = New(fakeDB)
	if fmt.Sprintf("%T", m) != "models.Models" {
		t.Error("wrong type", fmt.Sprintf("%T", m))
	}
}

func TestGetInsertID(t *testing.T) {
	var id db2.ID
	id = int64(1)

	returnedID := getInsertID(id)
	if fmt.Sprintf("%T", returnedID) != "int" {
		t.Error("wrong type", fmt.Sprintf("%T", returnedID))
	}

	id = 1
	returnedID = getInsertID(id)
	if fmt.Sprintf("%T", returnedID) != "int" {
		t.Error("wrong type", fmt.Sprintf("%T", returnedID))
	}
}
