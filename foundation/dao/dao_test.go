package dao_test

import (
	"fmt"
	"testing"

	"github.com/muidea/magicCommon/foundation/dao"
)

type User struct {
	id      int
	address string
}

const user = "root"
const password = "rootkit"
const svrAddress = "localhost:3306"
const dbName = "testDB"

func TestDatabase(t *testing.T) {
	dao, err := dao.Fetch(user, password, svrAddress, "", "")
	if err != nil {
		t.Errorf("Fetch dao failed, err:%s", err.Error())
	}
	defer dao.Release()

	err = dao.CreateDatabase("supetl")
	if err != nil {
		t.Errorf("create database error:%s", err.Error())
		return
	}

	err = dao.UseDatabase("supetl")
	if err != nil {
		t.Errorf("use database error:%s", err.Error())
		return
	}

	nDao, nErr := dao.Duplicate()
	if nErr != nil {
		t.Errorf("duplicate database error:%s", err.Error())
		return
	}
	err = nDao.CreateDatabase("A1000")
	if err != nil {
		t.Errorf("create database error:%s", err.Error())
		return
	}

	defer func() {
		dropDbSql := fmt.Sprintf("drop database if exists %s", dbName)
		dao.Execute(dropDbSql)
	}()

	createDbSql := fmt.Sprintf("create database if not exists %s", dbName)
	num, _ := dao.Execute(createDbSql)
	if num != 1 {
		t.Errorf("create database failed")
	}
}

func initFunc(dao dao.Dao) {
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
	dao, err := dao.Fetch(user, password, svrAddress, "", "")
	if err != nil {
		t.Errorf("Fetch dao failed, err:%s", err.Error())
	}
	defer dao.Release()

	initFunc(dao)

	insertSql := fmt.Sprintf("%s", "insert into user (address) values(\"abc\")")
	num, _ := dao.Execute(insertSql)
	if num != 1 {
		t.Errorf("Insert data failed")
	}

	querySql := "select * from user where id=1"
	dao.Query(querySql)
}

func TestQuery(t *testing.T) {
	dao, err := dao.Fetch(user, password, svrAddress, "", "")
	if err != nil {
		t.Errorf("Fetch dao failed, err:%s", err.Error())
	}
	defer dao.Release()

	initFunc(dao)

	selectSql := fmt.Sprint("select id,address from user")

	dao.Query(selectSql)

	for dao.Next() {
		user := User{}

		dao.GetField(&user.id, &user.address)
	}
}
