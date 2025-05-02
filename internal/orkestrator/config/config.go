package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AdditionTime       time.Duration
	SubtractionTime    time.Duration
	MultiplicationTime time.Duration
	DivisionTime       time.Duration

	OrkestratorPort string
}

func NewConfig() (*Config, error) {
	err := godotenv.Load("./env/.env")
	if err != nil {
		return nil, err
	}

	additionTime, _ := strconv.Atoi(os.Getenv("TIME_ADDITION_MS"))
	substractionTime, _ := strconv.Atoi(os.Getenv("TIME_SUBTRACTION_MS"))
	multiplicationTime, _ := strconv.Atoi(os.Getenv("TIME_MULTIPLICATIONS_MS"))
	divisionTime, _ := strconv.Atoi(os.Getenv("TIME_DIVISIONS_MS"))

	port := os.Getenv("PORT_ORKESTRATOR")

	cfg := &Config{
		AdditionTime:       time.Duration(additionTime),
		SubtractionTime:    time.Duration(substractionTime),
		MultiplicationTime: time.Duration(multiplicationTime),
		DivisionTime:       time.Duration(divisionTime),
		OrkestratorPort:    port,
	}

	return cfg, nil
}

// также, можно обрабатывать ошибки ввода некорректных значений для переменных
