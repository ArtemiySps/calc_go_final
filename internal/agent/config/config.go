package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	GetTaskInterval time.Duration
	ComputingPower  int

	OrkestratorPort string
	AgentPort       string
}

func NewConfig() (*Config, error) {
	err := godotenv.Load("./env/.env")
	if err != nil {
		return nil, err
	}

	getTaskInterval, _ := strconv.Atoi(os.Getenv("GET_TASK_INTERVAL_MS"))
	computingPower, _ := strconv.Atoi(os.Getenv("COMPUTING_POWER"))

	orkestratorPort := os.Getenv("PORT_ORKESTRATOR")
	agentPort := os.Getenv("PORT_AGENT")

	cfg := &Config{
		GetTaskInterval: time.Duration(getTaskInterval) * time.Millisecond,
		ComputingPower:  computingPower,
		OrkestratorPort: orkestratorPort,
		AgentPort:       agentPort,
	}

	return cfg, nil
}

// также, можно обрабатывать ошибки ввода некорректных значений для переменных
