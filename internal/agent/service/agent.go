package agent

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ArtemiySps/calc_go_final/internal/agent/config"
	"github.com/ArtemiySps/calc_go_final/pkg/models"
	"go.uber.org/zap"

	pb "github.com/ArtemiySps/calc_go_final/proto"
	"google.golang.org/grpc"
)

type Agent struct {
	Config *config.Config
	log    *zap.Logger

	resp AgResponse

	pb.CalcServiceServer
}

type AgResponse struct {
	result float32
	err    error
}

func NewAgent(cfg *config.Config, logger *zap.Logger) *Agent {
	return &Agent{
		Config: cfg,
		log:    logger,
	}
}

func (a *Agent) Calculation(ctx context.Context, in *pb.TaskRequest) (*pb.ResResponse, error) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	resultChan := make(chan float32, 1)
	errorChan := make(chan error, 1)

	a.log.Info("got task. calculating...")

	wg.Add(a.Config.ComputingPower)
	go func() {
		defer wg.Done()

		for range a.Config.ComputingPower {
			go a.Worker(ctx, in, resultChan, errorChan)
		}
	}()

	select {
	case res := <-resultChan:
		cancel()
		a.resp.result = res
	case err := <-errorChan:
		cancel()
		a.resp.err = err
	}
	wg.Wait()

	a.log.Info("calculating completed! waiting for the next task...")

	return &pb.ResResponse{Res: float32(a.resp.result)}, a.resp.err
}

// воркер
func (a *Agent) Worker(ctx context.Context, in *pb.TaskRequest, resultChan chan<- float32, errorChan chan<- error) {
	time.Sleep(time.Duration(a.Config.OperationTimes[in.Opr]) * time.Millisecond)

	select {
	case <-ctx.Done():
		return
	default:
		switch in.Opr {
		case "+":
			resultChan <- in.Arg1 + in.Arg2
		case "-":
			resultChan <- in.Arg1 - in.Arg2
		case "*":
			resultChan <- in.Arg1 * in.Arg2
		case "/":
			if in.Arg2 == 0 {
				errorChan <- models.ErrDivisionByZero
			}
			resultChan <- in.Arg1 / in.Arg2
		default:
			errorChan <- models.ErrUnexpectedSymbol
		}
	}
}

func RunServer(cfg *config.Config, logger *zap.Logger) error {
	logger.Info("Server (agent) is starting on port: " + cfg.AgentPort + " and host: " + cfg.AgentHost)

	addr := fmt.Sprintf("%s:%s", cfg.AgentHost, cfg.AgentPort)
	lis, err := net.Listen("tcp", addr)

	if err != nil {
		logger.Info("")
		return models.ErrStartingListener
	}

	grpcServer := grpc.NewServer()

	calcServiceAgent := NewAgent(cfg, logger)

	pb.RegisterCalcServiceServer(grpcServer, calcServiceAgent)

	if err := grpcServer.Serve(lis); err != nil {
		return models.ErrServingGRPC
	}
	return nil
}

// curl -X POST http://localhost:8081/api/v1/calculate -H "Content-Type:application/json" -d "{\"expression\":\"2+2*3+(3-(3+4)*2-4)+2\"}"
// curl -X POST http://localhost:8080/internal/calculator
