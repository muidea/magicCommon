package generator

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

type Generator interface {
	GenCode() string
}

/*
prefix-{num}
prefix-{123}
prefix-{fixed(12):123}
prefix-{fixed(12):num}
prefix-{YYYYMMDDHHmmSS}-{num}
prefix-{YYYYMMDDHHmmSS}-{123}
prefix-{YYYYMMDDHHmmSS}-{fixed(12):123}
prefix-{YYYYMMDDHHmmSS}-{fixed(12):num}
*/

func New(mask string) (ret Generator, err error) {
	maskFlag, maskErr := regexp.MatchString(maskPattern, mask)
	if !maskFlag || maskErr != nil {
		err = fmt.Errorf("illegal mask pattern, expect mask pattern:%s", maskPattern)
		return
	}

	fixedWidth := "4"
	numMask := numberReg.FindString(mask)
	fixedStr := regexp.MustCompile("fixed\\([0-9]+\\)").FindString(numMask)
	if fixedStr != "" {
		fixedWidth = regexp.MustCompile("[0-9]+").FindString(fixedStr)
	}

	currentNum := 0
	numStr := regexp.MustCompile("(num|[0-9]+)}$").FindString(numMask)
	if numStr != "" {
		numStr = regexp.MustCompile("(num|[0-9]+)").FindString(numStr)
		iVal, iErr := strconv.Atoi(numStr)
		if iErr == nil {
			currentNum = iVal
		}
	}

	return &genImpl{maskStr: mask, fixedWidth: fixedWidth, currentNum: currentNum}, nil
}

// ^(?!-)([a-zA-Z]+[a-zA-Z0-9]*)?(?!--)(-{1})?({YYYYMMDDHHmmSS})?(?!--)(-{1})?{(fixed\([0-9]+\)\:)?(num|[0-9]+)}$
var maskPattern = "^([a-zA-Z]+[a-zA-Z0-9]*)?(-{1})?({YYYYMMDDHHmmSS})?(-{1})?{(fixed\\([0-9]+\\)\\:)?(num|[0-9]+)}$"
var prefixReg = regexp.MustCompile("^[a-zA-Z]+[a-zA-Z0-9]*")
var dateTimeReg = regexp.MustCompile("{YYYYMMDDHHmmSS}")
var numberReg = regexp.MustCompile("{(fixed\\([0-9]+\\)\\:)?(num|[0-9]+)}$")

type genImpl struct {
	maskStr    string
	fixedWidth string
	currentNum int
}

func (s *genImpl) GenCode() string {
	result := dateTimeReg.ReplaceAllString(s.maskStr, s.genDateTime())
	result = numberReg.ReplaceAllString(result, s.genNum())
	return result
}

func (s *genImpl) genDateTime() string {
	return time.Now().Local().Format("20060102150405")
}

func (s *genImpl) genNum() string {
	s.currentNum++
	mask := fmt.Sprintf("%%0%sd", s.fixedWidth)
	return fmt.Sprintf(mask, s.currentNum)
}
