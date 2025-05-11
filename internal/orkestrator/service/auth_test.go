package service

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/ArtemiySps/calc_go_final/internal/orkestrator/config"
	"github.com/ArtemiySps/calc_go_final/pkg/models"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func setupAuthTest(t *testing.T) (*Orkestrator, func()) {
	tmpfile, err := os.CreateTemp("", "auth_test-*.db")
	assert.NoError(t, err)
	tmpfile.Close()

	db, err := sql.Open("sqlite3", tmpfile.Name())
	assert.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE users(
			token TEXT UNIQUE,
			login TEXT UNIQUE,
			password TEXT
		);
	`)
	assert.NoError(t, err)

	o := &Orkestrator{
		users: db,
		ctx:   context.Background(),
		log:   zap.NewNop(),
		Config: &config.Config{
			UsersDBPath: tmpfile.Name(),
		},
	}

	cleanup := func() {
		db.Close()
		os.Remove(tmpfile.Name())
	}

	return o, cleanup
}

func TestAuthFlow(t *testing.T) {
	o, cleanup := setupAuthTest(t)
	defer cleanup()

	t.Run("Successful registration", func(t *testing.T) {
		err := o.Register("testuser", "password123")
		assert.NoError(t, err)
	})

	t.Run("Duplicate registration", func(t *testing.T) {
		err := o.Register("testuser", "password123")
		assert.ErrorIs(t, err, models.ErrUserAlreadyExists)
	})

	t.Run("Successful login", func(t *testing.T) {
		token, err := o.Login("testuser", "password123")
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("Wrong password", func(t *testing.T) {
		_, err := o.Login("testuser", "wrongpass")
		assert.ErrorIs(t, err, models.ErrIncorrectPassword)
	})

	t.Run("Non-existent user", func(t *testing.T) {
		_, err := o.Login("nonexistent", "password123")
		assert.ErrorIs(t, err, models.ErrUserNotRegistered)
	})
}
