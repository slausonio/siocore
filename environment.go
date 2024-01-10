package siocore

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

const (
	EnvKeyAppName    = "APP_NAME"
	EnvKeyCurrentEnv = "CURRENT_ENV"
	EnvKeyPort       = "PORT"
	EnvKeyLokiHost   = "LOKI_HOST"

	DefaultFilePath    = "env/.env"
	CurrentEnvFilePath = "env/%s.env"
)

var (
	ErrNoEnvFile    = errors.New("no .env file found in root of project")
	ErrNoAppName    = errors.New("no APP_NAME env var found")
	ErrNoCurrentEnv = errors.New("no CURRENT_ENV env var found")
)

// Env is a type that represents a map of string key-value pairs for environment variables.
type Env map[string]string

// Value retrieves the value associated with the specified key in the SioWSEnv map.
// If the key does not exist in the map, an empty string is returned.
func (e Env) Value(key string) string {
	return e[key]
}

// Update modifies the value associated with the given key in the SioWSEnv map. If the key does not exist, a new key-value pair is added.
func (e Env) Update(key, value string) {
	e[key] = value
}

type AppEnv struct {
	defaultEnv Env
	currentEnv Env
	mergedEnv  Env
	log        *slog.Logger
}

// NewEnvironment creates a new SioWSEnv environment.
// It reads the default environment variables from a file,
// merges them with environment-specific variables,
// and sets the environment variables to the system.
// It returns the merged environment.
func NewAppEnv(log *slog.Logger) *AppEnv {
	appEnv := &AppEnv{log: log}
	env := appEnv.readEnvironment()

	appEnv.setEnvToSystem()

	return &AppEnv{
		defaultEnv: env,
		log:        log,
	}
}

// readEnvironment reads the environment configuration by merging the default environment file,
// the current environment file, and setting the environment variables
func (ae *AppEnv) readEnvironment() Env {
	defaultEnvMap := readDefaultEnvFile(ae.log)
	defaultEnvMap.setEnvToSystem()
	ae.defaultEnv = defaultEnvMap

	currentEnv := readDefaultEnvFile(ae.log)
	currentEnvMap := readEnvironmentSpecificFile(currentEnv.Value(EnvKeyCurrentEnv), ae.log)
	ae.currentEnv = currentEnvMap

	mergedEnv := MergeMaps(defaultEnvMap, currentEnvMap)

	return mergedEnv
}

// setEnvToSystem sets the environment variables in the SioWSEnv map to the system.
// It iterates over the key-value pairs in the map and uses os.Setenv to set each variable.
// If there is an error setting the variable, it panics with the error.
func (e Env) setEnvToSystem() {
	for key, value := range e {
		err := os.Setenv(key, value)
		if err != nil {
			panic(err)
		}
	}
}

// readDefaultEnvFile reads the default environment file located at DefaultFilePath and returns its contents as a SioWSEnv map.
// If the file cannot be read or an error occurs, it logs the error and panics with ErrNoEnvFile.
func readDefaultEnvFile(log *slog.Logger) Env {
	defaultEnvFile, err := godotenv.Read(DefaultFilePath)
	if err != nil {
		log.Error("dotenv error: ", slog.AnyValue(err))
		panic(ErrNoEnvFile)
	}

	return defaultEnvFile
}

// readEnvironmentSpecificFile reads the environment-specific file based on the given environment.
// It takes an `env` string parameter indicating the environment.
// It returns an instance of the `SioWSEnv` type that represents the environment-specific file.
func readEnvironmentSpecificFile(env string, log *slog.Logger) Env {
	fileName := fmt.Sprintf(CurrentEnvFilePath, env)

	defaultEnvFile, err := godotenv.Read(fileName)
	if err != nil {
		log.Info("dotenv error: ", slog.AnyValue(err))
	}

	return defaultEnvFile
}

// readCurrentEnv reads the value of the `CURRENT_ENV` environment variable.
// If the environment variable is not found, it raises an error and panics.
// It returns the value of the `CURRENT_ENV` environment variable.
func readCurrentEnv(log *slog.Logger) string {
	appName, ok := os.LookupEnv(EnvKeyCurrentEnv)
	if !ok {
		err := fmt.Errorf("new environment: %w", ErrNoCurrentEnv)

		log.Error(err.Error())
		panic(err)
	}

	return appName
}

// readAppName reads the value of the environment variable specified by AppNameKey,
// which is the key for the application name.
// If the environment variable is not found, it logs an error and panics with an error message.
// It returns the value of the environment variable as a string.
func readAppName(log *slog.Logger) string {
	appName, ok := os.LookupEnv(EnvKeyAppName)
	if !ok {
		err := fmt.Errorf("new environment: %w", ErrNoAppName)

		log.Error(err.Error())
		panic(err)
	}

	return appName
}
