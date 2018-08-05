package viperstruct

import (
	"reflect"
	"errors"
	"strings"
	"fmt"
	"github.com/spf13/viper"
	"github.com/spf13/cast"
)

// TODO support string slices
// TODO Tests

func ReadConfig(dest interface{}) error {
	ptrVal := reflect.ValueOf(dest)

	// Check if dest is a pointer
	if ptrVal.Kind() != reflect.Ptr {
		return errors.New("viperstruct: argument must be a pointer to a struct")
	}

	destVal := ptrVal.Elem()

	// Check if dest is a struct
	if destVal.Kind() != reflect.Struct {
		return errors.New("viperstruct: target object must be a struct")
	}

	// Enumerate fields in struct
	for i := 0; i < destVal.NumField(); i++ {
		// Check if tag exists
		fieldType := destVal.Type().Field(i)
		tag := fieldType.Tag.Get("viper")
		if tag == "" { continue }

		// Tag options
		optional := false

		// Parse tag
		elements := strings.Split(tag, ",")
		name := elements[0]
		for _, option := range elements[1:] {
			switch option {
			case "optional": optional = true
			default: continue
			}
		}

		// Check if writable
		fieldValue := destVal.Field(i)
		if !fieldValue.IsValid() { continue }
		if !fieldValue.CanSet() {
			return fmt.Errorf("viperstruct: unexported field: %s", name)
		}

		// Get config value
		configValue := viper.Get(name)
		if configValue == nil || configValue == "" {
			if optional {
				continue
			} else {
				return fmt.Errorf("viperstruct: field \"%s\" is not set", name)
			}
		}

		// Get type
		switch fieldValue.Kind() {
		case reflect.Bool:
			v, err := cast.ToBoolE(configValue)
			if err != nil { return err }
			fieldValue.SetBool(v)
		case reflect.String:
			v, err := cast.ToStringE(configValue)
			if err != nil { return err }
			fieldValue.SetString(v)
		case reflect.Int:
			v, err := cast.ToIntE(configValue)
			if err != nil { return err }
			fieldValue.SetInt(int64(v))
		case reflect.Int8:
			v, err := cast.ToInt8E(configValue)
			if err != nil { return err }
			fieldValue.SetInt(int64(v))
		case reflect.Int16:
			v, err := cast.ToInt16E(configValue)
			if err != nil { return err }
			fieldValue.SetInt(int64(v))
		case reflect.Int32:
			v, err := cast.ToInt32E(configValue)
			if err != nil { return err }
			fieldValue.SetInt(int64(v))
		case reflect.Int64:
			v, err := cast.ToInt64E(configValue)
			if err != nil { return err }
			fieldValue.SetInt(int64(v))
		case reflect.Uint:
			v, err := cast.ToUintE(configValue)
			if err != nil { return err }
			fieldValue.SetUint(uint64(v))
		case reflect.Uint8:
			v, err := cast.ToUint8E(configValue)
			if err != nil { return err }
			fieldValue.SetUint(uint64(v))
		case reflect.Uint16:
			v, err := cast.ToUint16E(configValue)
			if err != nil { return err }
			fieldValue.SetUint(uint64(v))
		case reflect.Uint32:
			v, err := cast.ToUint32E(configValue)
			if err != nil { return err }
			fieldValue.SetUint(uint64(v))
		case reflect.Uint64:
			v, err := cast.ToUint64E(configValue)
			if err != nil { return err }
			fieldValue.SetUint(uint64(v))
		case reflect.Float32:
			v, err := cast.ToFloat32E(configValue)
			if err != nil { return err }
			fieldValue.SetFloat(float64(v))
		case reflect.Float64:
			v, err := cast.ToFloat64E(configValue)
			if err != nil { return err }
			fieldValue.SetFloat(float64(v))
		default:
			return fmt.Errorf("viperstruct: unsupported type %s", fieldValue.Type().String())
		}
	}

	return nil
}