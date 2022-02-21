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

var numReg = regexp.MustCompile("[\\d]+")

func SplitNum(code string) (ret string) {
	return numReg.FindString(code)
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
// ^(?!-)([a-zA-Z]+[a-zA-Z0-9]*)?(?!--)(-{1})?({YYYYMMDDHHmmSS})?(?!--)(-{1})?{(fixed\([0-9]+\)\:)?(num|[0-9]+)}$
var maskPatternVal = "^([a-zA-Z]+[a-zA-Z0-9]*)?(-{1})?({YYYYMMDDHHmmSS})?(-{1})?{(fixed\\(\\d+\\):)?(num|\\d+)}$"
var maskInitVal = "^([a-zA-Z]+[a-zA-Z0-9]*)?(-{1})?(\\d{14})?(-{1})?(\\d+)$"
var maskPrefix = "^[a-zA-Z]+[a-zA-Z0-9]*"
var maskMiddle = "{YYYYMMDDHHmmSS}|\\d{14}"
var maskDateTime = "YYYYMMDDHHmmSS|\\d{14}"
var maskSuffix = "({(fixed\\(\\d+\\):)?(num|\\d+)}|\\d+)$"
var maskFixed = "fixed\\(\\d+\\)"
var maskInit = ":\\d+"
var maskNumber = "(num|\\d+)"
var prefixReg = regexp.MustCompile(maskPrefix)
var middleReg = regexp.MustCompile(maskMiddle)
var suffixReg = regexp.MustCompile(maskSuffix)
var dateTimeReg = regexp.MustCompile(maskDateTime)
var fixedReg = regexp.MustCompile(maskFixed)
var initReg = regexp.MustCompile(maskInit)
var numberReg = regexp.MustCompile(maskNumber)

func New(patternVal string) (ret Generator, err error) {
	maskFlag, maskErr := regexp.MatchString(maskPatternVal, patternVal)
	if !maskFlag || maskErr != nil {
		err = fmt.Errorf("illegal patternVal pattern, expect patternVal pattern:%s", maskPatternVal)
		return
	}

	fixedWidth := "4"
	suffixVal := suffixReg.FindString(patternVal)
	fixedStr := fixedReg.FindString(suffixVal)
	if fixedStr != "" {
		fixedWidth = numReg.FindString(fixedStr)
	}

	currentNum := 0
	initStr := initReg.FindString(suffixVal)
	if initStr != "" {
		numStr := numberReg.FindString(initStr)
		iVal, iErr := strconv.Atoi(numStr)
		if iErr == nil {
			currentNum = iVal
		}
	}

	return &genImpl{patternMask: patternVal, fixedWidth: fixedWidth, currentNum: currentNum}, nil
}

func splitPatternValue(val string) (prefixVal, middleVal, suffixVal string, err error) {
	validFlag, validErr := regexp.MatchString(maskPatternVal, val)
	if !validFlag || validErr != nil {
		err = fmt.Errorf("illegal patternVal, expect :%s", maskPatternVal)
		return
	}

	prefixVal = prefixReg.FindString(val)
	middleVal = middleReg.FindString(val)
	suffixVal = suffixReg.FindString(val)
	return
}

func splitPatternMiddle(middleVal string) (dateTimeVal string, err error) {
	dateTimeVal = dateTimeReg.FindString(middleVal)
	return
}

func splitPatternSuffix(suffixVal string) (numberWidth string, numberVal int, err error) {
	numberWidth = "4"
	fixedStr := fixedReg.FindString(suffixVal)
	if fixedStr != "" {
		numberWidth = numReg.FindString(fixedStr)
	}

	numberVal = 0
	initStr := initReg.FindString(suffixVal)
	if initStr != "" {
		numStr := numberReg.FindString(initStr)
		iVal, iErr := strconv.Atoi(numStr)
		if iErr != nil {
			err = iErr
			return
		}

		numberVal = iVal
	}

	return
}

func splitInitValue(val string) (prefixVal, dateTimeVal, numberVal string, err error) {
	validFlag, validErr := regexp.MatchString(maskInitVal, val)
	if !validFlag || validErr != nil {
		err = fmt.Errorf("illegal initVal, expect :%s", maskInitVal)
		return
	}

	prefixVal = prefixReg.FindString(val)
	dateTimeVal = dateTimeReg.FindString(val)
	numberVal = suffixReg.FindString(val)
	return
}

func splitInitSuffix(suffixVal string) (numberWidth string, numberVal int, err error) {
	numberWidth = fmt.Sprintf("%d", len(suffixVal))
	iVal, iErr := strconv.Atoi(suffixVal)
	if iErr != nil {
		err = iErr
		return
	}

	numberVal = iVal

	return
}

func NewWithVal(patternVal, initVal string) (ret Generator, err error) {
	patternPrefix, patternMiddle, patternSuffix, patternErr := splitPatternValue(patternVal)
	if patternErr != nil {
		err = patternErr
		return
	}

	patternDateTime, patternErr := splitPatternMiddle(patternMiddle)
	if patternErr != nil {
		err = patternErr
		return
	}

	patternWidth, patternNumber, patternErr := splitPatternSuffix(patternSuffix)
	if patternErr != nil {
		err = patternErr
		return
	}

	initNumber := patternNumber
	if initVal != "" {
		initPrefix, initMiddle, initSuffix, initErr := splitInitValue(initVal)
		if initErr != nil {
			err = initErr
			return
		}

		initDateTime, initErr := splitPatternMiddle(initMiddle)
		if initErr != nil {
			err = initErr
			return
		}

		initWidth := patternWidth
		initWidth, initNumber, initErr = splitInitSuffix(initSuffix)
		if initErr != nil {
			err = initErr
			return
		}

		if patternPrefix != initPrefix || len(patternDateTime) != len(initDateTime) || patternWidth != initWidth {
			err = fmt.Errorf("illegal initval")
			return
		}
	}

	return &genImpl{patternMask: patternVal, fixedWidth: patternWidth, currentNum: initNumber}, nil
}

type genImpl struct {
	patternMask string
	fixedWidth  string
	currentNum  int
}

func (s *genImpl) GenCode() string {
	result := middleReg.ReplaceAllString(s.patternMask, s.genDateTime())
	result = suffixReg.ReplaceAllString(result, s.genNum())
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
