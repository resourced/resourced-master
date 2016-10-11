package contexthelper

import (
	"context"
	"errors"

	"github.com/Sirupsen/logrus"

	"github.com/resourced/resourced-master/config"
	"github.com/resourced/resourced-master/mailer"
	"github.com/resourced/resourced-master/messagebus"
)

func GetGeneralConfig(ctx context.Context) (config.GeneralConfig, error) {
	valInterface := ctx.Value("GeneralConfig")
	if valInterface == nil {
		return config.GeneralConfig{}, errors.New("Application general config should never be nil")
	}

	return valInterface.(config.GeneralConfig), nil
}

func GetPGDBConfig(ctx context.Context) (*config.PGDBConfig, error) {
	valInterface := ctx.Value("PGDBConfig")
	if valInterface == nil {
		return nil, errors.New("PG config should never be nil")
	}

	return valInterface.(*config.PGDBConfig), nil
}

func GetCassandraDBConfig(ctx context.Context) (*config.CassandraDBConfig, error) {
	valInterface := ctx.Value("CassandraDBConfig")
	if valInterface == nil {
		return nil, errors.New("Cassandra config should never be nil")
	}

	return valInterface.(*config.CassandraDBConfig), nil
}

func GetMessageBus(ctx context.Context) (*messagebus.MessageBus, error) {
	valInterface := ctx.Value("bus")
	if valInterface == nil {
		return nil, errors.New("bus is nil")
	}

	return valInterface.(*messagebus.MessageBus), nil
}

func GetLogger(ctx context.Context, name string) (*logrus.Logger, error) {
	valInterface := ctx.Value(name)
	if valInterface == nil {
		return nil, errors.New(name + " is nil")
	}

	return valInterface.(*logrus.Logger), nil
}

func GetMailer(ctx context.Context, name string) (*mailer.Mailer, error) {
	valInterface := ctx.Value("mailer." + name)
	if valInterface == nil {
		return nil, errors.New(name + " is nil")
	}

	return valInterface.(*mailer.Mailer), nil
}
