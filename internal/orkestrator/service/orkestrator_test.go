package service

import (
	"context"
	"database/sql"
	"testing"

	"github.com/ArtemiySps/calc_go_final/internal/orkestrator/config"
	"github.com/ArtemiySps/calc_go_final/pkg/models"
	pb "github.com/ArtemiySps/calc_go_final/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type MockCalcClient struct {
	mock.Mock
}

func (m *MockCalcClient) Calculation(ctx context.Context, in *pb.TaskRequest, opts ...grpc.CallOption) (*pb.ResResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.ResResponse), args.Error(1)
}

func TestOrkestrator_ExpressionOperations(t *testing.T) {
	tests := []struct {
		name        string
		expr        string
		mockRes     float32
		mockErr     error
		expectedRes float64
		expectedErr error
	}{
		{
			name:        "Simple addition",
			expr:        "2+3",
			mockRes:     5,
			mockErr:     nil,
			expectedRes: 5,
			expectedErr: nil,
		},
		{
			name:        "Division by zero",
			expr:        "2/0",
			mockRes:     0,
			mockErr:     models.ErrDivisionByZero,
			expectedRes: 0,
			expectedErr: models.ErrDivisionByZero,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockCalcClient)
			mockClient.On("Calculation", mock.Anything, mock.Anything).Return(
				&pb.ResResponse{Res: tt.mockRes}, tt.mockErr,
			)

			db, err := sql.Open("sqlite3", ":memory:")
			assert.NoError(t, err)
			defer db.Close()

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

			cfg := &config.Config{OperationTimes: map[string]int{"+": 0, "/": 0}}
			logger := zap.NewNop()
			o := &Orkestrator{
				Config:     cfg,
				log:        logger,
				exprs:      db,
				ctx:        context.Background(),
				grpcClient: mockClient,
			}

			res, err := o.ExpressionOperations(tt.expr, "testuser")

			assert.Equal(t, tt.expectedRes, res)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
			mockClient.AssertExpectations(t)
		})
	}
}
