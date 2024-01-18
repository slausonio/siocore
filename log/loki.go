package log

import (
	"errors"
	"log/slog"

	"github.com/grafana/loki-client-go/loki"
	slogloki "github.com/samber/slog-loki/v3"
	"github.com/slausonio/siocore"
)

var (
	ErrNoLokiHost = errors.New("no LOKI_HOST env var found")
)

type AppLogger struct {
	logger *slog.Logger
	client *loki.Client
}

func NewSlog(env siocore.Env) *AppLogger {
	config, _ := loki.NewDefaultConfig(getLokiHost(env))
	config.TenantID = "xyz"
	client, _ := loki.New(config)

	logger := slog.New(slogloki.Option{Level: slog.LevelDebug, Client: client}.NewLokiHandler())
	logger = logger.
		With("env", env.Value(siocore.EnvKeyCurrentEnv))

	return &AppLogger{logger, client}
}

func (al *AppLogger) Logger() *slog.Logger {
	return al.logger
}

func (al *AppLogger) SetAsDefault() {
	slog.SetDefault(al.logger)
}

func getLokiHost(env siocore.Env) string {
	lokiHost := env.Value(siocore.EnvKeyLokiHost)
	if lokiHost == "" {
		panic(ErrNoLokiHost)
	}

	return lokiHost
}
