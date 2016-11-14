package pg

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	sqlx_types "github.com/jmoiron/sqlx/types"
	"github.com/nytlabs/gojsonexplode"

	"github.com/resourced/resourced-master/contexthelper"
	"github.com/resourced/resourced-master/models/cassandra"
	"github.com/resourced/resourced-master/models/pg/querybuilder"
)

func NewHost(ctx context.Context, clusterID int64) *Host {
	host := &Host{}
	host.AppContext = ctx
	host.table = "hosts"
	host.hasID = true
	host.clusterID = clusterID
	host.i = host

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
	ID            int64               `db:"id" json:"-"`
	AccessTokenID int64               `db:"access_token_id" json:"-"`
	ClusterID     int64               `db:"cluster_id"`
	Hostname      string              `db:"hostname"`
	Updated       time.Time           `db:"updated"`
	Tags          sqlx_types.JSONText `db:"tags" json:",omitempty"`
	MasterTags    sqlx_types.JSONText `db:"master_tags" json:",omitempty"`
	Data          sqlx_types.JSONText `db:"data" json:",omitempty"`
}

func (h *HostRow) GetTags() map[string]string {
	tags := make(map[string]string)
	h.Tags.Unmarshal(&tags)

	return tags
}

func (h *HostRow) GetMasterTags() map[string]interface{} {
	tags := make(map[string]interface{})
	h.MasterTags.Unmarshal(&tags)

	return tags
}

func (h *HostRow) GetClusterID() int64 {
	return h.ClusterID
}

func (h *HostRow) GetHostname() string {
	return h.Hostname
}

func (h *HostRow) DataAsFlatKeyValue() map[string]map[string]interface{} {
	inputData := make(map[string]map[string]interface{})

	outputData := make(map[string]map[string]interface{})

	err := json.Unmarshal(h.Data, &inputData)
	if err != nil {
		return outputData
	}

	for path, innerData := range inputData {
		innerDataJson, err := json.Marshal(innerData)
		if err != nil {
			continue
		}

		flattenMapJson, err := gojsonexplode.Explodejson(innerDataJson, ".")
		if err != nil {
			continue
		}

		flattenMap := make(map[string]interface{})

		err = json.Unmarshal(flattenMapJson, &flattenMap)
		if err != nil {
			continue
		}
		outputData[path] = flattenMap
	}

	return outputData
}

type Host struct {
	Base
	clusterID int64
}

func (h *Host) GetPGDB() (*sqlx.DB, error) {
	pgdbs, err := contexthelper.GetPGDBConfig(h.AppContext)
	if err != nil {
		return nil, err
	}
	if pgdbs == nil {
		return nil, fmt.Errorf("Database handler went missing")
	}

	return pgdbs.GetHost(h.clusterID), nil
}

func (h *Host) hostRowFromSqlResult(tx *sqlx.Tx, sqlResult sql.Result) (*HostRow, error) {
	hostId, err := sqlResult.LastInsertId()
	if err != nil {
		return nil, err
	}

	return h.GetByID(tx, hostId)
}

// AllByClusterID returns all rows.
func (h *Host) AllByClusterID(tx *sqlx.Tx, clusterID int64) ([]*HostRow, error) {
	pgdb, err := h.GetPGDB()
	if err != nil {
		return nil, err
	}

	hosts := []*HostRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 ORDER BY updated DESC", h.table)
	err = pgdb.Select(&hosts, query, clusterID)

	return hosts, err
}

// AllByClusterIDAndUpdatedInterval returns all rows.
func (h *Host) AllByClusterIDAndUpdatedInterval(tx *sqlx.Tx, clusterID int64, updatedInterval string) ([]*HostRow, error) {
	pgdb, err := h.GetPGDB()
	if err != nil {
		return nil, err
	}

	hosts := []*HostRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND updated >= (NOW() at time zone 'utc' - INTERVAL '%v')", h.table, updatedInterval)
	err = pgdb.Select(&hosts, query, clusterID)

	return hosts, err
}

// AllCompactByClusterIDQueryAndUpdatedInterval returns all rows by resourced query.
func (h *Host) AllCompactByClusterIDQueryAndUpdatedInterval(tx *sqlx.Tx, clusterID int64, resourcedQuery, updatedInterval string) ([]*HostRow, error) {
	pgdb, err := h.GetPGDB()
	if err != nil {
		return nil, err
	}

	pgQuery := querybuilder.Parse(resourcedQuery)
	if pgQuery == "" {
		return h.AllByClusterID(tx, clusterID)
	}

	hosts := []*HostRow{}
	query := fmt.Sprintf("SELECT id, cluster_id, access_token_id, hostname, updated, tags, master_tags FROM %v WHERE cluster_id=$1 AND updated >= (NOW() at time zone 'utc' - INTERVAL '%v') AND %v", h.table, updatedInterval, pgQuery)

	err = pgdb.Select(&hosts, query, clusterID)

	return hosts, err
}

// AllByClusterIDQueryAndUpdatedInterval returns all rows by resourced query.
func (h *Host) AllByClusterIDQueryAndUpdatedInterval(tx *sqlx.Tx, clusterID int64, resourcedQuery, updatedInterval string) ([]*HostRow, error) {
	pgdb, err := h.GetPGDB()
	if err != nil {
		return nil, err
	}

	pgQuery := querybuilder.Parse(resourcedQuery)
	if pgQuery == "" {
		return h.AllByClusterIDAndUpdatedInterval(tx, clusterID, updatedInterval)
	}

	hosts := []*HostRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND updated >= (NOW() at time zone 'utc' - INTERVAL '%v') AND %v", h.table, updatedInterval, pgQuery)
	err = pgdb.Select(&hosts, query, clusterID)

	return hosts, err
}

// AllByClusterIDAndHostnames returns all rows by hostnames.
func (h *Host) AllByClusterIDAndHostnames(tx *sqlx.Tx, clusterID int64, hostnames []string) ([]*HostRow, error) {
	pgdb, err := h.GetPGDB()
	if err != nil {
		return nil, err
	}

	inPlaceHolders := make([]string, len(hostnames))

	for i := 0; i < len(hostnames); i++ {
		inPlaceHolders[i] = fmt.Sprintf("$%v", i+2)
	}

	hosts := []*HostRow{}

	query := fmt.Sprintf("SELECT * FROM %v WHERE cluster_id=$1 AND hostname IN (%v)", h.table, strings.Join(inPlaceHolders, ","))

	args := make([]interface{}, len(hostnames)+1)
	args[0] = clusterID

	for i := 1; i < len(hostnames)+1; i++ {
		args[i] = hostnames[i-1]
	}

	err = pgdb.Select(&hosts, query, args...)

	return hosts, err
}

// GetByID returns record by id.
func (h *Host) GetByID(tx *sqlx.Tx, id int64) (*HostRow, error) {
	pgdb, err := h.GetPGDB()
	if err != nil {
		return nil, err
	}

	hostRow := &HostRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE id=$1", h.table)
	err = pgdb.Get(hostRow, query, id)

	return hostRow, err
}

// GetByHostname returns record by hostname.
func (h *Host) GetByHostname(tx *sqlx.Tx, hostname string) (*HostRow, error) {
	pgdb, err := h.GetPGDB()
	if err != nil {
		return nil, err
	}

	hostRow := &HostRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE hostname=$1", h.table)
	err = pgdb.Get(hostRow, query, hostname)

	return hostRow, err
}

func (h *Host) parseAgentResourcePayload(tx *sqlx.Tx, accessTokenRow *cassandra.AccessTokenRow, jsonData []byte) (map[string]interface{}, error) {
	resourcedPayload := AgentResourcePayload{}

	err := json.Unmarshal(jsonData, &resourcedPayload)
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	data["access_token_id"] = accessTokenRow.ID
	data["cluster_id"] = accessTokenRow.ClusterID
	data["hostname"] = resourcedPayload.Host.Name

	tagsInJson, err := json.Marshal(resourcedPayload.Host.Tags)
	if err != nil {
		return nil, err
	}
	data["tags"] = tagsInJson

	resourcedPayloadJustJson, err := json.Marshal(resourcedPayload.Data)
	if err != nil {
		return nil, err
	}

	data["data"] = resourcedPayloadJustJson

	return data, nil
}

// CreateOrUpdate performs insert/update for one host data.
func (h *Host) CreateOrUpdate(tx *sqlx.Tx, accessTokenRow *cassandra.AccessTokenRow, jsonData []byte) (*HostRow, error) {
	data, err := h.parseAgentResourcePayload(tx, accessTokenRow, jsonData)
	if err != nil {
		return nil, err
	}

	if data["hostname"] == nil {
		return nil, errors.New("Hostname cannot be empty.")
	}

	hostRow, err := h.GetByHostname(tx, data["hostname"].(string))

	// Perform INSERT
	if hostRow == nil || err != nil {
		sqlResult, err := h.InsertIntoTable(tx, data)
		if err != nil {
			return nil, err
		}

		return h.hostRowFromSqlResult(tx, sqlResult)
	}

	if _, ok := data["updated"]; !ok {
		data["updated"] = time.Now().UTC()
	}

	// Perform UPDATE
	_, err = h.UpdateByKeyValueString(tx, data, "hostname", data["hostname"].(string))
	if err != nil {
		return nil, err
	}

	return hostRow, nil
}

// UpdateMasterTagsByID updates master tags by ID.
func (h *Host) UpdateMasterTagsByID(tx *sqlx.Tx, id int64, tags map[string]interface{}) error {
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	data["master_tags"] = tagsJSON

	_, err = h.UpdateByID(tx, data, id)
	return err
}

// UpdateMasterTagsByHostname updates master tags by hostname.
func (h *Host) UpdateMasterTagsByHostname(tx *sqlx.Tx, hostname string, tags map[string]interface{}) error {
	hostRow, err := h.GetByHostname(tx, hostname)
	if err != nil {
		return err
	}

	return h.UpdateMasterTagsByID(tx, hostRow.ID, tags)
}
