package datastore

import (
	"database/sql"
	"time"

	"github.com/gregory90/go-file/model"
)

const (
	table string = "files"

	selectQuery string = `
        SELECT 
            lower(hex(uid)), 
            name, 
            mime, 
            uniqueID, 
            fileType, 
            tmp, 
            createdAt 
        FROM ` + table + " "

	deleteQuery string = `
        DELETE FROM  ` + table + ` `
)

func scanSelect(m *model.File, rows *sql.Rows) error {
	err := rows.Scan(
		&m.UID,
		&m.Name,
		&m.Mime,
		&m.UniqueID,
		&m.Type,
		&m.Tmp,
		&m.CreatedAt,
	)
	return err
}

func getAll(rows *sql.Rows) ([]model.File, error) {
	rs := []model.File{}
	defer rows.Close()
	r := &model.File{}
	for rows.Next() {
		err := scanSelect(r, rows)
		if err != nil {
			return nil, err
		}
		rs = append(rs, *r)
	}
	err := rows.Err()

	if err != nil {
		return nil, err
	}

	return rs, nil
}

func Get(tx *sql.Tx, uid string) (*model.File, error) {
	m := &model.File{}
	err := tx.QueryRow(selectQuery+"WHERE uid = unhex(?)", uid).Scan(&m.UID, &m.Name, &m.Mime, &m.UniqueID, &m.Type, &m.Tmp, &m.CreatedAt)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func GetOlderThan(tx *sql.Tx, t int) ([]model.File, error) {
	now := time.Now().UTC().Add(-1 * time.Minute * time.Duration(t))
	rows, err := tx.Query(selectQuery+"WHERE createdAt < ? AND tmp = true ORDER BY createdAt DESC", now)
	if err != nil {
		return nil, err
	}

	rs, err := getAll(rows)
	if err != nil {
		return nil, err
	}
	return rs, nil
}

func Create(tx *sql.Tx, m *model.File) error {
	stmt, err := tx.Prepare("INSERT " + table + " SET uid=unhex(?),name=?,mime=?,uniqueID=?,fileType=?,tmp=?,createdAt=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(m.UID, m.Name, m.Mime, m.UniqueID, m.Type, m.Tmp, m.CreatedAt)
	return err
}

func Update(tx *sql.Tx, m *model.File) error {
	stmt, err := tx.Prepare("UPDATE " + table + " SET name=?,tmp=? WHERE uid=unhex(?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(m.Name, m.Tmp, m.UID)
	return err
}

func Delete(tx *sql.Tx, uid string) error {
	stmt, err := tx.Prepare(deleteQuery + "WHERE uid=unhex(?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(uid)
	if err != nil {
		return err
	}

	return nil
}
