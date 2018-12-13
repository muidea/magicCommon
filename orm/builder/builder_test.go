package builder

import (
	"testing"
)

// Unit 单元信息
type Unit struct {
	//ID 唯一标示单元
	ID int `json:"id" orm:"id key auto"`
	// Name 名称
	Name  string  `json:"name" orm:"name"`
	Value float32 `json:"value" orm:"value"`
}

func TestBuilder(t *testing.T) {
	obj := &Unit{ID: 10, Name: "Hello world", Value: 12.3456}

	builder := NewBuilder(obj)
	if builder == nil {
		t.Error("new Builder failed")
	}

	str, err := builder.BuildCreateSchema()
	if err != nil {
		t.Error("build create schema failed")
	}
	if str != "CREATE TABLE `builder_Unit` (\n\t`id` INT NOT NULL AUTO_INCREMENT,\n\t`name` TEXT NOT NULL ,\n\t`value` FLOAT NOT NULL ,\n\tPRIMARY KEY (`id`)\n)\n" {
		t.Error("build create schema failed")
	}

	str, err = builder.BuildDropSchema()
	if err != nil {
		t.Error("build drop schema failed")
	}
	if str != "DROP TABLE IF EXISTS `builder_Unit`" {
		t.Error("build drop schema failed")
	}

	str, err = builder.BuildInsert()
	if err != nil {
		t.Error("build insert failed")
	}
	if str != "INSERT INTO `builder_Unit` (`name`,`value`) VALUES ('Hello world',12.345600)" {
		t.Error("build insert failed")
	}

	str, err = builder.BuildUpdate()
	if err != nil {
		t.Error("build update failed")
	}
	if str != "UPDATE `builder_Unit` SET `name`='Hello world',`value`=12.345600 WHERE `id`=10" {
		t.Error("build update failed")
	}

	str, err = builder.BuildDelete()
	if err != nil {
		t.Error("build delete failed")
	}
	if str != "DELETE FROM `builder_Unit` WHERE `id`=10" {
		t.Error("build delete failed")
	}

	str, err = builder.BuildQuery()
	if err != nil {
		t.Error("build query failed")
	}
	if str != "SELECT `id`,`name`,`value` FROM `builder_Unit` WHERE `id`=10" {
		t.Error("build query failed")
	}
}
