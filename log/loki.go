package log

import (
	"errors"
	"github.com/grafana/loki-client-go/loki"
	slogloki "github.com/samber/slog-loki/v3"
	"github.com/slausonio/siocore"
	"log/slog"
)

var (
	ErrNoLokiHost = errors.New("no CURRENT_ENV env var found")
)

func NewLokiClient(env siocore.Env) (*slog.Logger, *loki.Client) {
	config, _ := loki.NewDefaultConfig(getLokiHost(env))
	config.TenantID = "xyz"
	client, _ := loki.New(config)

	logger := slog.New(slogloki.Option{Level: slog.LevelDebug, Client: client}.NewLokiHandler())
	logger = logger.
		With("env", env.Value(siocore.EnvKeyCurrentEnv))

	return logger, client
}

func getLokiHost(env siocore.Env) string {
	lokiHost := env.Value(siocore.EnvKeyLokiHost)
	if lokiHost == "" {
		panic(ErrNoLokiHost)
	}

	return lokiHost
}
