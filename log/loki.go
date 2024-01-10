package log

import (
	"github.com/grafana/loki-client-go/loki"
	slogloki "github.com/samber/slog-loki/v3"
	"github.com/slausonio/siocore"
	"log/slog"
)

func NewLokiClient(env siocore.Environment) (*slog.Logger, *loki.Client) {
	config, _ := loki.NewDefaultConfig(env.Value(siocore.EnvKeyLokiHost))
	config.TenantID = "xyz"
	client, _ := loki.New(config)

	logger := slog.New(slogloki.Option{Level: slog.LevelDebug, Client: client}.NewLokiHandler())
	logger = logger.
		With("env", env.Value(siocore.EnvKeyCurrentEnv))

	return logger, client
}
