/*
 * Xenon
 *
 * Copyright 2018 The Xenon Authors.
 * Code is licensed under the GPLv3.
 *
 */

package mysql

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestMysqlSimpleQuery(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	want := "Mysql5.7"
	columns := []string{"VERSION"}
	query := "SELECT VERSION"
	mockrows := sqlmock.NewRows(columns).AddRow(want)
	mock.ExpectQuery(query).WillReturnRows(mockrows)
	rows, err := Query(db, query)
	assert.Nil(t, err)

	got := rows[0]["VERSION"]
	assert.Equal(t, want, got)
}

func TestMysqlSimpleQueryWithTimeout(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	columns := []string{"VERSION"}
	query := "SELECT VERSION"
	mockrows := sqlmock.NewRows(columns).AddRow("Mysql5.7")
	mock.ExpectQuery(query).WillReturnRows(mockrows)

	{
		rows, err := QueryWithTimeout(db, 100, query)
		assert.Nil(t, err)
		want := "Mysql5.7"
		got := rows[0]["VERSION"]
		assert.Equal(t, want, got)
	}

	{
		_, err = QueryWithTimeout(db, 0, query)
		assert.NotNil(t, err)
	}
}

func TestDBExecuteTimeout(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	queryList := []string{"STOP SLAVE",
		"RESET SLAVE ALL"}
	mock.ExpectExec(queryList[0]).WillReturnResult(sqlmock.NewResult(1, 1))
	err = ExecuteWithTimeout(db, 100, queryList[0])
	assert.Nil(t, err)

	mock.ExpectExec(queryList[1]).WillReturnResult(sqlmock.NewResult(1, 1))
	err = ExecuteSuperQueryListWithTimeout(db, 0, queryList)

	assert.NotNil(t, err)
}

func TestMysqlComplexQuery(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	query := "SELECT a,b FROM t1"
	columns := []string{"a", "b"}
	tableRows := sqlmock.NewRows(columns).
		AddRow(1, 1).
		AddRow(2, 2)

	mock.ExpectQuery(query).WillReturnRows(tableRows)
	rows, err := Query(db, query)
	assert.Nil(t, err)
	want := "2"
	got := rows[1]["a"]
	assert.Equal(t, want, got)
}

func TestMysqlExecuteQueryList(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	defer db.Close()

	queryList := []string{"STOP SLAVE", "SELECT 1"}

	mock.ExpectExec(queryList[0]).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(queryList[1]).WillReturnResult(sqlmock.NewResult(1, 1))
	err = ExecuteSuperQueryList(db, queryList)
	assert.Nil(t, err)
}
