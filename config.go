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
}{
	AllowTerminalInsertions: false,
	TrainingIterationLimit:  1,
	TrainingSampleLimit:     -1,
}

func init() {
	if val, ok := os.LookupEnv("ALLOW_TERMINAL_INSERTIONS"); ok {
		if b, err := strconv.ParseBool(val); err == nil {
			Config.AllowTerminalInsertions = b
		}
	}

	if val, ok := os.LookupEnv("TRAINING_DATA_PATH"); ok {
		Config.TrainingDataPath = val
	}

	if val, ok := os.LookupEnv("TRAINING_ITERATION_LIMIT"); ok {
		if i, err := strconv.Atoi(val); err == nil {
			Config.TrainingIterationLimit = i
		}
	}

	if val, ok := os.LookupEnv("TRAINING_SAMPLE_LIMIT"); ok {
		if i, err := strconv.Atoi(val); err == nil {
			Config.TrainingSampleLimit = i
		}
	}

	if val, ok := os.LookupEnv("CORE_NLP_URL"); ok {
		Config.CoreNLPUrl = val
	}

	if val, ok := os.LookupEnv("GRAPH_EXPORT_DIRECTORY"); ok {
		Config.GraphExportDirectory = val
	}

	ensureDirectoryExists(Config.GraphExportDirectory)

	if val, ok := os.LookupEnv("MODEL_EXPORT_DIRECTORY"); ok {
		Config.ModelExportDirectory = val
	}

	ensureDirectoryExists(Config.ModelExportDirectory)
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
