/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package mysql

import (
	"database/sql"

	"xbase/common"

	// driver.
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

// Query executes a query that returns rows
func Query(db *sql.DB, query string, args ...interface{}) ([]map[string]string, error) {
	var err error
	var results *sql.Rows
	var columns []string
	var rows []map[string]string

	if results, err = db.Query(query, args...); err != nil {
		return nil, errors.WithStack(err)
	}

	// get the rows
	if columns, err = results.Columns(); err != nil {
		return nil, errors.WithStack(err)
	}

	values := make([]interface{}, len(columns))
	for results.Next() {
		for i := 0; i < len(columns); i++ {
			values[i] = new([]byte)
		}
		if err = results.Scan(values...); err != nil {
			return nil, errors.WithStack(err)
		}

		row := make(map[string]string)
		for index, columnName := range columns {
			row[columnName] = string(*values[index].(*[]byte))
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// QueryWithTimeout used to execute the query with maxTime.
func QueryWithTimeout(db *sql.DB, maxTime int, query string, args ...interface{}) ([]map[string]string, error) {
	var err error
	var rows []map[string]string
	rsp := make(chan error, 1)

	go func() {
		rows, err = Query(db, query, args...)
		rsp <- err
	}()

	timeout := common.NormalTimeout(maxTime)
	defer common.NormalTimerRelaese(timeout)

	select {
	case <-timeout.C:
		return nil, errors.Errorf("db.query.timeout[%v, %v]", maxTime, query)
	case err := <-rsp:
		return rows, err
	}
}

// Execute executes a query without returning any rows
func Execute(db *sql.DB, query string, args ...interface{}) error {
	var err error

	if _, err = db.Exec(query, args...); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// ExecuteWithTimeout executes a query without returning any rows.
func ExecuteWithTimeout(db *sql.DB, maxTime int, query string, args ...interface{}) error {
	rsp := make(chan error, 1)

	go func() {
		_, err := db.Exec(query, args...)
		rsp <- err
	}()

	timeout := common.NormalTimeout(maxTime)
	defer common.NormalTimerRelaese(timeout)

	select {
	case <-timeout.C:
		return errors.Errorf("db.Exec.timeout[%v, %v]", maxTime, query)

	case err := <-rsp:
		return err
	}
}

// ExecuteSuperQueryList alows the user to execute queries as a super user.
func ExecuteSuperQueryList(db *sql.DB, queryList []string) error {
	for _, query := range queryList {
		if err := Execute(db, query); err != nil {
			return err
		}
	}
	return nil
}

// ExecuteSuperQueryListWithTimeout alows the user to execute queries as a super user.
func ExecuteSuperQueryListWithTimeout(db *sql.DB, maxTime int, queryList []string) error {
	for _, query := range queryList {
		if err := ExecuteWithTimeout(db, maxTime, query); err != nil {
			return err
		}
	}
	return nil
}
