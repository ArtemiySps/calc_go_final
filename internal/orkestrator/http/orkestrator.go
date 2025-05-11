package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/tabwriter"

	"github.com/ArtemiySps/calc_go_final/pkg/models"
)

func writeTable(expressions map[string]models.Expression) string {
	var sb strings.Builder

	w := tabwriter.NewWriter(&sb, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "ID\tExpression\tStatus\tResult\tError")

	for _, expr := range expressions {
		resultStr := fmt.Sprintf("%.2f", expr.Result)
		resultErr := expr.Error
		if expr.Status == models.StatusFailed || expr.Status == models.StatusPending {
			resultStr = "none"
		} else {
			resultErr = "none"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			expr.ID,
			expr.Expr,
			expr.Status,
			resultStr,
			resultErr,
		)
	}

	w.Flush()
	return sb.String()
}

// хендлер для оркестратора. доступен по ручке "/api/v1/calculate"
func (t *TransportHttp) OrkestratorHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Expression string `json:"expression"`
	}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tokenString := r.Header.Get("Authorization")
	login, err := t.s.GetLogin(tokenString)
	if err != nil {
		t.log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := t.s.ExpressionOperations(request.Expression, login)
	if err != nil {
		t.log.Error(err.Error())
		switch err {
		case models.ErrBadExpression:
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		case models.ErrUnexpectedSymbol:
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	data, err := json.Marshal(map[string]any{
		"result": res,
	})
	if err != nil {
		t.log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(data)
	if err != nil {
		t.log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// хендлер для получения всех выражений. доступен по ручке "/api/v1/expressions"
func (t *TransportHttp) GetAllExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("Authorization")
	login, err := t.s.GetLogin(tokenString)
	if err != nil {
		t.log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	expressions, err := t.s.GetAllExpressions(login)
	if err != nil {
		t.log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if t.table {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		table := writeTable(expressions)
		w.Write([]byte(table))
		return
	}

	response := struct {
		Exprs map[string]models.Expression `json:"expressions"`
	}{
		Exprs: expressions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// хендлер для получения выражения по id. доступен по ручке "/api/v1/expression/<id>"
func (t *TransportHttp) GetExpressionHandler(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("Authorization")
	login, err := t.s.GetLogin(tokenString)
	if err != nil {
		t.log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	expression, err := t.s.GetExpression(r.URL.Path[len("/api/v1/expression/"):], login)
	if err != nil {
		t.log.Error(models.ErrCannotFindObject.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if t.table {
		expressions := make(map[string]models.Expression)
		expressions[expression.ID] = expression

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		table := writeTable(expressions)
		w.Write([]byte(table))
		return
	}

	response := struct {
		Exprs models.Expression `json:"expression"`
	}{
		Exprs: expression,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (t *TransportHttp) ClearHandler(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("Authorization")
	login, err := t.s.GetLogin(tokenString)
	if err != nil {
		t.log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := t.s.Clear(login)
	if err != nil {
		t.log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rows == 0 {
		fmt.Fprintf(w, "nothing to clear\n")
		return
	}
	fmt.Fprintf(w, "deleted %d expressions\n", rows)
}
