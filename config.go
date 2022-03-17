package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
)

var Config = struct {
	ReplaceSparseTokens            bool
	SparseTokenThreshold           int
	EnableInteriorInsertions       bool
	EnableTerminalInsertions       bool
	EnablePhrasalTranslations      bool
	EnableInterior1ToNTranslations bool
	EnableInteriorNTo1Translations bool
	EnableFertilityDecomposition   bool
	PhraseLengthLimit              int
	PhraseFrequencyCutoff          int
	TrainingDataPath               string
	TrainingIterationLimit         int
	TrainingSampleLimit            int
	TrainingComplexityLimit        int
	ConcurrentSampleEvaluations    int
	ParaphraseThreshold            float64
	InitModelPath                  string
	InitModelIteration             int
	PrintCorpusLikelihood          bool
	ExportGraphs                   bool
	ExportModel                    bool
	GraphExportDirectory           string
	ModelExportDirectory           string
}{}

func init() {
	defer fmt.Printf("%+v\n\n", &Config)

	Config.ReplaceSparseTokens, _, _ = parseEnvBool("REPLACE_SPARSE_TOKENS", false)
	Config.SparseTokenThreshold, _, _ = parseEnvInt("SPARSE_TOKEN_THRESHOLD", 1)

	Config.EnableInteriorInsertions, _, _ = parseEnvBool("ENABLE_INTERIOR_INSERTIONS", false)
	Config.EnableTerminalInsertions, _, _ = parseEnvBool("ENABLE_TERMINAL_INSERTIONS", true)

	Config.EnablePhrasalTranslations, _, _ = parseEnvBool("ENABLE_PHRASAL_TRANSLATIONS", true)
	Config.EnableInterior1ToNTranslations, _, _ = parseEnvBool("ENABLE_INTERIOR_1_TO_N_TRANSLATIONS", true)
	Config.EnableInteriorNTo1Translations, _, _ = parseEnvBool("ENABLE_INTERIOR_N_TO_1_TRANSLATIONS", true)
	Config.EnableFertilityDecomposition, _, _ = parseEnvBool("ENABLE_FERTILITY_DECOMPOSITION", true)
	Config.PhraseLengthLimit, _, _ = parseEnvInt("PHRASE_LENGTH_LIMIT", 0)
	Config.PhraseFrequencyCutoff, _, _ = parseEnvInt("PHRASE_FREQUENCY_CUTOFF", 1)

	Config.TrainingDataPath, _ = parseEnvString("TRAINING_DATA_PATH", "")
	Config.TrainingIterationLimit, _, _ = parseEnvInt("TRAINING_ITERATION_LIMIT", 1)
	Config.TrainingSampleLimit, _, _ = parseEnvInt("TRAINING_SAMPLE_LIMIT", -1)
	Config.TrainingComplexityLimit, _, _ = parseEnvInt("TRAINING_COMPLEXITY_LIMIT", -1)

	Config.ConcurrentSampleEvaluations, _, _ = parseEnvInt("CONCURRENT_SAMPLE_EVALUATIONS", 1)

	Config.ParaphraseThreshold, _, _ = parseEnvFloat64("PARAPHRASE_THRESHOLD", math.SmallestNonzeroFloat64)

	Config.InitModelPath, _ = parseEnvString("INIT_MODEL_PATH", "")
	Config.InitModelIteration, _, _ = parseEnvInt("INIT_MODEL_ITERATION", 1)

	Config.PrintCorpusLikelihood, _, _ = parseEnvBool("PRINT_CORPUS_LIKELIHOOD", false)

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

func parseEnvFloat64(key string, def float64) (float64, bool, error) {
	if val, ok := os.LookupEnv(key); ok {
		i, err := strconv.ParseFloat(val, 64)

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
