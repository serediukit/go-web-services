package main

import (
	"fmt"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {
	// fmt.Printf("\nGOT:\n%T\n%#v\nOUT:\n%T\n%#v\n\n", data, data, out, out)

	value := reflect.ValueOf(out)

	if value.Kind() != reflect.Ptr {
		return fmt.Errorf("data must be a pointer")
	}

	value = value.Elem()

	switch value.Kind() {
	case reflect.Struct:
		d, ok := data.(map[string]interface{})

		if !ok {
			return fmt.Errorf("data is not a map")
		}

		for i := 0; i < value.NumField(); i++ {
			fieldName := value.Type().Field(i).Name

			v, ok := d[fieldName]

			if !ok {
				return fmt.Errorf("field %s not found", fieldName)
			}

			if err := i2s(v, value.Field(i).Addr().Interface()); err != nil {
				return fmt.Errorf("failed to reflect struct %s", fieldName)
			}
		}
	case reflect.Slice:
		fmt.Printf("Slice:%+v\n", value)
	case reflect.Int:
		d, ok := data.(float64)

		if !ok {
			return fmt.Errorf("data must be float64")
		}

		value.SetInt(int64(d))

	case reflect.String:
		d, ok := data.(string)

		if !ok {
			return fmt.Errorf("data must be string")
		}

		value.SetString(d)

	case reflect.Bool:
		d, ok := data.(bool)

		if !ok {
			return fmt.Errorf("data must be bool")
		}

		value.SetBool(d)

	default:
		return fmt.Errorf("unknown type")
	}

	fmt.Println()
	return nil
}
