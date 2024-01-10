package siocore

import (
	"fmt"
	"github.com/slausonio/siotest"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var currentEnvMap = Environment{"test1": "test", "test2": "test2"}

var happyEnvMap = Environment{
	EnvKeyCurrentEnv: "test",
	EnvKeyAppName:    "go-webserver",
	EnvKeyPort:       "8080",
}

func EnvSetup(t *testing.T) {
	t.Helper()

	siotest.CreateFile(t, DefaultFilePath)
	siotest.CreateFile(t, fmt.Sprintf(CurrentEnvFilePath, "test"))

	siotest.WriteEnvToFile(t, DefaultFilePath, happyEnvMap)
	siotest.WriteEnvToFile(t, fmt.Sprintf(CurrentEnvFilePath, "test"), currentEnvMap)

}

func EnvCleanup(t *testing.T) {
	t.Helper()

	t.Cleanup(func() {
		siotest.RemoveFileAndDirs(t, DefaultFilePath)
	})
}

func TestNewEnvironment(t *testing.T) {
	checkOsFunc := func() {
		for key, value := range happyEnvMap {
			os.Getenv(key)
			assert.Equalf(t, os.Getenv(key), value, "expected %v, got %v", os.Getenv(key), value)
		}
	}
	EnvSetup(t)
	EnvCleanup(t)

	env := NewAppEnv()
	assert.Equalf(t, env.Value(EnvKeyCurrentEnv), "test", "expected %v, got %v", env.Value(EnvKeyCurrentEnv), "test")
	assert.Equalf(t, env.Value(EnvKeyAppName), "go-webserver", "expected %v, got %v", env.Value(EnvKeyAppName), "go-webserver")
	assert.Equalf(t, env.Value(EnvKeyPort), "8080", "expected %v, got %v", env.Value(EnvKeyPort), "8080")

	checkOsFunc()
}

func TestSioWSEnv_Value(t *testing.T) {
	tt := []struct {
		name     string
		env      Environment
		key      string
		expected string
	}{
		{
			name:     "Existing Key",
			env:      Environment{"existingKey": "existingValue"},
			key:      "existingKey",
			expected: "existingValue",
		},
		{
			name:     "Non-Existing Key",
			env:      Environment{"existingKey": "existingValue"},
			key:      "nonExistingKey",
			expected: "",
		},
		{
			name:     "Empty Key",
			env:      Environment{"": "emptyKey"},
			key:      "",
			expected: "emptyKey",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			value := tc.env.Value(tc.key)

			if value != tc.expected {
				t.Errorf("expected: %s, got: %s", tc.expected, value)
			}
		})
	}
}
