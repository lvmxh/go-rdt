package util

import (
	"fmt"
	"reflect"
	"strconv"
)

// Set obj's 'name' field with proper type
func SetField(obj interface{}, name string, value interface{}) error {
	structValue := reflect.ValueOf(obj).Elem()
	structFieldValue := structValue.FieldByName(name)

	if !structFieldValue.IsValid() {
		return fmt.Errorf("No such field: %s in obj", name)
	}

	if !structFieldValue.CanSet() {
		return fmt.Errorf("Cannot set %s field value", name)
	}

	val := reflect.ValueOf(value)
	structFieldType := structFieldValue.Type()

	switch structFieldType.Name() {
	case "int":
		v := value.(string)
		v_int, err := strconv.Atoi(v)
		if err != nil {
			// add log
			return err
		}
		val = reflect.ValueOf(v_int)
	}

	structFieldValue.Set(val)
	return nil
}
