package service

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE expressions(
			user TEXT NOT NULL,
			id TEXT NOT NULL,
			expr TEXT NOT NULL,
			status TEXT NOT NULL,
			result REAL,
			error TEXT
		);
	`)
	assert.NoError(t, err)

	return db
}

func TestAddExpressionToStorage_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	o := &Orkestrator{
		exprs: db,
		ctx:   context.Background(),
	}

	id, err := o.AddExpressionToStorage("2+2", "testuser")
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
}

func TestGetExpression_Integration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.Exec(`
		INSERT INTO expressions (user, id, expr, status, result, error)
		VALUES ('testuser', '123', '2+2', 'completed', 4, '')
	`)
	assert.NoError(t, err)

	o := &Orkestrator{
		exprs: db,
		ctx:   context.Background(),
	}

	expr, err := o.GetExpression("123", "testuser")
	assert.NoError(t, err)
	assert.Equal(t, "2+2", expr.Expr)
	assert.Equal(t, 4.0, expr.Result)
}
