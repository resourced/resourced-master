package cassandra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gocql/gocql"

	"github.com/resourced/resourced-master/contexthelper"
	"github.com/resourced/resourced-master/models/cassandra/querybuilder"
)

func NewHostData(ctx context.Context) *HostData {
	host := &HostData{}
	host.AppContext = ctx
	host.table = "hosts_data"

	return host
}

type HostDataRow struct {
	ID        string `db:"id" json:"-"`
	ClusterID int64  `db:"cluster_id"`
	Updated   int64  `db:"updated"`
	Key       string `db:"key"`
	Value     string `db:"value"`
}

type HostData struct {
	Base
}

func (h *HostData) GetCassandraSession() (*gocql.Session, error) {
	cassandradbs, err := contexthelper.GetCassandraDBConfig(h.AppContext)
	if err != nil {
		return nil, err
	}

	return cassandradbs.HostSession, nil
}

// AllByID returns all hosts data by id.
func (h *HostData) AllByID(id string) (map[string]interface{}, error) {
	session, err := h.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT key, value FROM %v WHERE id=?", h.table)

	result := make(map[string]interface{})

	var scannedKey, scannedValue string

	iter := session.Query(query, id).Iter()
	for iter.Scan(&scannedKey, &scannedValue) {
		valueFloat, err := strconv.ParseFloat(scannedValue, 64)
		if err == nil {
			result[scannedKey] = valueFloat
		} else {
			result[scannedKey] = scannedValue
		}
	}
	if err := iter.Close(); err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
		logrus.WithFields(logrus.Fields{"Method": "HostData.AllByID"}).Error(err)

		return nil, err
	}

	return result, nil
}

func (h *HostData) parseAgentResourcePayload(jsonData []byte) (AgentResourcePayload, error) {
	resourcedPayload := AgentResourcePayload{}

	err := json.Unmarshal(jsonData, &resourcedPayload)
	if err != nil {
		return resourcedPayload, err
	}

	return resourcedPayload, nil
}

// CreateOrUpdate performs insert/update for one host data.
func (h *HostData) CreateOrUpdate(accessTokenRow *AccessTokenRow, jsonData []byte) (map[string]string, error) {
	resourcedPayload, err := h.parseAgentResourcePayload(jsonData)
	if err != nil {
		return nil, err
	}

	if resourcedPayload.Host.Name == "" {
		return nil, errors.New("Hostname cannot be empty.")
	}

	session, err := h.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	id := resourcedPayload.Host.Name

	query := fmt.Sprintf("INSERT INTO %v (id, key, value) VALUES (?, ?, ?)", h.table)

	for key, value := range resourcedPayload.Data {
		err = session.Query(query, id, key, value).Exec()
		if err != nil {
			return nil, err
		}
	}

	return resourcedPayload.Data, nil
}

// AllByClusterIDQueryAndUpdatedInterval returns all rows by resourced query.
func (h *HostData) AllByClusterIDQueryAndUpdatedInterval(clusterID int64, resourcedQuery, updatedInterval string) (map[string][]*HostDataRow, error) {
	session, err := h.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	updatedDuration, err := time.ParseDuration(updatedInterval)
	if err != nil {
		return nil, err
	}

	updated := time.Now().UTC().Add(-1 * updatedDuration)
	updatedUnix := updated.UTC().Unix()

	var query string

	luceneQuery := querybuilder.Parse(resourcedQuery, nil)

	if luceneQuery == "" {
		query = fmt.Sprintf(`SELECT id, cluster_id, access_token_id, hostname, updated, tags, master_tags FROM %v WHERE expr(idx_hosts_lucene, '{
    filter: {
        type: "boolean",
        must: [
            {type: "match", field: "cluster_id", value: %v},
            {type:"range", field:"updated", lower:%v, include_lower: true}
        ]
    }
}')`, h.table, clusterID, updatedUnix)

	} else {
		query = fmt.Sprintf(`SELECT id, cluster_id, access_token_id, hostname, updated, tags, master_tags FROM %v WHERE expr(idx_hosts_lucene, '{
    filter: {
        type: "boolean",
        must: [
            {type: "match", field: "cluster_id", value: %v},
            {type:"range", field:"updated", lower:%v, include_lower: true},
            %v
        ]
    }
}')`, h.table, clusterID, updatedUnix, luceneQuery)

	}

	println(query)

	result := make(map[string][]*HostDataRow)

	var scannedClusterID, scannedUpdated int64
	var scannedID, scannedKey, scannedValue string

	iter := session.Query(query).Iter()
	for iter.Scan(&scannedID, &scannedClusterID, &scannedUpdated, &scannedKey, &scannedValue) {
		hostData := &HostDataRow{
			ID:        scannedID,
			ClusterID: scannedClusterID,
			Updated:   scannedUpdated,
			Key:       scannedKey,
			Value:     scannedValue,
		}

		_, ok := result[scannedID]
		if !ok {
			result[scannedID] = make([]*HostDataRow, 0)
		}

		result[scannedID] = append(result[scannedID], hostData)
	}
	if err := iter.Close(); err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
		logrus.WithFields(logrus.Fields{"Method": "HostData.AllByClusterIDQueryAndUpdatedInterval"}).Error(err)

		return nil, err
	}

	return result, err
}

func (h *HostData) AllByClusterIDAndHostnames(clusterID int64, hostnames []string) (map[string][]*HostDataRow, error) {
	return nil, nil
}
