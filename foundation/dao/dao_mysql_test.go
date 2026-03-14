//go:build mysql
// +build mysql

package dao

import (
	"fmt"
	"testing"
)

type User struct {
	id      int
	address string
}

const gUser = "root"
const gPassword = "rootkit"
const gSvrAddress = "localhost:3306"
const gDBName = "testdb"

func fetchOrSkip(t *testing.T) Dao {
	t.Helper()

	dao, err := Fetch(gUser, gPassword, gSvrAddress, "")
	if err != nil {
		t.Skipf("MySQL not available, skipping DAO integration test: %v", err)
	}

	return dao
}

func TestDatabase(t *testing.T) {
	dao := fetchOrSkip(t)
	defer func() { _ = dao.Release() }()

	err := dao.CreateDatabase("supetl")
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

	defer func() {
		dropDbSql := fmt.Sprintf("drop database if exists %s", gDBName)
		dao.Execute(dropDbSql)
	}()

	createDbSql := fmt.Sprintf("create database if not exists %s", gDBName)
	num, _ := dao.Execute(createDbSql)
	if num != 1 {
		t.Errorf("create database failed")
	}
}

func initFunc(dao Dao, dbName string) {
	dbSql := fmt.Sprintf("create database if not exists %s", dbName)
	dao.Execute(dbSql)

	useSql := fmt.Sprintf("use %s", dbName)
	dao.Execute(useSql)

	tableSql :=
		`
CREATE TABLE IF NOT EXISTS user (
  id int(11) NOT NULL AUTO_INCREMENT,
  address text(125),
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8
`
	dao.Execute(tableSql)
}

func TestInsert(t *testing.T) {
	dao := fetchOrSkip(t)
	defer func() { _ = dao.Release() }()

	initFunc(dao, gDBName)
	defer func() { _ = dao.DropDatabase(gDBName) }()

	insertSql := "insert into user (address) values(?),(?),(?),(?)"
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

	querySql := "select * from user where address like ?"
	param := "%a%"
	err := dao.Query(querySql, param)
	if err != nil {
		t.Errorf("dao.Query(querySql) failed, error:%s", err.Error())
		return
	}

	querySql = "select * from user where id in (?,?,?,?)"
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
	dao := fetchOrSkip(t)
	defer func() { _ = dao.Release() }()

	initFunc(dao, gDBName)
	defer func() { _ = dao.DropDatabase(gDBName) }()

	selectSql := "select id,address from user"

	err := dao.Query(selectSql)
	if err != nil {
		t.Errorf("dao.Query(selectSql) failed, error:%s", err.Error())
		return
	}
	defer func() { _ = dao.Finish() }()

	for dao.Next() {
		user := User{}

		err = dao.GetField(&user.id, &user.address)
		if err != nil {
			t.Errorf("dao.GetField(&user.id, &user.address) failed, error:%s", err.Error())
			return
		}
	}
}
