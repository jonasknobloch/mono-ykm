package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

var Config = struct {
	AllowTerminalInsertions     bool
	ModelErrorStrategy          string
	TrainingDataPath            string
	TrainingIterationLimit      int
	TrainingSampleLimit         int
	TrainingComplexityLimit     int
	ConcurrentSampleEvaluations int
	CoreNLPUrl                  string
	TreeMockDataPath            string
	InitModelPath               string
	InitModelIteration          int
	ExportGraphs                bool
	ExportModel                 bool
	GraphExportDirectory        string
	ModelExportDirectory        string
}{}

func init() {
	defer fmt.Printf("%+v\n\n", &Config)

	Config.AllowTerminalInsertions, _, _ = parseEnvBool("ALLOW_TERMINAL_INSERTIONS", false)

	Config.ModelErrorStrategy, _ = parseEnvString("MODEL_ERROR_STRATEGY", ErrorStrategyReset)

	validateConst(Config.ModelErrorStrategy, ErrorStrategyIgnore, ErrorStrategyKeep, ErrorStrategyReset)

	Config.TrainingDataPath, _ = parseEnvString("TRAINING_DATA_PATH", "")
	Config.TrainingIterationLimit, _, _ = parseEnvInt("TRAINING_ITERATION_LIMIT", 1)
	Config.TrainingSampleLimit, _, _ = parseEnvInt("TRAINING_SAMPLE_LIMIT", -1)
	Config.TrainingComplexityLimit, _, _ = parseEnvInt("TRAINING_COMPLEXITY_LIMIT", -1)

	Config.ConcurrentSampleEvaluations, _, _ = parseEnvInt("CONCURRENT_SAMPLE_EVALUATIONS", 1)

	Config.CoreNLPUrl, _ = parseEnvString("CORE_NLP_URL", "")
	Config.TreeMockDataPath, _ = parseEnvString("TREE_MOCK_DATA_PATH", "")

	Config.InitModelPath, _ = parseEnvString("INIT_MODEL_PATH", "")
	Config.InitModelIteration, _, _ = parseEnvInt("INIT_MODEL_ITERATION", 1)

	Config.ExportGraphs, _, _ = parseEnvBool("EXPORT_GRAPHS", false)
	Config.ExportModel, _, _ = parseEnvBool("EXPORT_MODEL", true)

	Config.GraphExportDirectory, _ = parseEnvString("GRAPH_EXPORT_DIRECTORY", "")
	Config.ModelExportDirectory, _ = parseEnvString("MODEL_EXPORT_DIRECTORY", "")

	ensureDirectoryExists(Config.GraphExportDirectory)
	ensureDirectoryExists(Config.ModelExportDirectory)
}

func parseEnvString(key, def string) (string, bool) {
	if val, ok := os.LookupEnv(key); ok {
		return val, ok
	}

	return def, false
}

func parseEnvInt(key string, def int) (int, bool, error) {
	if val, ok := os.LookupEnv(key); ok {
		i, err := strconv.Atoi(val)

		if err != nil {
			return def, false, err
		}

		return i, ok, nil
	}

	return def, false, nil
}

func parseEnvBool(key string, def bool) (bool, bool, error) {
	if val, ok := os.LookupEnv(key); ok {
		b, err := strconv.ParseBool(val)

		if err != nil {
			return def, false, err
		}

		return b, ok, nil
	}

	return def, false, nil
}

func validateConst(value string, candidates ...string) {
	valid := false

	for _, c := range candidates {
		if c != value {
			continue
		}

		valid = true
	}

	if !valid {
		log.Fatalf("Invalid value for const: %s not in %v", value, candidates)
	}
}

func ensureDirectoryExists(name string) {
	if name == "" {
		return
	}

	if _, err := os.Stat(name); err != nil && os.IsNotExist(err) {
		if err := os.Mkdir(name, 0755); err != nil {
			panic(err)
		}
	}
}
