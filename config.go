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
	CoreNLPUrl              string
	GraphExportDirectory    string
	ModelExportDirectory    string
	ValidateEdges           bool
	CountOrphans            bool
}{}

func init() {
	Config.AllowTerminalInsertions, _, _ = parseEnvBool("ALLOW_TERMINAL_INSERTIONS", false)

	Config.TrainingDataPath, _ = parseEnvString("TRAINING_DATA_PATH", "")
	Config.TrainingIterationLimit, _, _ = parseEnvInt("TRAINING_ITERATION_LIMIT", 1)
	Config.TrainingSampleLimit, _, _ = parseEnvInt("TRAINING_SAMPLE_LIMIT", -1)

	Config.CoreNLPUrl, _ = parseEnvString("CORE_NLP_URL", "")

	Config.GraphExportDirectory, _ = parseEnvString("GRAPH_EXPORT_DIRECTORY", "")
	Config.ModelExportDirectory, _ = parseEnvString("MODEL_EXPORT_DIRECTORY", "")

	ensureDirectoryExists(Config.GraphExportDirectory)
	ensureDirectoryExists(Config.ModelExportDirectory)

	Config.ValidateEdges, _, _ = parseEnvBool("VALIDATE_EDGES", false)
	Config.CountOrphans, _, _ = parseEnvBool("COUNT_ORPHANS", false)
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
