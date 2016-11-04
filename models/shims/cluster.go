package shims

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/resourced/resourced-master/contexthelper"
	"github.com/resourced/resourced-master/models/cassandra"
	"github.com/resourced/resourced-master/models/pg"
	"github.com/resourced/resourced-master/models/shared"
)

func NewCluster(ctx context.Context) *Cluster {
	c := &Cluster{}
	c.AppContext = ctx
	return c
}

type Cluster struct {
	Base
}

func (c *Cluster) GetDBType() string {
	generalConfig, err := contexthelper.GetGeneralConfig(c.AppContext)
	if err != nil {
		return ""
	}

	return generalConfig.GetCoreDBType()
}

func (c *Cluster) GetPGDB() (*sqlx.DB, error) {
	pgdbs, err := contexthelper.GetPGDBConfig(c.AppContext)
	if err != nil {
		return nil, err
	}

	return pgdbs.Core, nil
}

func (c *Cluster) GetByID(id int64) (*ClusterRow, error) {
	if c.GetDBType() == "pg" {
		return pg.NewCluster(c.AppContext).GetByID(nil, id)

	} else if c.GetDBType() == "cassandra" {
		return cassandra.NewCluster(c.AppContext).GetByID(id)
	}

	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}

// Create a cluster row record with default settings.
func (c *Cluster) Create(creator *shared.UserRow, name string) (*ClusterRow, error) {
	if c.GetDBType() == "pg" {
		return pg.NewCluster(c.AppContext).Create(nil, creator, name)

	} else if c.GetDBType() == "cassandra" {
		return cassandra.NewCluster(c.AppContext).Create(creator, name)
	}

	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}

// AllByUserID returns all clusters rows by user ID.
func (c *Cluster) AllByUserID(userId int64) ([]*ClusterRow, error) {
	if c.GetDBType() == "pg" {
		return pg.NewCluster(c.AppContext).AllByUserID(nil, userId)

	} else if c.GetDBType() == "cassandra" {
		return cassandra.NewCluster(c.AppContext).AllByUserID(userId)
	}

	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}

// All returns all clusters rows.
func (c *Cluster) All() ([]*ClusterRow, error) {
	if c.GetDBType() == "pg" {
		return pg.NewCluster(c.AppContext).All(nil)

	} else if c.GetDBType() == "cassandra" {
		return cassandra.NewCluster(c.AppContext).All()
	}

	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}

// AllSplitToDaemons returns all rows divided into daemons equally.
func (c *Cluster) AllSplitToDaemons(daemons []string) (map[string][]*ClusterRow, error) {
	if c.GetDBType() == "pg" {
		return pg.NewCluster(c.AppContext).AllSplitToDaemons(nil, daemons)

	} else if c.GetDBType() == "cassandra" {
		return cassandra.NewCluster(c.AppContext).AllSplitToDaemons(daemons)
	}

	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}

// UpdateMember adds or updates cluster member information.
func (c *Cluster) UpdateMember(id int64, user *shared.UserRow, level string, enabled bool) error {
	if c.GetDBType() == "pg" {
		return pg.NewCluster(c.AppContext).UpdateMember(nil, id, user, level, enabled)

	} else if c.GetDBType() == "cassandra" {
		return cassandra.NewCluster(c.AppContext).UpdateMember(id, user, level, enabled)
	}

	return fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}

// RemoveMember from a cluster.
func (c *Cluster) RemoveMember(id int64, user *shared.UserRow) error {
	if c.GetDBType() == "pg" {
		return pg.NewCluster(c.AppContext).RemoveMember(nil, id, user)

	} else if c.GetDBType() == "cassandra" {
		return cassandra.NewCluster(c.AppContext).RemoveMember(id, user)
	}

	return fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}
