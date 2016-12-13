package forge_test

import (
	"testing"

	"github.com/brettlangdon/forge"
)

func TestNewPrimative(t *testing.T) {
	t.Parallel()

	value, err := forge.NewPrimative(true)
	if err != nil {
		t.Error(err)
		return
	}

	if value.GetType() != forge.BOOLEAN {
		t.Error("value is not a BOOLEAN")
		return
	}

	rawVal := value.GetValue()
	switch rawVal.(type) {
	case bool:
		// this is what we want to happen
	default:
		t.Error("expected value.GetValue() to be a bool")
		return
	}

	if rawVal.(bool) != true {
		t.Error("expected value.GetValue() to be equal to 'true'")
	}
}

func TestNewBoolean(t *testing.T) {
	t.Parallel()

	value := forge.NewBoolean(true)

	if value.GetType() != forge.BOOLEAN {
		t.Error("value is not a BOOLEAN")
		return
	}

	rawVal := value.GetValue()
	switch rawVal.(type) {
	case bool:
		// this is what we want to happen
	default:
		t.Error("expected value.GetValue() to be a bool")
		return
	}

	if rawVal.(bool) != true {
		t.Error("expected value.GetValue() to be equal to 'true'")
	}
}

func TestNewFloat(t *testing.T) {
	t.Parallel()

	value := forge.NewFloat(50.5)

	if value.GetType() != forge.FLOAT {
		t.Error("value is not a FLOAT")
		return
	}

	rawVal := value.GetValue()
	switch rawVal.(type) {
	case float64:
		// this is what we want to happen
	default:
		t.Error("expected value.GetValue() to be a float64")
		return
	}

	if rawVal.(float64) != 50.5 {
		t.Error("expected value.GetValue() to be equal to '50.5'")
	}
}

func TestNewInteger(t *testing.T) {
	t.Parallel()

	value := forge.NewInteger(50)

	if value.GetType() != forge.INTEGER {
		t.Error("value is not a INTEGER")
		return
	}

	rawVal := value.GetValue()
	switch rawVal.(type) {
	case int64:
		// this is what we want to happen
	default:
		t.Error("expected value.GetValue() to be a int64")
		return
	}

	if rawVal.(int64) != 50 {
		t.Error("expected value.GetValue() to be equal to '50'")
	}
}

func TestNewNull(t *testing.T) {
	t.Parallel()

	value := forge.NewNull()

	if value.GetType() != forge.NULL {
		t.Error("value is not a NULL")
		return
	}

	rawVal := value.GetValue()
	switch rawVal.(type) {
	case nil:
		// this is what we want to happen
	default:
		t.Error("expected value.GetValue() to be a nil")
		return
	}

	if rawVal != nil {
		t.Error("expected value.GetValue() to be equal to 'nil'")
	}
}

func TestNewString(t *testing.T) {
	t.Parallel()

	value := forge.NewString("hello")

	if value.GetType() != forge.STRING {
		t.Error("value is not a STRING")
		return
	}

	rawVal := value.GetValue()
	switch rawVal.(type) {
	case string:
		// this is what we want to happen
	default:
		t.Error("expected value.GetValue() to be a string")
		return
	}

	if rawVal.(string) != "hello" {
		t.Error("expected value.GetValue() to be equal to '\"hello\"'")
	}
}

func TestUpdateValue(t *testing.T) {
	t.Parallel()

	value := forge.NewNull()

	// Boolean
	err := value.UpdateValue(true)
	if err != nil {
		t.Error(err)
		return
	}
	if value.GetType() != forge.BOOLEAN {
		t.Error("value is not a BOOLEAN")
		return
	}
	if value.GetValue().(bool) != true {
		t.Error("expected value.GetValue() to be equal to 'true'")
	}

	// Float
	err = value.UpdateValue(50.5)
	if err != nil {
		t.Error(err)
		return
	}
	if value.GetType() != forge.FLOAT {
		t.Error("value is not a FLOAT")
		return
	}
	if value.GetValue().(float64) != 50.5 {
		t.Error("expected value.GetValue() to be equal to '50.5'")
	}

	// Integer int
	err = value.UpdateValue(50)
	if err != nil {
		t.Error(err)
		return
	}
	if value.GetType() != forge.INTEGER {
		t.Error("value is not a INTEGER")
		return
	}
	if value.GetValue().(int64) != 50 {
		t.Error("expected value.GetValue() to be equal to '50'")
	}

	// Integer int64
	err = value.UpdateValue(int64(50))
	if err != nil {
		t.Error(err)
		return
	}
	if value.GetType() != forge.INTEGER {
		t.Error("value is not a INTEGER")
		return
	}
	if value.GetValue().(int64) != 50 {
		t.Error("expected value.GetValue() to be equal to '50'")
	}

	// Null
	err = value.UpdateValue(nil)
	if err != nil {
		t.Error(err)
		return
	}
	if value.GetType() != forge.NULL {
		t.Error("value is not a NULL")
		return
	}
	if value.GetValue() != nil {
		t.Error("expected value.GetValue() to be equal to '50'")
	}

	// String
	err = value.UpdateValue("hello")
	if err != nil {
		t.Error(err)
		return
	}
	if value.GetType() != forge.STRING {
		t.Error("value is not a STRING")
		return
	}
	if value.GetValue().(string) != "hello" {
		t.Error("expected value.GetValue() to be equal to '\"hello\"'")
		return
	}

}

func TestUpdateValueUnknown(t *testing.T) {
	t.Parallel()
	value, err := forge.NewPrimative("hello")
	if err != nil {
		t.Error(err)
		return
	}

	newVal := []string{"hello", "world"}
	err = value.UpdateValue(newVal)
	if err == nil {
		t.Error("expected an error, got none")
		return
	}
}

func TestAsBoolean(t *testing.T) {
	t.Parallel()

	// Boolean true
	value := forge.NewBoolean(true)
	val, err := value.AsBoolean()
	if err != nil {
		t.Error(err)
		return
	}
	if val != true {
		t.Error("expected value to be 'true'")
		return
	}

	// Boolean false
	err = value.UpdateValue(false)
	if err != nil {
		t.Error(err)
		return
	}
	val, err = value.AsBoolean()
	if err != nil {
		t.Error(err)
		return
	}
	if val != false {
		t.Error("expected value to be 'false'")
		return
	}

	// Int true
	err = value.UpdateValue(1)
	if err != nil {
		t.Error(err)
		return
	}
	val, err = value.AsBoolean()
	if err != nil {
		t.Error(err)
		return
	}
	if val != true {
		t.Error("expected value to be 'true'")
		return
	}

	// Int false
	err = value.UpdateValue(0)
	if err != nil {
		t.Error(err)
		return
	}
	val, err = value.AsBoolean()
	if err != nil {
		t.Error(err)
		return
	}
	if val != false {
		t.Error("expected value to be 'false'")
		return
	}

	// Float true
	err = value.UpdateValue(float64(1))
	if err != nil {
		t.Error(err)
		return
	}
	val, err = value.AsBoolean()
	if err != nil {
		t.Error(err)
		return
	}
	if val != true {
		t.Error("expected value to be 'true'")
		return
	}

	// Float false
	err = value.UpdateValue(float64(0))
	if err != nil {
		t.Error(err)
		return
	}
	val, err = value.AsBoolean()
	if err != nil {
		t.Error(err)
		return
	}
	if val != false {
		t.Error("expected value to be 'false'")
		return
	}

	// Null true
	err = value.UpdateValue(nil)
	if err != nil {
		t.Error(err)
		return
	}
	val, err = value.AsBoolean()
	if err != nil {
		t.Error(err)
		return
	}
	if val != false {
		t.Error("expected value to be 'false'")
		return
	}

	// String true
	err = value.UpdateValue("anything")
	if err != nil {
		t.Error(err)
		return
	}
	val, err = value.AsBoolean()
	if err != nil {
		t.Error(err)
		return
	}
	if val != true {
		t.Error("expected value to be 'true'")
		return
	}

	// String false
	err = value.UpdateValue("")
	if err != nil {
		t.Error(err)
		return
	}
	val, err = value.AsBoolean()
	if err != nil {
		t.Error(err)
		return
	}
	if val != false {
		t.Error("expected value to be 'false'")
		return
	}
}
