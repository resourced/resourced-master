package dal

import (
	"database/sql"
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
	Tags          []string            `db:"tags"`
	Data          sqlx_types.JsonText `db:"data"`
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

// CreateRow performs insert for one host data.
func (h *Host) CreateRow(tx *sqlx.Tx, accessTokenId int64, jsonData []byte) (*HostRow, error) {
	data := make(map[string]interface{})
	data["access_token_id"] = accessTokenId
	data["name"] = ""
	data["data"] = jsonData

	sqlResult, err := h.InsertIntoTable(tx, data)
	if err != nil {
		return nil, err
	}

	return h.hostRowFromSqlResult(tx, sqlResult)
}
