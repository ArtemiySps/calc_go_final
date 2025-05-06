package service

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/mattn/go-sqlite3"

	"github.com/ArtemiySps/calc_go_final/internal/orkestrator/config"
	"github.com/ArtemiySps/calc_go_final/pkg/models"
	pb "github.com/ArtemiySps/calc_go_final/proto"
)

type TaskStorage struct {
	mu    sync.Mutex
	Tasks map[string]models.Task
}

type Orkestrator struct {
	Config *config.Config
	log    *zap.Logger

	exprs       *sql.DB
	ctx         context.Context
	TaskStorage *TaskStorage

	conn       *grpc.ClientConn
	conn_err   error
	grpcClient pb.CalcServiceClient
}

func NewTaskStorage() *TaskStorage {
	strg := make(map[string]models.Task)
	return &TaskStorage{
		Tasks: strg,
	}
}

func NewOrkestrator(cfg *config.Config, logger *zap.Logger) (*Orkestrator, error) {
	db, err := ExpressionStorageOperations()
	if err != nil {
		return &Orkestrator{}, err
	}

	return &Orkestrator{
		Config:      cfg,
		log:         logger,
		exprs:       db,
		ctx:         context.TODO(),
		TaskStorage: NewTaskStorage(),
	}, nil
}

// создание/открытие БД для выражений
func ExpressionStorageOperations() (*sql.DB, error) {
	_, err := os.Stat("expressions.db")
	if os.IsNotExist(err) {
		ctx := context.TODO()

		db, err := sql.Open("sqlite3", "expressions.db")
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

	db, err := sql.Open("sqlite3", "expressions.db")
	if err != nil {
		return nil, models.ErrDatabaseCreating
	}
	return db, nil
}

// добавляет выражение в БД
func (o *Orkestrator) AddExpressionToStorage(expr string) (string, error) {
	id := models.MakeID()

	var q = `
	INSERT INTO expressions (id, expr, status, result, error) values ($1, $2, $3, $4, $5)
	`
	_, err := o.exprs.ExecContext(o.ctx, q, id, expr, models.StatusPending, 0, "")
	if err != nil {
		return id, err
	}
	return id, nil
}

// получение всех выражений
func (o *Orkestrator) GetAllExpressions() (map[string]models.Expression, error) {
	expressions := make(map[string]models.Expression)
	var q = "SELECT id, expr, status, result, error FROM expressions"

	rows, err := o.exprs.QueryContext(o.ctx, q)
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
func (o *Orkestrator) GetExpression(id string) (models.Expression, error) {
	e := models.Expression{}
	var q = "SELECT id, expr, status, result, error FROM expressions WHERE id = $1"
	err := o.exprs.QueryRowContext(o.ctx, q, id).Scan(&e.ID, &e.Expr, &e.Status, &e.Result, &e.Error)
	if err != nil {
		return models.Expression{}, err
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
		return nil
	}

	var q = "UPDATE expressions SET status = $1, error = $2 WHERE id = $3"
	_, err2 := o.exprs.ExecContext(o.ctx, q, models.StatusFailed, err, id)
	if err2 != nil {
		return err2
	}
	return nil
}

func (o *Orkestrator) ChangeTask(id string, task models.Task) {
	o.TaskStorage.mu.Lock()
	defer o.TaskStorage.mu.Unlock()

	//o.log.Info(id + ": task info changed")

	o.TaskStorage.Tasks[id] = task
	//fmt.Println(o.TaskStorage.Tasks)
}

func (o *Orkestrator) AddTaskToStorage(task models.Task) {
	o.TaskStorage.mu.Lock()
	defer o.TaskStorage.mu.Unlock()

	//o.log.Info(task.ID + ": task added to storage")
	o.TaskStorage.Tasks[task.ID] = task
}

func (o *Orkestrator) DeleteTaskFromStorage(id string) {
	o.TaskStorage.mu.Lock()
	defer o.TaskStorage.mu.Unlock()

	delete(o.TaskStorage.Tasks, id)
}

func (o *Orkestrator) ConnectToServer() error {
	o.log.Info("Client (orkestrator) is connecting to server (agent) with server address: " + o.Config.AgentHost + ":" + o.Config.AgentPort)

	addr := fmt.Sprintf("%s:%s", o.Config.AgentHost, o.Config.AgentPort)

	o.conn, o.conn_err = grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if o.conn_err != nil {
		o.conn_err = nil
		return models.ErrConnectingGRPC
	}

	o.grpcClient = pb.NewCalcServiceClient(o.conn)
	return nil
}

func (o *Orkestrator) ExpressionOperations(expr string) (float64, error) {
	id, err := o.AddExpressionToStorage(expr)
	if err != nil {
		o.log.Info(id + ": failed to add to storage")
		return 0, err
	}
	o.log.Info(id + ": added to storage")

	expr_rpn, err := models.InfixToPostfix(expr)
	if err != nil {
		return 0, err
	}
	o.log.Info(id + ": modified to postfix")

	var stack []float64
	for _, char := range expr_rpn {
		if num, err := strconv.Atoi(string(char)); err == nil {
			stack = append(stack, float64(num))
		} else {
			if len(stack) < 2 {
				o.ChangeExpressionStatus(id, 0, false, models.ErrBadExpression.Error())
				return 0, models.ErrBadExpression
			}

			operand2 := stack[len(stack)-1]
			operand1 := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			task := models.Task{
				ID: models.MakeID(),

				Arg1:      operand1,
				Arg2:      operand2,
				Operation: string(char),
			}

			o.AddTaskToStorage(task)
			//fmt.Println(o.TaskStorage.Tasks)

			resp, err := o.grpcClient.Calculation(context.Background(), &pb.TaskRequest{
				Arg1: float32(operand1),
				Arg2: float32(operand2),
				Opr:  string(char),
			})

			if err != nil {
				task.Error = err.Error()
				o.ChangeTask(task.ID, task)

				o.log.Info(id + ": expression status changed")
				o.ChangeExpressionStatus(id, 0, false, err.Error())
				return 0, err
			}
			//fmt.Println("result:", resp.Res)

			task.Result = float64(resp.Res)
			o.ChangeTask(task.ID, task)

			stack = append(stack, float64(resp.Res))
		}
	}

	o.log.Info(id + ": expression status changed")
	if len(stack) != 1 {
		o.ChangeExpressionStatus(id, 0, false, models.ErrBadExpression.Error())
		return 0, models.ErrBadExpression
	}

	o.ChangeExpressionStatus(id, stack[0], true, "")
	return stack[0], nil
}
