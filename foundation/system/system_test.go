package system

import (
	"fmt"
	"testing"

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
	params := []interface{}{1, "test", 13, "Abc"}

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
