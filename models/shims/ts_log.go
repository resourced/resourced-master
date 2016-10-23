package shims

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/resourced/resourced-master/contexthelper"
	"github.com/resourced/resourced-master/models/cassandra"
	"github.com/resourced/resourced-master/models/pg"
	"github.com/resourced/resourced-master/models/shared"
)

func NewTSLog(ctx context.Context, clusterID int64) *TSLog {
	ts := &TSLog{}
	ts.AppContext = ctx
	ts.ClusterID = clusterID
	return ts
}

type TSLog struct {
	Base
	ClusterID int64
}

func (ts *TSLog) GetDBType() string {
	generalConfig, err := contexthelper.GetGeneralConfig(ts.AppContext)
	if err != nil {
		return ""
	}

	return generalConfig.GetLogsDBType()
}

func (ts *TSLog) GetPGDB() (*sqlx.DB, error) {
	pgdbs, err := contexthelper.GetPGDBConfig(ts.AppContext)
	if err != nil {
		return nil, err
	}

	return pgdbs.GetTSLog(ts.ClusterID), nil
}

func (ts *TSLog) CreateFromJSON(clusterID int64, jsonData []byte, deletedFrom int64, ttl time.Duration) error {
	if ts.GetDBType() == "pg" {
		return pg.NewTSLog(ts.AppContext, ts.ClusterID).CreateFromJSON(nil, clusterID, jsonData, deletedFrom)

	} else if ts.GetDBType() == "cassandra" {
		return cassandra.NewTSLog(ts.AppContext).CreateFromJSON(clusterID, jsonData, ttl)
	}

	return fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}

// LastByClusterID returns the last row by cluster id.
func (ts *TSLog) LastByClusterID(clusterID int64) (shared.ICreatedUnix, error) {
	if ts.GetDBType() == "pg" {
		return pg.NewTSLog(ts.AppContext, ts.ClusterID).LastByClusterID(nil, clusterID)

	} else if ts.GetDBType() == "cassandra" {
		return cassandra.NewTSLog(ts.AppContext).LastByClusterID(clusterID)
	}

	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}

// AllByClusterIDRangeAndQuery returns all rows by cluster id, unix timestamp range, and resourced query.
func (ts *TSLog) AllByClusterIDRangeAndQuery(clusterID int64, from, to int64, resourcedQuery string, deletedFrom int64) (interface{}, error) {
	if ts.GetDBType() == "pg" {
		return pg.NewTSLog(ts.AppContext, ts.ClusterID).AllByClusterIDRangeAndQuery(nil, clusterID, from, to, resourcedQuery, deletedFrom)

	} else if ts.GetDBType() == "cassandra" {
		return cassandra.NewTSLog(ts.AppContext).AllByClusterIDRangeAndQuery(clusterID, from, to, resourcedQuery)
	}

	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}
