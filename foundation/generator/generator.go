package generator

import (
	"fmt"
	"regexp"
	"time"
)

type Generator interface {
	GenCode() string
}

/*
prefix-{YYYYMMDDHHmmSS}-{num}
*/
func New(mask string) (ret Generator, err error) {
	maskFlag, maskErr := regexp.MatchString(maskPattern, mask)
	if !maskFlag || maskErr != nil {
		err = fmt.Errorf("illegal mask pattern, expect mask pattern:%s", maskPattern)
		return
	}

	return &genImpl{maskStr: mask, currentNum: 0}, nil
}

var maskPattern = "^[a-zA-Z]+[a-zA-Z0-9]*-|{YYYYMMDDHHmmSS}|-{num}$"
var prefixReg = regexp.MustCompile("^[a-zA-Z]+[a-zA-Z0-9]*")
var dateTimeReg = regexp.MustCompile("{YYYYMMDDHHmmSS}")
var numReg = regexp.MustCompile("{num}$")

type genImpl struct {
	maskStr    string
	currentNum int
}

func (s *genImpl) GenCode() string {
	s.currentNum++
	datTime := time.Now().Local().Format("20060102150405")
	result := dateTimeReg.ReplaceAllString(s.maskStr, datTime)
	result = numReg.ReplaceAllString(result, fmt.Sprintf("%04d", s.currentNum))
	return result
}
