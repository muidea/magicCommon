package system

import (
	"fmt"
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/stretchr/testify/assert" // Assuming you use testify for assertions
	// Assuming you use testify for mocking
)

type MockEntity struct {
}

func (s *MockEntity) TestMethod(iVal int, strVal string) {
	fmt.Printf("%d,%s", iVal, strVal)
}

func (s *MockEntity) TestDemo() {
	fmt.Println("test demo")
}

func (s *MockEntity) TestReuslt() *cd.Result {
	return cd.NewResult(cd.InvalidAuthority, "test result")
}

// TestInvokeEntityFuncNoMethod tests the scenario where the method does not exist on the entityVal
func TestInvokeEntityFuncNoMethod(t *testing.T) {
	entityVal := &MockEntity{}
	funcName := "NonExistentMethod"

	result := InvokeEntityFunc(entityVal, funcName)
	assert.NotNil(t, result)
	assert.Equal(t, result.Fail(), true) // Assuming there's a Type field in the cd.Result type to check the error type
}

// TestInvokeEntityFuncWithMethod tests the successful invocation of an existing method
func TestInvokeEntityFuncWithMethod(t *testing.T) {
	entityVal := &MockEntity{}

	funcName := "TestMethod"
	params := []interface{}{1, "test"}

	result := InvokeEntityFunc(entityVal, funcName, params...)
	assert.Nil(t, result)

	params = []interface{}{"test"}
	result = InvokeEntityFunc(entityVal, funcName, params...)
	assert.NotNil(t, result)
}

func TestInvokeEntityFuncWithDemo(t *testing.T) {
	entityVal := &MockEntity{}

	funcName := "TestDemo"
	params := []interface{}{1, "test"}

	result := InvokeEntityFunc(entityVal, funcName, params...)
	assert.Nil(t, result)
}

func TestInvokeEntityFuncWithResult(t *testing.T) {
	entityVal := &MockEntity{}
	funcName := "TestReuslt"
	params := []interface{}{}
	result := InvokeEntityFunc(entityVal, funcName, params...)
	assert.NotNil(t, result)
	assert.Equal(t, result.ErrorCode, cd.ErrorCode(cd.InvalidAuthority))
}

// TestInvokeEntityFuncNilEntity tests the scenario where the entityVal is nil
func TestInvokeEntityFuncNilEntity(t *testing.T) {
	funcName := "TestMethod"
	params := []interface{}{1, "test"}

	result := InvokeEntityFunc(nil, funcName, params...)
	assert.NotNil(t, result)
	assert.Equal(t, result.ErrorCode, cd.ErrorCode(cd.IllegalParam))
}

// TestInvokeEntityFuncInvalidParamType tests the scenario where the parameter type is invalid
func TestInvokeEntityFuncInvalidParamType(t *testing.T) {
	entityVal := &MockEntity{}
	funcName := "TestMethod"
	params := []interface{}{"invalid", "test"}

	result := InvokeEntityFunc(entityVal, funcName, params...)
	assert.NotNil(t, result)
	assert.Equal(t, result.ErrorCode, cd.ErrorCode(cd.IllegalParam))
}
