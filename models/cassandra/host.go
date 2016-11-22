package cassandra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gocql/gocql"

	"github.com/resourced/resourced-master/contexthelper"
	"github.com/resourced/resourced-master/models/pg/querybuilder"
)

func NewHost(ctx context.Context) *Host {
	host := &Host{}
	host.AppContext = ctx
	host.table = "hosts"

	return host
}

type AgentResourcePayload struct {
	Data map[string]interface{}
	Host struct {
		Name string
		Tags map[string]string
	}
}

type HostRowsWithError struct {
	Hosts []*HostRow
	Error error
}

type HostRow struct {
	ID            string            `db:"id" json:"-"`
	AccessTokenID int64             `db:"access_token_id" json:"-"`
	ClusterID     int64             `db:"cluster_id"`
	Hostname      string            `db:"hostname"`
	Updated       int64             `db:"updated"`
	Tags          map[string]string `db:"tags" json:",omitempty"`
	MasterTags    map[string]string `db:"master_tags" json:",omitempty"`
	Data          map[string]string `db:"data" json:",omitempty"`
}

func (h *HostRow) GetClusterID() int64 {
	return h.ClusterID
}

func (h *HostRow) GetHostname() string {
	return h.Hostname
}

func (h *HostRow) DataAsFlatKeyValue() map[string]map[string]interface{} {
	outputData := make(map[string]map[string]interface{})

	for path, value := range h.Data {
		pathChunks := strings.Split(path, ".")

		if len(pathChunks) > 0 {
			innerPath := strings.Replace(path, pathChunks[0]+".", "", 1)

			innerOutputData, ok := outputData[pathChunks[0]]
			if !ok {
				innerOutputData = make(map[string]interface{})
			}

			valueFloat, err := strconv.ParseFloat(value, 64)
			if err == nil {
				innerOutputData[innerPath] = valueFloat
			} else {
				innerOutputData[innerPath] = value
			}

			outputData[pathChunks[0]] = innerOutputData
		}
	}

	return outputData
}

type Host struct {
	Base
}

func (h *Host) GetCassandraSession() (*gocql.Session, error) {
	cassandradbs, err := contexthelper.GetCassandraDBConfig(h.AppContext)
	if err != nil {
		return nil, err
	}

	return cassandradbs.HostSession, nil
}

// AllByClusterID returns all rows by cluster_id.
func (h *Host) AllByClusterID(clusterID int64) ([]*HostRow, error) {
	session, err := h.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	rows := []*HostRow{}

	query := fmt.Sprintf(`SELECT id, cluster_id, access_token_id, hostname, updated, tags, master_tags, data FROM %v WHERE cluster_id=? ALLOW FILTERING`, h.table)

	var scannedClusterID, scannedAccessTokenID, scannedUpdated int64
	var scannedID, scannedHostname string
	var scannedTags, scannedMasterTags, scannedData map[string]string

	iter := session.Query(query, clusterID).Iter()
	for iter.Scan(&scannedID, &scannedClusterID, &scannedAccessTokenID, &scannedHostname, &scannedUpdated, &scannedTags, &scannedMasterTags, &scannedData) {
		rows = append(rows, &HostRow{
			ID:            scannedID,
			ClusterID:     scannedClusterID,
			AccessTokenID: scannedAccessTokenID,
			Hostname:      scannedHostname,
			Updated:       scannedUpdated,
			Tags:          scannedTags,
			MasterTags:    scannedMasterTags,
			Data:          scannedData,
		})
	}
	if err := iter.Close(); err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
		logrus.WithFields(logrus.Fields{"Method": "Host.AllByClusterID"}).Error(err)

		return nil, err
	}
	return rows, err
}

// AllByClusterIDAndUpdatedInterval returns all rows.
func (h *Host) AllByClusterIDAndUpdatedInterval(clusterID int64, updatedInterval string) ([]*HostRow, error) {
	session, err := h.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	rows := []*HostRow{}

	query := fmt.Sprintf(`SELECT id, cluster_id, access_token_id, hostname, updated, tags, master_tags, data FROM %v WHERE cluster_id=? AND updated >= ? ALLOW FILTERING`, h.table)

	var scannedClusterID, scannedAccessTokenID, scannedUpdated int64
	var scannedID, scannedHostname string
	var scannedTags, scannedMasterTags, scannedData map[string]string

	// TODO: change 1 based on updatedInt
	// old: 	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND updated >= (NOW() at time zone 'utc' - INTERVAL '%v')", h.table, updatedInterval)
	iter := session.Query(query, clusterID, 1).Iter()
	for iter.Scan(&scannedID, &scannedClusterID, &scannedAccessTokenID, &scannedHostname, &scannedUpdated, &scannedTags, &scannedMasterTags, &scannedData) {
		rows = append(rows, &HostRow{
			ID:            scannedID,
			ClusterID:     scannedClusterID,
			AccessTokenID: scannedAccessTokenID,
			Hostname:      scannedHostname,
			Updated:       scannedUpdated,
			Tags:          scannedTags,
			MasterTags:    scannedMasterTags,
			Data:          scannedData,
		})
	}
	if err := iter.Close(); err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
		logrus.WithFields(logrus.Fields{"Method": "Host.AllByClusterID"}).Error(err)

		return nil, err
	}
	return rows, err
}

// AllCompactByClusterIDQueryAndUpdatedInterval returns all rows by resourced query.
func (h *Host) AllCompactByClusterIDQueryAndUpdatedInterval(clusterID int64, resourcedQuery, updatedInterval string) ([]*HostRow, error) {
	session, err := h.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	luceneQuery := querybuilder.Parse(resourcedQuery)
	if luceneQuery == "" {
		return h.AllByClusterID(clusterID)
	}

	rows := []*HostRow{}

	query := fmt.Sprintf(`SELECT id, cluster_id, access_token_id, hostname, updated, tags, master_tags FROM %v WHERE cluster_id=? AND updated >= ? AND %v ALLOW FILTERING`, h.table, luceneQuery)

	var scannedClusterID, scannedAccessTokenID, scannedUpdated int64
	var scannedID, scannedHostname string
	var scannedTags, scannedMasterTags map[string]string

	// TODO: change 1 based on updatedInt
	// old: 	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND updated >= (NOW() at time zone 'utc' - INTERVAL '%v')", h.table, updatedInterval)
	iter := session.Query(query, clusterID, 1).Iter()
	for iter.Scan(&scannedID, &scannedClusterID, &scannedAccessTokenID, &scannedHostname, &scannedUpdated, &scannedTags, &scannedMasterTags) {
		rows = append(rows, &HostRow{
			ID:            scannedID,
			ClusterID:     scannedClusterID,
			AccessTokenID: scannedAccessTokenID,
			Hostname:      scannedHostname,
			Updated:       scannedUpdated,
			Tags:          scannedTags,
			MasterTags:    scannedMasterTags,
		})
	}
	if err := iter.Close(); err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
		logrus.WithFields(logrus.Fields{"Method": "Host.AllCompactByClusterIDQueryAndUpdatedInterval"}).Error(err)

		return nil, err
	}
	return rows, err
}

// AllByClusterIDQueryAndUpdatedInterval returns all rows by resourced query.
func (h *Host) AllByClusterIDQueryAndUpdatedInterval(clusterID int64, resourcedQuery, updatedInterval string) ([]*HostRow, error) {
	session, err := h.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	luceneQuery := querybuilder.Parse(resourcedQuery)
	if luceneQuery == "" {
		return h.AllByClusterID(clusterID)
	}

	rows := []*HostRow{}

	query := fmt.Sprintf(`SELECT id, cluster_id, access_token_id, hostname, updated, tags, master_tags, data FROM %v WHERE cluster_id=? AND updated >= ? AND %v ALLOW FILTERING`, h.table, luceneQuery)

	var scannedClusterID, scannedAccessTokenID, scannedUpdated int64
	var scannedID, scannedHostname string
	var scannedTags, scannedMasterTags, scannedData map[string]string

	// TODO: change 1 based on updatedInt
	// old: 	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND updated >= (NOW() at time zone 'utc' - INTERVAL '%v')", h.table, updatedInterval)
	iter := session.Query(query, clusterID, 1).Iter()
	for iter.Scan(&scannedID, &scannedClusterID, &scannedAccessTokenID, &scannedHostname, &scannedUpdated, &scannedTags, &scannedMasterTags, &scannedData) {
		rows = append(rows, &HostRow{
			ID:            scannedID,
			ClusterID:     scannedClusterID,
			AccessTokenID: scannedAccessTokenID,
			Hostname:      scannedHostname,
			Updated:       scannedUpdated,
			Tags:          scannedTags,
			MasterTags:    scannedMasterTags,
			Data:          scannedData,
		})
	}
	if err := iter.Close(); err != nil {
		err = fmt.Errorf("%v. Query: %v", err.Error(), query)
		logrus.WithFields(logrus.Fields{"Method": "Host.AllByClusterIDQueryAndUpdatedInterval"}).Error(err)

		return nil, err
	}
	return rows, err
}

// AllByClusterIDAndHostnames returns all rows by hostnames.
// TODO: finish this.
func (h *Host) AllByClusterIDAndHostnames(clusterID int64, hostnames []string) ([]*HostRow, error) {
	// session, err := h.GetCassandraSession()
	// if err != nil {
	// 	return nil, err
	// }

	// inPlaceHolders := make([]string, len(hostnames))

	// for i := 0; i < len(hostnames); i++ {
	// 	inPlaceHolders[i] = fmt.Sprintf("$%v", i+2)
	// }

	// hosts := []*HostRow{}

	// query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND hostname IN (%v)", h.table, strings.Join(inPlaceHolders, ","))

	// args := make([]interface{}, len(hostnames)+1)
	// args[0] = clusterID

	// for i := 1; i < len(hostnames)+1; i++ {
	// 	args[i] = hostnames[i-1]
	// }

	// err = pgdb.Select(&hosts, query, args...)

	// return hosts, err
	return nil, nil
}

// GetByID returns record by id.
func (h *Host) GetByID(id int64) (*HostRow, error) {
	session, err := h.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT id, cluster_id, access_token_id, hostname, updated, tags, master_tags, data FROM %v WHERE id=?", h.table)

	var scannedClusterID, scannedAccessTokenID, scannedUpdated int64
	var scannedID, scannedHostname string
	var scannedTags, scannedMasterTags, scannedData map[string]string

	err = session.Query(query, id).Scan(&scannedID, &scannedClusterID, &scannedAccessTokenID, &scannedHostname, &scannedUpdated, &scannedTags, &scannedMasterTags, &scannedData)
	if err != nil {
		return nil, err
	}

	row := &HostRow{
		ID:            scannedID,
		ClusterID:     scannedClusterID,
		AccessTokenID: scannedAccessTokenID,
		Hostname:      scannedHostname,
		Updated:       scannedUpdated,
		Tags:          scannedTags,
		MasterTags:    scannedMasterTags,
		Data:          scannedData,
	}

	return row, err
}

// GetByHostname returns record by hostname.
func (h *Host) GetByHostname(hostname string) (*HostRow, error) {
	session, err := h.GetCassandraSession()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT id, cluster_id, access_token_id, hostname, updated, tags, master_tags, data FROM %v WHERE hostname=?", h.table)

	var scannedClusterID, scannedAccessTokenID, scannedUpdated int64
	var scannedID, scannedHostname string
	var scannedTags, scannedMasterTags, scannedData map[string]string

	err = session.Query(query, hostname).Scan(&scannedID, &scannedClusterID, &scannedAccessTokenID, &scannedHostname, &scannedUpdated, &scannedTags, &scannedMasterTags, &scannedData)
	if err != nil {
		return nil, err
	}

	row := &HostRow{
		ID:            scannedID,
		ClusterID:     scannedClusterID,
		AccessTokenID: scannedAccessTokenID,
		Hostname:      scannedHostname,
		Updated:       scannedUpdated,
		Tags:          scannedTags,
		MasterTags:    scannedMasterTags,
		Data:          scannedData,
	}

	return row, err
}

func (h *Host) parseAgentResourcePayload(jsonData []byte) (AgentResourcePayload, error) {
	resourcedPayload := AgentResourcePayload{}

	err := json.Unmarshal(jsonData, &resourcedPayload)
	if err != nil {
		return resourcedPayload, err
	}

	return resourcedPayload, nil

	// data := make(map[string]interface{})
	// data["access_token_id"] = accessTokenRow.ID
	// data["cluster_id"] = accessTokenRow.ClusterID
	// data["hostname"] = resourcedPayload.Host.Name

	// tagsInJson, err := json.Marshal(resourcedPayload.Host.Tags)
	// if err != nil {
	// 	return nil, err
	// }
	// data["tags"] = tagsInJson

	// resourcedPayloadJustJson, err := json.Marshal(resourcedPayload.Data)
	// if err != nil {
	// 	return nil, err
	// }

	// data["data"] = resourcedPayloadJustJson

	// return data, nil
}

// CreateOrUpdate performs insert/update for one host data.
// TODO finish this
func (h *Host) CreateOrUpdate(accessTokenRow *AccessTokenRow, jsonData []byte) (*HostRow, error) {
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
	updated := time.Now().UTC().Unix()

	query := fmt.Sprintf("INSERT INTO %v (id, cluster_id, access_token_id, hostname, updated, tags, data) VALUES (?, ?, ?, ?, ?, ?, ?)", h.table)

	err = session.Query(query, id, accessTokenRow.ClusterID, accessTokenRow.ID, resourcedPayload.Host.Name, updated, resourcedPayload.Host.Tags, resourcedPayload.Data).Exec()
	if err != nil {
		return nil, err
	}

	return &HostRow{
		ID:            id,
		ClusterID:     accessTokenRow.ClusterID,
		AccessTokenID: accessTokenRow.ID,
		Hostname:      resourcedPayload.Host.Name,
		Updated:       updated,
		Tags:          resourcedPayload.Host.Tags,
	}, nil
}

// UpdateMasterTagsByID updates master tags by ID.
func (h *Host) UpdateMasterTagsByID(id int64, tags map[string]string) error {
	session, err := h.GetCassandraSession()
	if err != nil {
		return err
	}

	query := fmt.Sprintf("UPDATE %v SET tags=? WHERE id=?", h.table)

	return session.Query(query, tags, id).Exec()
}

// UpdateMasterTagsByHostname updates master tags by hostname.
func (h *Host) UpdateMasterTagsByHostname(hostname string, tags map[string]string) error {
	session, err := h.GetCassandraSession()
	if err != nil {
		return err
	}

	query := fmt.Sprintf("UPDATE %v SET tags=? WHERE hostname=?", h.table)

	return session.Query(query, tags, hostname).Exec()
}
