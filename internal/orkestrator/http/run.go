package http

import (
	"net/http"
	"os"
	"strconv"

	"github.com/ArtemiySps/calc_go_final/pkg/models"
	"go.uber.org/zap"
)

type Service interface {
	ExpressionOperations(expr string, user string) (float64, error)
	GetAllExpressions(user string) (map[string]models.Expression, error)
	GetExpression(id string, user string) (models.Expression, error)
	Clear(user string) (int64, error)

	Register(login string, password string) error
	Login(login string, password string) (string, error)
	GetLogin(token string) (string, error)
}

type TransportHttp struct {
	s     Service
	port  string
	table bool
	log   *zap.Logger
}

func NewTransportHttp(s Service, port string, logger *zap.Logger) (*TransportHttp, error) {
	table, err := strconv.ParseBool(os.Getenv("TABLE_FORMAT"))
	if err != nil {
		return &TransportHttp{}, models.ErrTableFormat
	}

	t := &TransportHttp{
		s:     s,
		port:  port,
		table: table,
		log:   logger,
	}

	http.Handle("/api/v1/calculate", AuthMiddleware(http.HandlerFunc(t.OrkestratorHandler)))
	http.Handle("/api/v1/expressions", AuthMiddleware(http.HandlerFunc(t.GetAllExpressionsHandler)))
	http.Handle("/api/v1/expression/", AuthMiddleware(http.HandlerFunc(t.GetExpressionHandler)))
	http.Handle("/api/v1/clear", AuthMiddleware(http.HandlerFunc(t.ClearHandler)))

	http.HandleFunc("/api/v1/register", t.RegisterHandler)
	http.HandleFunc("/api/v1/login", t.LoginHandler)

	return t, nil
}

// запуск http сервера
func (t *TransportHttp) RunServer() {
	t.log.Info("Server (orkestrator) starting on port " + t.port)

	http.ListenAndServe(":"+t.port, nil)
}
