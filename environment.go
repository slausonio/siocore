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

func (e Env) LookupValue(key string) (string, bool) {
	val := e.Value(key)
	if val == "" {
		return "", false
	}

	return val, true
}

// ValuesPresent validates required authz variables are present.  If not, the thread will panic
func (e Env) ValuesPresent(envVarKeys []string) {
	for _, key := range envVarKeys {
		value, present := e.LookupValue(key)
		if !present || value == "" {
			panic(
				fmt.Sprintf(
					"The environment variable %s is not present.\n Unable to start application.",
					key,
				),
			)
		}
	}
}

// setToSystem sets the environment variables in the SioWSEnv map to the system.
// It iterates over the key-value pairs in the map and uses os.Setenv to set each variable.
// If there is an error setting the variable, it panics with the error.
func (e Env) setToSystem() {
	for key, value := range e {
		err := os.Setenv(key, value)
		if err != nil {
			panic(err)
		}
	}
}

type AppEnv struct {
	env Env
}

// NewAppEnv creates a new SioWSEnv environment.
// It reads the default environment variables from a file,
// merges them with environment-specific variables,
// and sets the environment variables to the system.
// It returns the merged environment.
func NewAppEnv() *AppEnv {
	appEnv := &AppEnv{}
	appEnv.readEnvironment()

	return appEnv
}

func (ae *AppEnv) Env() Env {
	return ae.env
}

// readEnvironment reads the environment configuration by merging the default environment file,
// the current environment file, and setting the environment variables
func (ae *AppEnv) readEnvironment() {
	defaultEnvMap := readDefaultEnvFile()
	defaultEnvMap.setToSystem()

	defaultEnvMap.ValuesPresent([]string{EnvKeyAppName, EnvKeyCurrentEnv})

	currentEnvMap := readEnvironmentSpecificFile(defaultEnvMap.Value(EnvKeyCurrentEnv))
	currentEnvMap.setToSystem()

	mergedEnv := MergeEnvs(defaultEnvMap, currentEnvMap)

	ae.env = mergedEnv
}

// readDefaultEnvFile reads the default environment file located at DefaultFilePath and returns its contents as a SioWSEnv map.
// If the file cannot be read or an error occurs, it logs the error and panics with ErrNoEnvFile.
func readDefaultEnvFile() Env {
	defaultEnvFile, err := godotenv.Read(DefaultFilePath)
	if err != nil {
		slog.Error(fmt.Sprintf("default .env dotenv error: %v", err))
		panic(ErrNoEnvFile)
	}

	return defaultEnvFile
}

// readEnvironmentSpecificFile reads the environment-specific file based on the given environment.
// It takes an `env` string parameter indicating the environment.
// It returns an instance of the `SioWSEnv` type that represents the environment-specific file.
func readEnvironmentSpecificFile(env string) Env {
	fileName := fmt.Sprintf(CurrentEnvFilePath, env)

	defaultEnvFile, err := godotenv.Read(fileName)
	if err != nil {
		slog.Info("environment specific .env dotenv error: ", err)
	}

	return defaultEnvFile
}

// readCurrentEnv reads the value of the `CURRENT_ENV` environment variable.
// If the environment variable is not found, it raises an error and panics.
// It returns the value of the `CURRENT_ENV` environment variable.
func readCurrentEnv() string {
	appName, ok := os.LookupEnv(EnvKeyCurrentEnv)
	if !ok {
		err := fmt.Errorf("new environment: %w", ErrNoCurrentEnv)

		slog.Error(err.Error())
		panic(err)
	}

	return appName
}

// readAppName reads the value of the environment variable specified by AppNameKey,
// which is the key for the application name.
// If the environment variable is not found, it logs an error and panics with an error message.
// It returns the value of the environment variable as a string.
func readAppName() string {
	appName, ok := os.LookupEnv(EnvKeyAppName)
	if !ok {
		err := fmt.Errorf("new environment: %w", ErrNoAppName)

		slog.Error(err.Error())
		panic(err)
	}

	return appName
}
