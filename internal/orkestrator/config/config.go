package config

import (
	"os"
	"strconv"

	"github.com/ArtemiySps/calc_go_final/pkg/models"
	"github.com/joho/godotenv"
)

type Config struct {
	OperationTimes map[string]int

	OrkestratorPort string

	AgentPort string
	AgentHost string

	UsersDBPath string
}

func NewConfig() (*Config, error) {
	err := godotenv.Load("./env/.env")
	if err != nil {
		return nil, err
	}

	operationTimes := make(map[string]int)
	operationTimes["+"], err = strconv.Atoi(os.Getenv("TIME_ADDITION_MS"))
	if err != nil {
		return nil, models.ErrAdittionTime
	}
	operationTimes["-"], err = strconv.Atoi(os.Getenv("TIME_SUBTRACTION_MS"))
	if err != nil {
		return nil, models.ErrSubtracrionTime
	}
	operationTimes["*"], err = strconv.Atoi(os.Getenv("TIME_MULTIPLICATION_MS"))
	if err != nil {
		return nil, models.ErrMultiplicationTime
	}
	operationTimes["/"], err = strconv.Atoi(os.Getenv("TIME_DIVISION_MS"))
	if err != nil {
		return nil, models.ErrDivisionTime
	}

	port := os.Getenv("PORT_ORKESTRATOR")
	agentPort := os.Getenv("PORT_AGENT")
	agentHost := os.Getenv("HOST_AGENT")

	cfg := &Config{
		OperationTimes:  operationTimes,
		OrkestratorPort: port,
		AgentPort:       agentPort,
		AgentHost:       agentHost,
		UsersDBPath:     "./db/store.db",
	}

	return cfg, nil
}
