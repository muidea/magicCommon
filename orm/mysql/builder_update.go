package mysql

import (
	"fmt"
	"log"
)

// BuildUpdate  BuildUpdate
func (s *Builder) BuildUpdate() (ret string, err error) {
	str := ""
	for _, val := range *s.structInfo.GetFields() {
		fType := val.GetFieldType()
		fValue := val.GetFieldValue()
		fTag := val.GetFieldTag()

		if fValue == nil {
			continue
		}

		if fType.IsPtr() && fValue.IsNil() {
			continue
		}

		dependType, _ := fType.Depend()
		if dependType != nil {
			continue
		}

		if val != s.structInfo.GetPrimaryField() {
			fStr, ferr := fValue.GetValueStr()
			if ferr != nil {
				err = ferr
				break
			}
			if str == "" {
				str = fmt.Sprintf("`%s`=%s", fTag.Name(), fStr)
			} else {
				str = fmt.Sprintf("%s,`%s`=%s", str, fTag.Name(), fStr)
			}
		}
	}

	if err != nil {
		return
	}

	pkfValue := s.structInfo.GetPrimaryField().GetFieldValue()
	pkfTag := s.structInfo.GetPrimaryField().GetFieldTag()
	pkfStr, pkferr := pkfValue.GetValueStr()
	if pkferr == nil {
		str = fmt.Sprintf("UPDATE `%s` SET %s WHERE `%s`=%s", s.getTableName(s.structInfo), str, pkfTag.Name(), pkfStr)
		log.Print(str)
	}

	ret = str
	err = pkferr

	return
}
