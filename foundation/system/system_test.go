package system

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert" // Assuming you use testify for assertions
	// Assuming you use testify for mocking
)

type MockEntity struct {
}

func (m *MockEntity) TestMethod(iVal int, strVal string) {
	fmt.Printf("%d,%s", iVal, strVal)
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
}
