package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/ArtemiySps/calc_go_final/pkg/models"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Token          string
	Login          string
	Password       string
	OriginPassword string
}

func (o *Orkestrator) SetUsersDB(db *sql.DB) {
	o.users = db
}

func comparePassword(password_u1 string, u2 User) error {
	err := compare(u2.Password, password_u1)
	if err != nil {
		return err
	}

	return nil
}

func createTable(ctx context.Context, db *sql.DB) error {
	const usersTable = `
	CREATE TABLE IF NOT EXISTS users(
		token TEXT UNIQUE,
		login TEXT UNIQUE,
		password TEXT
	);`

	if _, err := db.ExecContext(ctx, usersTable); err != nil {
		return err
	}

	return nil
}

func insertUser(ctx context.Context, db *sql.DB, login string, password string) error {
	var q = `
	INSERT INTO users (login, password) values ($1, $2)
	`
	_, err := db.ExecContext(ctx, q, login, password)
	if err != nil {
		return err
	}

	return nil
}

func selectUser(ctx context.Context, db *sql.DB, login string) (User, error) {
	var (
		user User
		err  error
	)

	var q = "SELECT login, password FROM users WHERE login=$1"

	err = db.QueryRowContext(ctx, q, login).Scan(&user.Login, &user.Password)
	return user, err
}

func addToken(ctx context.Context, db *sql.DB, token string, login string) error {
	var q = "UPDATE users SET token=$1 WHERE login=$2"

	_, err := db.ExecContext(ctx, q, token, login)
	if err != nil {
		return err
	}
	return nil
}

func generate(s string) (string, error) {
	saltedBytes := []byte(s)
	hashedBytes, err := bcrypt.GenerateFromPassword(saltedBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	hash := string(hashedBytes[:])
	return hash, nil
}

func compare(hash string, s string) error {
	incoming := []byte(s)
	existing := []byte(hash)
	return bcrypt.CompareHashAndPassword(existing, incoming)
}

// функция для выдачи JWT токенов
func giveToken(login string) (string, error) {
	privateKey := "sosiski_123"
	now := time.Now()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  login,
		"nbf": now.Unix(),
		"exp": now.Add(24 * time.Hour).Unix(),
		"iat": now.Unix(),
	})

	tokenString, err := token.SignedString([]byte(privateKey))
	if err != nil {
		return "", err
	}

	return tokenString, err
}

// зарегистрировать нового пользователя
func (o *Orkestrator) Register(login string, password string) error {
	var err error

	o.users, err = sql.Open("sqlite3", o.Config.UsersDBPath)
	if err != nil {
		return err
	}
	defer o.users.Close()

	err = o.users.PingContext(o.ctx)
	if err != nil {
		return err
	}

	if err = createTable(o.ctx, o.users); err != nil {
		return err
	}

	hashed, err := generate(password)
	if err != nil {
		return err
	}
	user := &User{
		Login:          login,
		Password:       hashed,
		OriginPassword: password,
	}

	err = insertUser(o.ctx, o.users, user.Login, user.Password)
	if err != nil {
		o.log.Info("user already exists")
		return models.ErrUserAlreadyExists
	}
	return nil
}

// функция для входа пользователя
func (o *Orkestrator) Login(login string, password string) (string, error) {
	var err error

	o.users, err = sql.Open("sqlite3", o.Config.UsersDBPath)
	if err != nil {
		return "", err
	}
	defer o.users.Close()

	err = o.users.PingContext(o.ctx)
	if err != nil {
		return "", err
	}

	userFromDB, err := selectUser(o.ctx, o.users, login)
	if err != nil {
		return "", models.ErrUserNotRegistered
	}

	err = comparePassword(password, userFromDB)
	if err != nil {
		return "", models.ErrIncorrectPassword
	}

	token, err := giveToken(login)
	if err != nil {
		return "", err
	}

	err = addToken(o.ctx, o.users, token, login)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (o *Orkestrator) GetLogin(token string) (string, error) {
	var login string
	var err error

	o.users, err = sql.Open("sqlite3", o.Config.UsersDBPath)
	if err != nil {
		return "", err
	}
	defer o.users.Close()

	var q = "SELECT login FROM users WHERE token=$1"

	err = o.users.QueryRowContext(o.ctx, q, token).Scan(&login)
	return login, err
}
