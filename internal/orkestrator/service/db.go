package service

import (
	"context"
	"database/sql"
	"os"

	"github.com/ArtemiySps/calc_go_final/pkg/models"
)

// создание/открытие БД для выражений
func ExpressionStorageOperations() (*sql.DB, error) {
	_, err := os.Stat("expressions.db")
	if os.IsNotExist(err) {
		ctx := context.TODO()

		db, err := sql.Open("sqlite3", "./db/expressions.db")
		if err != nil {
			return nil, err
		}

		err = db.PingContext(ctx)
		if err != nil {
			return nil, err
		}

		const (
			expressionsTable = `
		CREATE TABLE IF NOT EXISTS expressions(
			user TEXT NOT NULL,
			id TEXT NOT NULL,
			expr TEXT NOT NULL,
			status TEXT NOT NULL,
			result REAL,
			error TEXT
		);`
		)

		if _, err := db.ExecContext(ctx, expressionsTable); err != nil {
			return nil, err
		}

		return db, nil
	}

	db, err := sql.Open("sqlite3", "./db/expressions.db")
	if err != nil {
		return nil, models.ErrDatabaseCreating
	}
	return db, nil
}

// добавляет выражение в БД
func (o *Orkestrator) AddExpressionToStorage(expr string, user string) (string, error) {
	id := models.MakeID()

	var q = `
	INSERT INTO expressions (user, id, expr, status, result, error) values ($1, $2, $3, $4, $5, $6)
	`
	_, err := o.exprs.ExecContext(o.ctx, q, user, id, expr, models.StatusPending, 0, "")
	if err != nil {
		return id, err
	}
	return id, nil
}

// получение всех выражений
func (o *Orkestrator) GetAllExpressions(user string) (map[string]models.Expression, error) {
	expressions := make(map[string]models.Expression)
	var q = "SELECT id, expr, status, result, error FROM expressions WHERE user = $1"

	rows, err := o.exprs.QueryContext(o.ctx, q, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		e := models.Expression{}
		err := rows.Scan(&e.ID, &e.Expr, &e.Status, &e.Result, &e.Error)
		if err != nil {
			return nil, err
		}
		expressions[e.ID] = e
	}
	return expressions, nil
}

// получить конкретное выражение по id
func (o *Orkestrator) GetExpression(id string, user string) (models.Expression, error) {
	e := models.Expression{}
	var q = "SELECT id, expr, status, result, error FROM expressions WHERE id = $1 AND user = $2"
	err := o.exprs.QueryRowContext(o.ctx, q, id, user).Scan(&e.ID, &e.Expr, &e.Status, &e.Result, &e.Error)
	if err != nil {
		return models.Expression{}, models.ErrCannotFindObject
	}
	return e, nil
}

// изменить статус и результат/ошибку выражения
func (o *Orkestrator) ChangeExpressionStatus(id string, res float64, ok bool, err string) error {
	if ok {
		var q = "UPDATE expressions SET status = $1, result = $2 WHERE id = $3"
		_, err2 := o.exprs.ExecContext(o.ctx, q, models.StatusCompleted, res, id)
		if err2 != nil {
			return err2
		}
		o.log.Info(id + ": expression status changed: " + models.StatusCompleted)
		return nil
	}

	var q = "UPDATE expressions SET status = $1, error = $2 WHERE id = $3"
	_, err2 := o.exprs.ExecContext(o.ctx, q, models.StatusFailed, err, id)
	if err2 != nil {
		return err2
	}
	o.log.Info(id + ": expression status changed: " + models.StatusFailed)
	return nil
}

func (o *Orkestrator) Clear(user string) (int64, error) {
	q := `DELETE FROM expressions WHERE user = $1`

	result, err := o.exprs.ExecContext(o.ctx, q, user)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}
