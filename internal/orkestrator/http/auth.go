package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ArtemiySps/calc_go_final/pkg/models"
	"github.com/golang-jwt/jwt/v5"
)

// хендлер для регистрации пользователя
func (t *TransportHttp) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = t.s.Register(request.Login, request.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
		t.log.Info("new user registered")
		fmt.Fprintf(w, "user registered successfully!\n")
	}
}

// хендлер для входа поьзователя
func (t *TransportHttp) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := t.s.Login(request.Login, request.Password)
	if err != nil {
		switch err {
		case models.ErrIncorrectPassword:
			http.Error(w, err.Error(), http.StatusUnauthorized)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	response := struct {
		Token string `json:"token"`
	}{
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte("sosiski_123"), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
