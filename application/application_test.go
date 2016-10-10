package application

import (
	"context"
	"os"
	"testing"

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

func TestConstructor(t *testing.T) {
	app, err := New("../tests/config-files")
	if err != nil {
		t.Errorf("Creating a new app should not fail using test config. Error: %v", err)
	}
	hostname, _ := os.Hostname()
	if app.Hostname != hostname {
		t.Errorf("Failed to configure hostname properly")
	}
}
