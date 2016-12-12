package cassandra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/gocql/gocql"

	"github.com/resourced/resourced-master/contexthelper"
)

func NewHostData(ctx context.Context) *HostData {
	host := &HostData{}
	host.AppContext = ctx
	host.table = "hosts_data"

	return host
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
