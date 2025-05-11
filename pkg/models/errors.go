package models

import (
	"errors"
)

var (
	// ошибки в переменных среды
	ErrAdittionTime       = errors.New("environment variable for addition wasn't set correctly")
	ErrSubtracrionTime    = errors.New("environment variable for subtraction wasn't set correctly")
	ErrMultiplicationTime = errors.New("environment variable for multiplication wasn't set correctly")
	ErrDivisionTime       = errors.New("environment variable for division wasn't set correctly")
	ErrTableFormat        = errors.New("environment variable TABLE_FORMAT wasn't set correctly")

	// ошибки в математическом выражении:
	ErrDivisionByZero   = errors.New("division by zero")
	ErrUnexpectedSymbol = errors.New("unexpected symbol")
	ErrBadExpression    = errors.New("incorrect expression")

	// ошибки grpc
	ErrStartingListener = errors.New("error starting tcp listener")
	ErrServingGRPC      = errors.New("error serving grpc")
	ErrConnectingGRPC   = errors.New("could not connect to grpc server")

	// ошибки database
	ErrDatabaseCreating = errors.New("error creating sql database")
	ErrPingContext      = errors.New("database not active")
	ErrCannotFindObject = errors.New("can't find expression")

	// ошибки auth
	ErrIncorrectPassword = errors.New("incorrecct password")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotRegistered = errors.New("user is not registered")
)
