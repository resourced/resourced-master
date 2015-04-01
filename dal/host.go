package dal

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	sqlx_types "github.com/jmoiron/sqlx/types"
)

func NewHost(db *sqlx.DB) *Host {
	host := &Host{}
	host.db = db
	host.table = "hosts"
	host.hasID = true

	return host
}

type HostRow struct {
	ID            int64               `db:"id"`
	AccessTokenID int64               `db:"access_token_id"`
	Name          string              `db:"name"`
	Tags          sqlx_types.JsonText `db:"tags"`
	Data          sqlx_types.JsonText `db:"data"`
}

type ResourcedPayload struct {
	Data     map[string]interface{}
	GoStruct string
	Host     struct {
		Name string
		Tags []string
	}
	Interval string
	Path     string
	UnixNano float64
}

type Host struct {
	Base
}

func (h *Host) hostRowFromSqlResult(tx *sqlx.Tx, sqlResult sql.Result) (*HostRow, error) {
	hostId, err := sqlResult.LastInsertId()
	if err != nil {
		return nil, err
	}

	return h.GetById(tx, hostId)
}

// AllHosts returns all user rows.
func (h *Host) AllHosts(tx *sqlx.Tx) ([]*HostRow, error) {
	hosts := []*HostRow{}
	query := fmt.Sprintf("SELECT id, name, tags, data FROM %v", h.table)
	err := h.db.Select(&hosts, query)

	return hosts, err
}

// GetById returns record by id.
func (h *Host) GetById(tx *sqlx.Tx, id int64) (*HostRow, error) {
	hostRow := &HostRow{}
	query := fmt.Sprintf("SELECT id, name, tags, data FROM %v WHERE id=$1", h.table)
	err := h.db.Get(hostRow, query, id)

	return hostRow, err
}

// GetByName returns record by name.
func (h *Host) GetByName(tx *sqlx.Tx, name string) (*HostRow, error) {
	hostRow := &HostRow{}
	query := fmt.Sprintf("SELECT id, name, tags, data FROM %v WHERE name=$1", h.table)
	err := h.db.Get(hostRow, query, name)

	return hostRow, err
}

func (h *Host) parseResourcedPayload(tx *sqlx.Tx, accessTokenId int64, jsonData []byte) (map[string]interface{}, error) {
	resourcedPayloads := make(map[string]*ResourcedPayload)

	err := json.Unmarshal(jsonData, &resourcedPayloads)
	if err != nil {
		return nil, err
	}

	resourcedPayloadJustData := make(map[string]map[string]interface{})

	data := make(map[string]interface{})
	data["access_token_id"] = accessTokenId

	for path, resourcedPayload := range resourcedPayloads {
		data["name"] = resourcedPayload.Host.Name

		tagsInJson, err := json.Marshal(resourcedPayload.Host.Tags)
		if err != nil {
			continue
		}
		data["tags"] = tagsInJson

		resourcedPayloadJustData[path] = resourcedPayload.Data
	}

	resourcedPayloadJustJson, err := json.Marshal(resourcedPayloadJustData)
	if err != nil {
		return nil, err
	}

	data["data"] = resourcedPayloadJustJson

	return data, nil
}

// Create performs insert for one host data.
func (h *Host) Create(tx *sqlx.Tx, accessTokenId int64, jsonData []byte) (*HostRow, error) {
	data, err := h.parseResourcedPayload(tx, accessTokenId, jsonData)
	if err != nil {
		return nil, err
	}

	sqlResult, err := h.InsertIntoTable(tx, data)
	if err != nil {
		return nil, err
	}

	return h.hostRowFromSqlResult(tx, sqlResult)
}

func (h *Host) CreateOrUpdate(tx *sqlx.Tx, accessTokenId int64, jsonData []byte) (*HostRow, error) {
	data, err := h.parseResourcedPayload(tx, accessTokenId, jsonData)
	if err != nil {
		return nil, err
	}

	hostRow, err := h.GetByName(tx, data["name"].(string))

	// Perform INSERT
	if hostRow == nil || err != nil {
		sqlResult, err := h.InsertIntoTable(tx, data)
		if err != nil {
			return nil, err
		}

		return h.hostRowFromSqlResult(tx, sqlResult)
	}

	// Perform UPDATE
	_, err = h.UpdateByKeyValueString(tx, data, "name", data["name"].(string))
	if err != nil {
		return nil, err
	}

	return hostRow, nil
}
