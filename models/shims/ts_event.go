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

func NewTSEvent(ctx context.Context, clusterID int64) *TSEvent {
	ts := &TSEvent{}
	ts.AppContext = ctx
	ts.ClusterID = clusterID
	return ts
}

type TSEvent struct {
	Base
	ClusterID int64
}

func (ts *TSEvent) GetDBType() string {
	generalConfig, err := contexthelper.GetGeneralConfig(ts.AppContext)
	if err != nil {
		return ""
	}

	return generalConfig.GetEventsDBType()
}

func (ts *TSEvent) GetPGDB() (*sqlx.DB, error) {
	pgdbs, err := contexthelper.GetPGDBConfig(ts.AppContext)
	if err != nil {
		return nil, err
	}

	return pgdbs.GetTSEvent(ts.ClusterID), nil
}

func (ts *TSEvent) CreateFromJSON(id, clusterID int64, jsonData []byte, deletedFrom int64, ttl time.Duration) (interface{}, error) {
	if ts.GetDBType() == "pg" {
		return pg.NewTSEvent(ts.AppContext, ts.ClusterID).CreateFromJSON(nil, id, clusterID, jsonData, deletedFrom)

	} else if ts.GetDBType() == "cassandra" {
		return cassandra.NewTSEvent(ts.AppContext).CreateFromJSON(id, clusterID, jsonData, ttl)
	}

	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}

func (ts *TSEvent) AllLinesByClusterIDAndCreatedFromRangeForHighchart(clusterID, from, to, deletedFrom int64) ([]shared.TSEventHighchartLinePayload, error) {
	if ts.GetDBType() == "pg" {
		return pg.NewTSEvent(ts.AppContext, ts.ClusterID).AllLinesByClusterIDAndCreatedFromRangeForHighchart(nil, clusterID, from, to, deletedFrom)

	} else if ts.GetDBType() == "cassandra" {
		return cassandra.NewTSEvent(ts.AppContext).AllLinesByClusterIDAndCreatedFromRangeForHighchart(clusterID, from, to)
	}

	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}

func (ts *TSEvent) AllBandsByClusterIDAndCreatedFromRangeForHighchart(clusterID, from, to, deletedFrom int64) ([]shared.TSEventHighchartLinePayload, error) {
	if ts.GetDBType() == "pg" {
		return pg.NewTSEvent(ts.AppContext, ts.ClusterID).AllBandsByClusterIDAndCreatedFromRangeForHighchart(nil, clusterID, from, to, deletedFrom)

	} else if ts.GetDBType() == "cassandra" {
		return cassandra.NewTSEvent(ts.AppContext).AllBandsByClusterIDAndCreatedFromRangeForHighchart(clusterID, from, to)
	}

	return nil, fmt.Errorf("Unrecognized DBType, valid options are: pg or cassandra")
}
