package siocore

import (
	"fmt"
	"os"
	"testing"

	"github.com/slausonio/siotest"
	"github.com/stretchr/testify/assert"
)

var currentEnvMap = Env{"test1": "test", "test2": "test2"}

var happyEnvMap = Env{
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

func TestAppNewEnv(t *testing.T) {
	checkOsFunc := func() {
		for key, value := range happyEnvMap {
			os.Getenv(key)
			assert.Equalf(t, os.Getenv(key), value, "expected %v, got %v", os.Getenv(key), value)
		}
	}
	EnvSetup(t)
	EnvCleanup(t)

	appEnv := NewAppEnv()
	env := appEnv.Env()
	assert.NotNilf(t, appEnv.Env(), "expected app env to not be nil")

	assert.Equalf(t, env.Value(EnvKeyCurrentEnv), "test", "expected %v, got %v", env.Value(EnvKeyCurrentEnv), "test")
	assert.Equalf(t, env.Value(EnvKeyAppName), "go-webserver", "expected %v, got %v", env.Value(EnvKeyAppName), "go-webserver")
	assert.Equalf(t, env.Value(EnvKeyPort), "8080", "expected %v, got %v", env.Value(EnvKeyPort), "8080")

	checkOsFunc()
}

func TestAppEnv_Value(t *testing.T) {
	tt := []struct {
		name     string
		env      Env
		key      string
		expected string
	}{
		{
			name:     "Existing Key",
			env:      Env{"existingKey": "existingValue"},
			key:      "existingKey",
			expected: "existingValue",
		},
		{
			name:     "Non-Existing Key",
			env:      Env{"existingKey": "existingValue"},
			key:      "nonExistingKey",
			expected: "",
		},
		{
			name:     "Empty Key",
			env:      Env{"": "emptyKey"},
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

func TestAppEnv_LookupValue(t *testing.T) {
	tt := []struct {
		name        string
		env         Env
		key         string
		expectedVal string
		valExists   bool
	}{
		{
			name:        "Existing Key",
			env:         Env{"existingKey": "existingValue"},
			key:         "existingKey",
			expectedVal: "existingValue",
			valExists:   true,
		},
		{
			name:        "Non-Existing Key",
			env:         Env{"existingKey": "existingValue"},
			key:         "nonExistingKey",
			expectedVal: "",
			valExists:   false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			value, ok := tc.env.LookupValue(tc.key)

			if ok != tc.valExists {
				t.Errorf("expected: %v, got %v", tc.valExists, ok)
			}

			if value != tc.expectedVal {
				t.Errorf("expected: %s, got: %s", tc.expectedVal, value)
			}
		})
	}
}

func TestAppEnv_Update(t *testing.T) {

	t.Run("happy", func(t *testing.T) {
		testKey := "test"
		testVal1 := "test1"
		testVal2 := "test2"
		env := Env{testKey: testVal1}

		value := env.Value(testKey)
		if value != testVal1 {
			t.Errorf("expected: %s, got: %s", testVal1, value)
		}

		env.Update(testKey, testVal2)
		value = env.Value(testKey)
		if value != testVal2 {
			t.Errorf("expected: %s, got: %s", testVal1, value)
		}
	})

}
