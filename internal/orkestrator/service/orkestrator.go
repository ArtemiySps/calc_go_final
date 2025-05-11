package service

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/mattn/go-sqlite3"

	"github.com/ArtemiySps/calc_go_final/internal/orkestrator/config"
	"github.com/ArtemiySps/calc_go_final/pkg/models"
	pb "github.com/ArtemiySps/calc_go_final/proto"
)

type Orkestrator struct {
	Config *config.Config
	log    *zap.Logger

	users *sql.DB
	exprs *sql.DB
	ctx   context.Context

	conn       *grpc.ClientConn
	conn_err   error
	grpcClient pb.CalcServiceClient
}

func NewOrkestrator(cfg *config.Config, logger *zap.Logger) (*Orkestrator, error) {
	db, err := ExpressionStorageOperations()
	if err != nil {
		return &Orkestrator{}, err
	}

	return &Orkestrator{
		Config: cfg,
		log:    logger,
		exprs:  db,
		ctx:    context.TODO(),
	}, nil
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

func (o *Orkestrator) ExpressionOperations(expr string, user string) (float64, error) {
	id, err := o.AddExpressionToStorage(expr, user)
	if err != nil {
		o.log.Info(id + ": failed to add to storage")
		return 0, err
	}
	o.log.Info(id + ": added to storage")

	expr_rpn, err := models.InfixToPostfix(expr)
	if err != nil {
		return 0, err
	}

	tokens := strings.Split(expr_rpn, " ")
	var stack []float64

	for _, token := range tokens {
		if num, err := strconv.ParseFloat(token, 64); err == nil {
			stack = append(stack, num)
		} else {
			if len(stack) < 2 {
				o.ChangeExpressionStatus(id, 0, false, models.ErrBadExpression.Error())
				return 0, models.ErrBadExpression
			}

			operand2 := stack[len(stack)-1]
			operand1 := stack[len(stack)-2]
			stack = stack[:len(stack)-2]

			resp, err := o.grpcClient.Calculation(context.Background(), &pb.TaskRequest{
				Arg1: float32(operand1),
				Arg2: float32(operand2),
				Opr:  token,
			})

			if err != nil {
				o.ChangeExpressionStatus(id, 0, false, err.Error())
				return 0, err
			}

			stack = append(stack, float64(resp.Res))
		}
	}

	if len(stack) != 1 {
		o.ChangeExpressionStatus(id, 0, false, models.ErrBadExpression.Error())
		return 0, models.ErrBadExpression
	}

	o.ChangeExpressionStatus(id, stack[0], true, "")
	return stack[0], nil
}
