package builder

import (
	"testing"
	"time"

	"muidea.com/magicCommon/orm/model"
)

// Unit 单元信息
type Unit struct {
	//ID 唯一标示单元
	ID int `orm:"id key auto"`
	// Name 名称
	Name      string    `orm:"name"`
	Value     float32   `orm:"value"`
	TimeStamp time.Time `orm:"ts"`
}

func TestBuilder(t *testing.T) {
	now, _ := time.ParseInLocation("2006-01-02 15:04:05:0000", "2018-01-02 15:04:05:0000", time.Local)
	obj := &Unit{ID: 10, Name: "Hello world", Value: 12.3456, TimeStamp: now}

	info, depends := model.GetStructInfo(obj)
	if info == nil {
		t.Errorf("GetStructInfo failed,")
		return
	}

	if len(depends) != 0 {
		t.Errorf("GetStructInfo failed,")
		return
	}

	err := info.Verify()
	if err != nil {
		t.Errorf("Verify StructInfo failed, err:%s", err.Error())
	}

	builder := NewBuilder(info)
	if builder == nil {
		t.Error("new Builder failed")
	}

	str, err := builder.BuildCreateSchema()
	if err != nil {
		t.Errorf("build create schema failed, err:%s", err.Error())
	}
	if str != "CREATE TABLE `builder_Unit` (\n\t`id` INT NOT NULL AUTO_INCREMENT,\n\t`name` TEXT NOT NULL ,\n\t`value` FLOAT NOT NULL ,\n\t`ts` DATETIME NOT NULL ,\n\tPRIMARY KEY (`id`)\n)\n" {
		t.Error("build create schema failed")
	}

	str, err = builder.BuildDropSchema()
	if err != nil {
		t.Errorf("build drop schema failed, err:%s", err.Error())
	}
	if str != "DROP TABLE IF EXISTS `builder_Unit`" {
		t.Error("build drop schema failed")
	}

	str, err = builder.BuildInsert()
	if err != nil {
		t.Errorf("build insert failed, err:%s", err.Error())
	}
	if str != "INSERT INTO `builder_Unit` (`name`,`value`,`ts`) VALUES ('Hello world',12.345600,'2018-01-02 15:04:05')" {
		t.Error("build insert failed")
	}

	str, err = builder.BuildUpdate()
	if err != nil {
		t.Errorf("build update failed, err:%s", err.Error())
	}
	if str != "UPDATE `builder_Unit` SET `name`='Hello world',`value`=12.345600,`ts`='2018-01-02 15:04:05' WHERE `id`=10" {
		t.Error("build update failed")
	}

	str, err = builder.BuildDelete()
	if err != nil {
		t.Errorf("build delete failed, err:%s", err.Error())
	}
	if str != "DELETE FROM `builder_Unit` WHERE `id`=10" {
		t.Error("build delete failed")
	}

	str, err = builder.BuildQuery()
	if err != nil {
		t.Errorf("build query failed, err:%s", err.Error())
	}
	if str != "SELECT `id`,`name`,`value`,`ts` FROM `builder_Unit` WHERE `id`=10" {
		t.Error("build query failed")
	}
}
