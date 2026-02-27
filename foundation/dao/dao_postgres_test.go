//go:build !mysql
// +build !mysql

package dao

import (
	"fmt"
	"testing"
)

type User struct {
	id      int
	address string
}

const gUser = "postgres"
const gPassword = "rootkit"
const gSvrAddress = "localhost:5432"
const gDBName = "testdb001"

func TestDatabase(t *testing.T) {
	dao, err := Fetch(gUser, gPassword, gSvrAddress, "")
	if err != nil {
		t.Errorf("Fetch dao failed, err:%s", err.Error())
	}
	defer func() { _ = dao.Release() }()

	err = dao.CreateDatabase("supetl")
	if err != nil {
		t.Errorf("create database error:%s", err.Error())
		return
	}
	defer func() { _ = dao.DropDatabase("supetl") }()

	err = dao.UseDatabase("supetl")
	if err != nil {
		t.Errorf("use database error:%s", err.Error())
		return
	}

	nDao, nErr := dao.Duplicate()
	if nErr != nil {
		t.Errorf("duplicate database error:%s", nErr.Error())
		return
	}
	defer func() { _ = nDao.Release() }()
	err = nDao.CreateDatabase("A1000")
	if err != nil {
		t.Errorf("create database error:%s", err.Error())
		return
	}
	defer func() { _ = nDao.DropDatabase("A1000") }()
}

func initFunc(dao Dao, dbName string) {
	dbSql := fmt.Sprintf("CREATE DATABASE \"%s\"", dbName)
	_, _ = dao.Execute(dbSql)

	// PostgreSQL 不需要 USE 语句，连接时已经指定了数据库

	tableSql :=
		`
CREATE TABLE IF NOT EXISTS "user" (
  id SERIAL PRIMARY KEY,
  address TEXT
)
`
	_, _ = dao.Execute(tableSql)
}

func TestInsert(t *testing.T) {
	dao, err := Fetch(gUser, gPassword, gSvrAddress, "")
	if err != nil {
		t.Errorf("Fetch dao failed, err:%s", err.Error())
	}
	defer func() { _ = dao.Release() }()

	initFunc(dao, gDBName)
	defer func() { _ = dao.DropDatabase(gDBName) }()

	insertSql := "insert into \"user\" (address) values($1),($2),($3),($4)"
	num, _ := dao.Execute(insertSql, "abc", "bcd", "cde", "def")
	if num != 4 {
		t.Errorf("Insert data failed")
	}
	num, _ = dao.Execute(insertSql, "abc", "bcd", "cde", "def")
	if num != 4 {
		t.Errorf("Insert data failed")
	}
	num, _ = dao.Execute(insertSql, "abc", "bcd", "cde", "def")
	if num != 4 {
		t.Errorf("Insert data failed")
	}
	num, _ = dao.Execute(insertSql, "abc", "bcd", "cde", "def")
	if num != 4 {
		t.Errorf("Insert data failed")
	}
	num, _ = dao.Execute(insertSql, "abc", "bcd", "cde", "def")
	if num != 4 {
		t.Errorf("Insert data failed")
	}

	querySql := "select * from \"user\" where address like $1"
	param := "%a%"
	err = dao.Query(querySql, param)
	if err != nil {
		t.Errorf("dao.Query(querySql) failed, error:%s", err.Error())
		return
	}

	querySql = "select * from \"user\" where id in ($1,$2,$3,$4)"
	ids := []any{1, 2, 3, 4}
	err = dao.Query(querySql, ids...)
	if err != nil {
		t.Errorf("dao.Query(querySql) failed, error:%s", err.Error())
		return
	}
	defer func() { _ = dao.Finish() }()
	if dao.Next() {
		u1 := User{}
		err = dao.GetField(&u1.id, &u1.address)
		if err != nil {
			t.Errorf("dao.GetField(&u1.id, &u1.address) failed, error:%s", err.Error())
			return
		}
		if u1.id != 1 {
			t.Errorf("dao.Query failed")
			return
		}
	}
}

func TestQuery(t *testing.T) {
	dao, err := Fetch(gUser, gPassword, gSvrAddress, "")
	if err != nil {
		t.Errorf("Fetch dao failed, err:%s", err.Error())
	}
	defer func() { _ = dao.Release() }()

	initFunc(dao, gDBName)
	defer func() { _ = dao.DropDatabase(gDBName) }()

	selectSql := "select id,address from \"user\""

	_ = dao.Query(selectSql)

	for dao.Next() {
		user := User{}

		_ = dao.GetField(&user.id, &user.address)
	}
}
