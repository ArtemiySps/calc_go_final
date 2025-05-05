package service

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ArtemiySps/calc_go_final/internal/orkestrator/config"
	"github.com/ArtemiySps/calc_go_final/pkg/models"
	pb "github.com/ArtemiySps/calc_go_final/proto"
)

type TaskStorage struct {
	mu    sync.Mutex
	Tasks map[string]models.Task
}

type ExpressionStorage struct {
	mu          sync.Mutex
	Expressions map[string]models.Expression
}

type Orkestrator struct {
	Config *config.Config
	log    *zap.Logger

	ExpressionStorage *ExpressionStorage
	TaskStorage       *TaskStorage

	conn       *grpc.ClientConn
	conn_err   error
	grpcClient pb.CalcServiceClient
}

func NewExpressionStorage() *ExpressionStorage {
	strg := make(map[string]models.Expression)
	return &ExpressionStorage{
		Expressions: strg,
	}
}

func NewTaskStorage() *TaskStorage {
	strg := make(map[string]models.Task)
	return &TaskStorage{
		Tasks: strg,
	}
}

func NewOrkestrator(cfg *config.Config, logger *zap.Logger) *Orkestrator {
	return &Orkestrator{
		Config:            cfg,
		log:               logger,
		ExpressionStorage: NewExpressionStorage(),
		TaskStorage:       NewTaskStorage(),
	}
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

/*func (o *Orkestrator) GetOperationTime(op rune) time.Duration {
	switch op {
	case '+':
		return o.Config.AdditionTime
	case '-':
		return o.Config.SubtractionTime
	case '*':
		return o.Config.MultiplicationTime
	case '/':
		return o.Config.DivisionTime
	default:
		return 0
	}
}*/

func (o *Orkestrator) AddExpressionToStorage() string {
	expression := models.Expression{
		ID:     models.MakeID(),
		Status: models.StatusPending,
		Result: 0,
	}

	o.ExpressionStorage.Expressions[expression.ID] = expression

	return expression.ID
}

func (o *Orkestrator) GetAllExpressions() map[string]models.Expression {
	o.ExpressionStorage.mu.Lock()
	defer o.ExpressionStorage.mu.Unlock()
	expressions := o.ExpressionStorage.Expressions
	return expressions
}

func (o *Orkestrator) GetExpression(id string) models.Expression {
	o.ExpressionStorage.mu.Lock()
	defer o.ExpressionStorage.mu.Unlock()
	expression := o.ExpressionStorage.Expressions[id]
	return expression
}

func (o *Orkestrator) ChangeExpressionStatus(id string, res float64, ok bool, err string) {
	o.ExpressionStorage.mu.Lock()
	defer o.ExpressionStorage.mu.Unlock()
	if ok {
		if entry, okk := o.ExpressionStorage.Expressions[id]; okk {
			entry.Status = models.StatusCompleted
			entry.Result = res
			o.ExpressionStorage.Expressions[id] = entry
			return
		}
	}
	if entry, okk := o.ExpressionStorage.Expressions[id]; okk {
		entry.Status = models.StatusFailed
		entry.Error = err
		o.ExpressionStorage.Expressions[id] = entry
	}
}

func (o *Orkestrator) ChangeTask(id string, task models.Task) {
	o.TaskStorage.mu.Lock()
	defer o.TaskStorage.mu.Unlock()

	o.log.Info(id + ": task info changed")

	o.TaskStorage.Tasks[id] = task
	//fmt.Println(o.TaskStorage.Tasks)
}

/*func (o *Orkestrator) ChangeTaskStatus(id string, status string) {
	o.TaskStorage.mu.Lock()
	defer o.TaskStorage.mu.Unlock()

	if entry, ok := o.TaskStorage.Tasks[id]; ok {
		entry.Status = status
		o.TaskStorage.Tasks[id] = entry
		return
	}
}*/

func (o *Orkestrator) AddTaskToStorage(task models.Task) {
	o.TaskStorage.mu.Lock()
	defer o.TaskStorage.mu.Unlock()

	o.log.Info(task.ID + ": task added to storage")
	o.TaskStorage.Tasks[task.ID] = task
}

func (o *Orkestrator) DeleteTaskFromStorage(id string) {
	o.TaskStorage.mu.Lock()
	defer o.TaskStorage.mu.Unlock()

	delete(o.TaskStorage.Tasks, id)
}

/*func (o *Orkestrator) GiveTask() (models.Task, error) {
	if len(o.TaskStorage.Tasks) > 0 {
		for id, t := range o.TaskStorage.Tasks {
			if t.Status == models.StatusNeedToSend {
				o.TaskStorage.mu.Lock()
				defer o.TaskStorage.mu.Unlock()

				t.Status = models.StatusPending
				o.TaskStorage.Tasks[id] = t

				o.log.Info(id + ": status changed to pending")
				return t, nil
			}
		}
	}
	return models.Task{}, models.ErrNoTasks
}*/

/*func (o *Orkestrator) WaitForResult(id string) (float64, string) {
	ticker := time.NewTicker(time.Duration(1000) * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		o.TaskStorage.mu.Lock()
		o.log.Info("--- check for update of my task")
		for k, v := range o.TaskStorage.Tasks {
			if k == id && (v.Status == models.StatusCompleted || v.Status == models.StatusFailed) {
				o.log.Info("delete from storage")
				o.TaskStorage.mu.Unlock()
				o.DeleteTaskFromStorage(id)
				return v.Result, v.Error
			}
		}
		o.TaskStorage.mu.Unlock()
	}
	return 0, "ticker stoped"
}*/

func (o *Orkestrator) ExpressionOperations(expr string) (float64, error) {
	id := o.AddExpressionToStorage()
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
