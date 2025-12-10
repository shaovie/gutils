package isqlite

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path"
	"sync"
)

type DBConn struct {
	file string
	db   *sql.DB
	mtx  sync.RWMutex // just for sqlite3
}

func New(f, pragma string) (*DBConn, error) {
	if f == "" {
		return nil, errors.New("sqlite filename is empty")
	}
	if err := os.MkdirAll(path.Dir(f), 0755); err != nil {
		return nil, errors.New("mkdir sqlite dir failed! " + err.Error())
	}
	dbR, err := sql.Open("sqlite3", f)
	if err != nil {
		return nil, errors.New("sqlite open failed! " + err.Error())
	}
	if pragma != "" {
		_, err = dbR.Exec(pragma)
		if err != nil {
			return nil, errors.New("sqlite PRAGMA set fail:" + err.Error())
		}
	}
	return &DBConn{
		db:   dbR,
		file: f,
	}, nil
}
func (c *DBConn) Insert(sql string, vals ...any) (int64, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	stmt, err := c.db.Prepare(sql)
	if err != nil {
		return 0, errors.New(sql + " prepare failed! " + err.Error())
	}
	defer stmt.Close()
	res, err := stmt.Exec(vals...)
	if err != nil {
		return 0, errors.New(sql + " exec failed! " + err.Error())
	}
	return res.LastInsertId()
}
func (c *DBConn) Exec(sql string, vals ...any) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	stmt, err := c.db.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(vals...)
	if err != nil {
		return err
	}
	return nil
}
func (c *DBConn) Query(query string, vals ...any) ([]map[string]string, error) {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	stmt, err := c.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(vals...)
	if err != nil {
		return nil, err
	}
	cols, _ := rows.Columns()
	results := make([]map[string]string, 0)
	for rows.Next() {
		values := make([]sql.RawBytes, len(cols))
		scans := make([]any, len(cols))
		for i := range values {
			scans[i] = &values[i]
		}
		if err = rows.Scan(scans...); err == nil {
			rm := make(map[string]string)
			for j := range values {
				rm[cols[j]] = string(values[j])
			}
			results = append(results, rm)
		}
	}
	return results, nil
}
func (c *DBConn) QueryOne(query string, vals ...any) (map[string]string, error) {
	results, err := c.Query(query, vals...)
	if err != nil {
		return nil, err
	}
	if len(results) > 0 {
		return results[0], nil
	}
	return nil, nil
}
