package shared

import (
	"context"
	"os"

	"github.com/resourced/resourced-master/config"
)

func AppContextForTest() context.Context {
	generalConfig := NewGeneralConfigForTest()

	pgDBConfig, _ := config.NewPGDBConfig(generalConfig)

	cassandraDBConfig, _ := config.NewCassandraDBConfig(generalConfig)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "GeneralConfig", NewGeneralConfigForTest())
	ctx = context.WithValue(ctx, "PGDBConfig", pgDBConfig)
	ctx = context.WithValue(ctx, "CassandraDBConfig", cassandraDBConfig)
	ctx = context.WithValue(ctx, "OutLogger", os.Stdout)
	ctx = context.WithValue(ctx, "ErrLogger", os.Stderr)
	ctx = context.WithValue(ctx, "Addr", "localhost:55655")
	ctx = context.WithValue(ctx, "CookieStore", "abc123")

	// for key, mailr := range app.Mailers {
	// 	ctx = context.WithValue(ctx, "mailer."+key, mailr)
	// }

	// ctx = context.WithValue(ctx, "bus", app.MessageBus)

	return ctx
}

func NewGeneralConfigForTest() config.GeneralConfig {
	cfg := config.GeneralConfig{}
	cfg.Addr = "localhost:55655"
	cfg.PostgreSQL.DSN = "postgres://localhost:5432/resourced-master-test?sslmode=disable"
	cfg.Hosts.PostgreSQL.DSN = "postgres://localhost:5432/resourced-master-test?sslmode=disable"
	cfg.Events.PostgreSQL.DSN = "postgres://localhost:5432/resourced-master-test?sslmode=disable"
	cfg.Logs.PostgreSQL.DSN = "postgres://localhost:5432/resourced-master-test?sslmode=disable"
	cfg.Checks.PostgreSQL.DSN = "postgres://localhost:5432/resourced-master-test?sslmode=disable"
	cfg.Metrics.PostgreSQL.DSN = "postgres://localhost:5432/resourced-master-test?sslmode=disable"
	return cfg
}
