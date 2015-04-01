package dal

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	sqlx_types "github.com/jmoiron/sqlx/types"
	"github.com/resourced/resourced-master/dal/types"
	"strings"
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
	Tags          []string            `db:"tags"`
	Data          sqlx_types.JsonText `db:"data"`
}

type ResourcedPayload struct {
	Data     map[string]interface{}
	GoStruct string
	Host     struct {
		Name string
		Tags types.StringSlice
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

// GetById returns record by id.
func (h *Host) GetById(tx *sqlx.Tx, id int64) (*HostRow, error) {
	hostRow := &HostRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE id=$1", h.table)
	err := h.db.Get(hostRow, query, id)

	return hostRow, err
}

// GetByName returns record by name.
func (h *Host) GetByName(tx *sqlx.Tx, name string) (*HostRow, error) {
	hostRow := &HostRow{}
	query := fmt.Sprintf("SELECT * FROM %v WHERE name=$1", h.table)
	err := h.db.Get(hostRow, query, name)

	return hostRow, err
}

// Create performs insert for one host data.
func (h *Host) Create(tx *sqlx.Tx, accessTokenId int64, jsonData []byte) (*HostRow, error) {
	resourcedPayloads := make(map[string]*ResourcedPayload)

	err := json.Unmarshal(jsonData, &resourcedPayloads)
	if err != nil {
		return nil, err
	}

	// Get a random path from payload.
	var path string
	for path, _ = range resourcedPayloads {
		break
	}

	data := make(map[string]interface{})
	data["access_token_id"] = accessTokenId
	data["data"] = jsonData
	data["name"] = resourcedPayloads[path].Host.Name

	if len(resourcedPayloads[path].Host.Tags) > 0 {
		data["tags"] = fmt.Sprintf("ARRAY[%s]", strings.Join(resourcedPayloads[path].Host.Tags, ","))
	}

	sqlResult, err := h.InsertIntoTable(tx, data)
	if err != nil {
		return nil, err
	}

	return h.hostRowFromSqlResult(tx, sqlResult)
}

func (h *Host) CreateOrUpdate(tx *sqlx.Tx, accessTokenId int64, jsonData []byte) (*HostRow, error) {
	resourcedPayloads := make(map[string]*ResourcedPayload)

	err := json.Unmarshal(jsonData, &resourcedPayloads)
	if err != nil {
		return nil, err
	}

	// Get a random path from payload.
	var path string
	for path, _ = range resourcedPayloads {
		break
	}

	data := make(map[string]interface{})
	data["access_token_id"] = accessTokenId
	data["data"] = jsonData
	data["name"] = resourcedPayloads[path].Host.Name

	if len(resourcedPayloads[path].Host.Tags) > 0 {
		data["tags"] = fmt.Sprintf("ARRAY[%s]", strings.Join(resourcedPayloads[path].Host.Tags, ","))
	}

	hostRow, err := h.GetByName(tx, resourcedPayloads[path].Host.Name)

	// Perform INSERT
	if hostRow == nil || err != nil {
		sqlResult, err := h.InsertIntoTable(tx, data)
		if err != nil {
			return nil, err
		}

		return h.hostRowFromSqlResult(tx, sqlResult)
	}

	// Perform UPDATE
	_, err = h.UpdateByKeyValueString(tx, data, "name", resourcedPayloads[path].Host.Name)
	if err != nil {
		println(err.Error())
		return nil, err
	}

	return hostRow, nil
}

// AllHosts returns all user rows.
func (h *Host) AllHosts(tx *sqlx.Tx) ([]*HostRow, error) {
	hosts := []*HostRow{}
	query := fmt.Sprintf("SELECT id, name, UNNEST(tags), data FROM %v", h.table)
	err := h.db.Select(&hosts, query)

	return hosts, err
}
