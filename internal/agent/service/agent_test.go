package agent

import (
	"context"
	"testing"
	"time"

	"github.com/ArtemiySps/calc_go_final/internal/agent/config"
	"github.com/ArtemiySps/calc_go_final/pkg/models"
	pb "github.com/ArtemiySps/calc_go_final/proto"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// тест для Calculation
func TestAgent_Calculation(t *testing.T) {
	tests := []struct {
		name        string
		req         *pb.TaskRequest
		expectedRes float32
		expectedErr error
	}{
		{
			name:        "Addition",
			req:         &pb.TaskRequest{Opr: "+", Arg1: 2, Arg2: 3},
			expectedRes: 5,
			expectedErr: nil,
		},
		{
			name:        "Subtraction",
			req:         &pb.TaskRequest{Opr: "-", Arg1: 5, Arg2: 3},
			expectedRes: 2,
			expectedErr: nil,
		},
		{
			name:        "Division by zero",
			req:         &pb.TaskRequest{Opr: "/", Arg1: 2, Arg2: 0},
			expectedRes: 0,
			expectedErr: models.ErrDivisionByZero,
		},
		{
			name:        "Unexpected symbol",
			req:         &pb.TaskRequest{Opr: "?", Arg1: 1, Arg2: 2},
			expectedRes: 0,
			expectedErr: models.ErrUnexpectedSymbol,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()

			cfg := &config.Config{
				ComputingPower: 1,
				OperationTimes: map[string]int{
					"+": 0,
					"-": 0,
					"*": 0,
					"/": 0,
					"?": 0,
				},
			}

			agent := NewAgent(cfg, logger)
			res, err := agent.Calculation(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.Equal(t, tt.expectedRes, res.Res)
				assert.NoError(t, err)
			}
		})
	}
}

// тесты для Worker
func TestAgent_Worker(t *testing.T) {
	tests := []struct {
		name        string
		req         *pb.TaskRequest
		expectedRes float32
		expectedErr error
	}{
		{
			name:        "Multiplication",
			req:         &pb.TaskRequest{Opr: "*", Arg1: 4, Arg2: 5},
			expectedRes: 20,
			expectedErr: nil,
		},
		{
			name:        "Division",
			req:         &pb.TaskRequest{Opr: "/", Arg1: 10, Arg2: 2},
			expectedRes: 5,
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()

			cfg := &config.Config{
				OperationTimes: map[string]int{
					"*": 0,
					"/": 0,
				},
			}

			agent := NewAgent(cfg, logger)

			resultChan := make(chan float32, 1)
			errorChan := make(chan error, 1)

			agent.Worker(context.Background(), tt.req, resultChan, errorChan)

			select {
			case res := <-resultChan:
				assert.Equal(t, tt.expectedRes, res)
			case err := <-errorChan:
				assert.ErrorIs(t, err, tt.expectedErr)
			case <-time.After(100 * time.Millisecond):
				t.Fatal("Worker timed out")
			}
		})
	}
}

// тесты для RunServer
func TestRunServer_InvalidAddress(t *testing.T) {
	logger := zap.NewNop()

	cfg := &config.Config{
		AgentHost: "invalid_host",
		AgentPort: "invalid_port",
	}

	err := RunServer(cfg, logger)
	assert.Error(t, err)
}
