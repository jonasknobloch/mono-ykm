package main

import (
	"os"
	"strconv"
)

var Config = struct {
	AllowTerminalInsertions bool
	TrainingDataPath        string
	TrainingIterationLimit  int
	TrainingSampleLimit     int
	TrainingComplexityLimit int
	CoreNLPUrl              string
	TreeMockDataPath        string
	ExportGraphs            bool
	ExportModel             bool
	GraphExportDirectory    string
	ModelExportDirectory    string
}{}

func init() {
	Config.AllowTerminalInsertions, _, _ = parseEnvBool("ALLOW_TERMINAL_INSERTIONS", false)

	Config.TrainingDataPath, _ = parseEnvString("TRAINING_DATA_PATH", "")
	Config.TrainingIterationLimit, _, _ = parseEnvInt("TRAINING_ITERATION_LIMIT", 1)
	Config.TrainingSampleLimit, _, _ = parseEnvInt("TRAINING_SAMPLE_LIMIT", -1)
	Config.TrainingComplexityLimit, _, _ = parseEnvInt("TRAINING_COMPLEXITY_LIMIT", -1)

	Config.CoreNLPUrl, _ = parseEnvString("CORE_NLP_URL", "")
	Config.TreeMockDataPath, _ = parseEnvString("TREE_MOCK_DATA_PATH", "")

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
