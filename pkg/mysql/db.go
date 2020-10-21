// Copyright (c) 2019-2020 Latona. All rights reserved.
package mysql

import (
	"database/sql"
	"fmt"

	// Mysql Driver
	_ "github.com/go-sql-driver/mysql"
)

// MyDB ... config connection
type MyDB struct {
	db *sql.DB
}

const dbArch = "mysql"

// NewConnection return mysqlconnection, err
func NewConnection(
	user string,
	password string,
	host string,
	port int,
	dbName string,
) (*MyDB, error) {
	if user == "" || password == "" || host == "" || port == 0 {
		return nil, fmt.Errorf("invalid param in InitializeDBConfig")
	}

	url := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		user, password, host, port, dbName)
	db, err := sql.Open(dbArch, url)

	if err != nil {
		return nil, err
	}

	return &MyDB{db}, nil
}

// CloseConnection return none
func (m *MyDB) CloseConnection() {
	if m.db != nil {
		m.db.Close()
	}
}

// FetchQuery return err
func (m *MyDB) FetchQuery(query string) (*sql.Rows, error) {
	rows, err := m.db.Query(query) //
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// ExecuteQuery return err
func (m *MyDB) ExecuteQuery(query string) error {
	_, err := m.db.Query(query)
	if err != nil {
		return err
	}

	return nil
}
