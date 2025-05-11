package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ArtemiySps/calc_go_final/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) ExpressionOperations(expr string, user string) (float64, error) {
	args := m.Called(expr, user)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockService) GetAllExpressions(user string) (map[string]models.Expression, error) {
	args := m.Called(user)
	return args.Get(0).(map[string]models.Expression), args.Error(1)
}

func (m *MockService) GetExpression(id string, user string) (models.Expression, error) {
	args := m.Called(id, user)
	return args.Get(0).(models.Expression), args.Error(1)
}

func (m *MockService) Clear(user string) (int64, error) {
	args := m.Called(user)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockService) Register(login string, password string) error {
	args := m.Called(login, password)
	return args.Error(0)
}

func (m *MockService) Login(login string, password string) (string, error) {
	args := m.Called(login, password)
	return args.String(0), args.Error(1)
}

func (m *MockService) GetLogin(token string) (string, error) {
	args := m.Called(token)
	return args.String(0), args.Error(1)
}

func TestRegisterHandler(t *testing.T) {
	mockService := new(MockService)

	logger := zap.NewNop()

	transport := &TransportHttp{
		s:    mockService,
		log:  logger,
		port: "8080",
	}

	tests := []struct {
		name           string
		payload        map[string]string
		mockReturn     error
		expectedStatus int
	}{
		{
			name: "Successful registration",
			payload: map[string]string{
				"login":    "testuser",
				"password": "password123",
			},
			mockReturn:     nil,
			expectedStatus: http.StatusOK,
		},
		{
			name: "Duplicate registration",
			payload: map[string]string{
				"login":    "existinguser",
				"password": "password123",
			},
			mockReturn:     models.ErrUserAlreadyExists,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.On("Register", tt.payload["login"], tt.payload["password"]).Return(tt.mockReturn)

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/api/v1/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			transport.RegisterHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			mockService.AssertExpectations(t)
		})
	}
}

func TestLoginHandler(t *testing.T) {
	mockService := new(MockService)
	logger := zap.NewNop()

	transport := &TransportHttp{
		s:    mockService,
		log:  logger,
		port: "8080",
	}

	tests := []struct {
		name           string
		payload        map[string]string
		mockToken      string
		mockError      error
		expectedStatus int
		expectToken    bool
	}{
		{
			name: "Successful login",
			payload: map[string]string{
				"login":    "testuser",
				"password": "correctpass",
			},
			mockToken:      "valid.token.here",
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectToken:    true,
		},
		{
			name: "Wrong password",
			payload: map[string]string{
				"login":    "testuser",
				"password": "wrongpass",
			},
			mockToken:      "",
			mockError:      models.ErrIncorrectPassword,
			expectedStatus: http.StatusUnauthorized,
			expectToken:    false,
		},
		{
			name: "Non-existent user",
			payload: map[string]string{
				"login":    "unknown",
				"password": "password123",
			},
			mockToken:      "",
			mockError:      models.ErrUserNotRegistered,
			expectedStatus: http.StatusInternalServerError,
			expectToken:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.On("Login", tt.payload["login"], tt.payload["password"]).
				Return(tt.mockToken, tt.mockError)

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			transport.LoginHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectToken {
				var response struct {
					Token string `json:"token"`
				}
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockToken, response.Token)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestOrkestratorHandler(t *testing.T) {
	mockService := new(MockService)
	logger := zap.NewNop()
	transport := &TransportHttp{
		s:    mockService,
		log:  logger,
		port: "8080",
	}

	tests := []struct {
		name           string
		payload        map[string]string
		token          string
		mockLogin      string
		mockResult     float64
		mockError      error
		expectedStatus int
		expectedResult float64
	}{
		{
			name: "Successful calculation",
			payload: map[string]string{
				"expression": "2+2",
			},
			token:          "valid.token",
			mockLogin:      "testuser",
			mockResult:     4,
			mockError:      nil,
			expectedStatus: http.StatusCreated,
			expectedResult: 4,
		},
		{
			name: "Invalid expression",
			payload: map[string]string{
				"expression": "2/0",
			},
			token:          "valid.token",
			mockLogin:      "testuser",
			mockResult:     0,
			mockError:      models.ErrBadExpression,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedResult: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.On("GetLogin", tt.token).Return(tt.mockLogin, nil)
			mockService.On("ExpressionOperations", tt.payload["expression"], tt.mockLogin).
				Return(tt.mockResult, tt.mockError)

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/api/v1/calculate", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", tt.token)

			rr := httptest.NewRecorder()

			transport.OrkestratorHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.mockError == nil {
				var response struct {
					Result float64 `json:"result"`
				}
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, response.Result)
			}

			mockService.AssertExpectations(t)
		})
	}
}
