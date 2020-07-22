package gts

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

const (
	k_BINDING_TAG                 = "form"
	k_BINDING_CLEANED_FUNC_PREFIX = "Cleaned"
	k_BINDING_NO_TAG              = "-"
	k_BINDING_CLEANED_DATA        = "CleanedData"
	k_BINDING_DEFAULT_FUNC_PREFIX = "Default"
)

func Bind(request *http.Request, result interface{}) (err error) {
	err = request.ParseForm()
	if err != nil {
		return err
	}

	var contentType = request.Header.Get("Content-Type")
	if strings.Contains(contentType, "multipart/form-data") {
		request.ParseMultipartForm(32 << 20)
	}

	err = BindForm(request.Form, result)
	return err
}

func BindForm(values map[string][]string, result interface{}) (err error) {
	var source = make(map[string]interface{})
	for key, value := range values {
		if len(value) > 1 {
			source[key] = value
		} else {
			source[key] = value[0]
		}
	}
	return BindWithTag(source, result, "form")
}

func BindWith(source map[string]interface{}, result interface{}) (err error) {
	return BindWithTag(source, result, k_BINDING_TAG)
}

func BindWithTag(source map[string]interface{}, result interface{}, tag string) error {
	return BindWithAdvanced(source, result, tag, tag)
}

func BindWithAdvanced(source map[string]interface{}, result interface{}, tag, cleanedTag string) error {
	var objType = reflect.TypeOf(result)
	var objValue = reflect.ValueOf(result)
	var objValueKind = objValue.Kind()

	if objValueKind == reflect.Struct {
		return errors.New("result argument is struct")
	}

	if objValue.IsNil() {
		return errors.New("result argument is nil")
	}

	for {
		if objValueKind == reflect.Ptr && objValue.IsNil() {
			objValue.Set(reflect.New(objType.Elem()))
		}

		if objValueKind == reflect.Ptr {
			objValue = objValue.Elem()
			objType = objType.Elem()
			objValueKind = objValue.Kind()
			continue
		}
		break
	}

	var cleanDataValue = objValue.FieldByName(k_BINDING_CLEANED_DATA)
	if cleanDataValue.IsValid() && cleanDataValue.IsNil() {
		cleanDataValue.Set(reflect.MakeMap(cleanDataValue.Type()))
	}
	return bindWithMap(objType, objValue, objValue, cleanDataValue, source, tag, cleanedTag)
}

func bindWithMap(objType reflect.Type, currentObjValue, objValue, cleanDataValue reflect.Value, source map[string]interface{}, tagName, cleanedTagName string) error {
	var numField = objType.NumField()
	for i := 0; i < numField; i++ {
		var fieldStruct = objType.Field(i)
		var fieldValue = objValue.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		var tag = fieldStruct.Tag.Get(tagName)

		if tag == "" && fieldStruct.Name != k_BINDING_CLEANED_DATA {
			tag = fieldStruct.Name

			if fieldValue.Kind() == reflect.Ptr {
				if fieldValue.IsNil() {
					fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
				}
				fieldValue = fieldValue.Elem()
			}

			if fieldValue.Kind() == reflect.Struct {
				if err := bindWithMap(fieldValue.Addr().Type().Elem(), currentObjValue, fieldValue, cleanDataValue, source, tagName, cleanedTagName); err != nil {
					return err
				}
				continue
			}
		} else if tag == k_BINDING_NO_TAG {
			continue
		}

		var value, exists = source[tag]
		if !exists {
			if ok := setDefaultValue(currentObjValue, objValue, fieldValue, fieldStruct); !ok {
				continue
			}
		} else {
			if err := setValue(currentObjValue, objValue, fieldValue, fieldStruct, value); err != nil {
				return err
			}
		}

		var cdTag = tag
		if tagName != cleanedTagName {
			cdTag = fieldStruct.Tag.Get(cleanedTagName)
		}
		setCleanedData(cleanDataValue, fieldValue, cdTag)
	}
	return nil
}

func setCleanedData(cleanDataValue, fieldValue reflect.Value, cdTag string) {
	if cleanDataValue.IsValid() {
		if cdTag == k_BINDING_NO_TAG {
			return
		}
		cleanDataValue.SetMapIndex(reflect.ValueOf(cdTag), fieldValue)
	}
}

func getFuncWithName(funcName string, currentObjValue, objValue reflect.Value) reflect.Value {
	var funcValue = currentObjValue.MethodByName(funcName)
	if funcValue.IsValid() == false {
		if currentObjValue.CanAddr() {
			funcValue = currentObjValue.Addr().MethodByName(funcName)
		}
	}
	if funcValue.IsValid() == false && currentObjValue != objValue {
		return getFuncWithName(funcName, objValue, objValue)
	}
	return funcValue
}

func setDefaultValue(currentObjValue, objValue, fieldValue reflect.Value, fieldStruct reflect.StructField) bool {
	var funcValue = getFuncWithName(k_BINDING_DEFAULT_FUNC_PREFIX+fieldStruct.Name, currentObjValue, objValue)
	if funcValue.IsValid() {
		var rList = funcValue.Call(nil)
		fieldValue.Set(rList[0])
		return true
	}
	return false
}

func setValue(currentObjValue, objValue, fieldValue reflect.Value, fieldStruct reflect.StructField, value interface{}) error {
	var vValue = reflect.ValueOf(value)
	var fieldValueKind = fieldValue.Kind()

	var mValue = getFuncWithName(k_BINDING_CLEANED_FUNC_PREFIX+fieldStruct.Name, currentObjValue, objValue)
	if mValue.IsValid() {
		var rList = mValue.Call([]reflect.Value{vValue})
		if len(rList) > 1 {
			var rValue1 = rList[1]
			if rValue1.IsNil() == false {
				return rValue1.Interface().(error)
			}
		}
		fieldValue.Set(rList[0])
	} else if fieldValueKind == reflect.Slice /* && fieldValue.IsNil() == false */ {
		var valueLen int
		if vValue.Kind() == reflect.Slice {
			// 如果绑定源数据也是 slice
			valueLen = vValue.Len()
			var s = reflect.MakeSlice(fieldValue.Type(), valueLen, valueLen)
			for i := 0; i < valueLen; i++ {
				if err := _setValue(s.Index(i), fieldStruct, vValue.Index(i)); err != nil {
					return err
				}
			}
			fieldValue.Set(s)
		} else {
			// 如果绑定源数据不是 slice
			valueLen = 1
			var s = reflect.MakeSlice(fieldValue.Type(), valueLen, valueLen)
			if err := _setValue(s.Index(0), fieldStruct, vValue); err != nil {
				return err
			}
			fieldValue.Set(s)
		}
	} else {
		return _setValue(fieldValue, fieldStruct, vValue)
	}
	return nil
}

func _setValue(fieldValue reflect.Value, fieldStruct reflect.StructField, value reflect.Value) error {
	var valueKind = value.Kind()
	var fieldKind = fieldValue.Kind()

	if valueKind == reflect.Slice {
		// 如果源数据是 slice, 则取出其第一个数据
		value = value.Index(0)
		valueKind = value.Kind()
	}

	if valueKind == fieldKind {
		return _setValueWithSameKind(fieldValue, fieldStruct, valueKind, value)
	}
	return _setValueWithDiffKind(fieldValue, fieldStruct, valueKind, value)
}

func _setValueWithSameKind(fieldValue reflect.Value, fieldStruct reflect.StructField, valueKind reflect.Kind, value reflect.Value) error {
	switch valueKind {
	case reflect.String:
		fieldValue.SetString(value.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fieldValue.SetInt(value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		fieldValue.SetUint(value.Uint())
	case reflect.Float32, reflect.Float64:
		fieldValue.SetFloat(value.Float())
	case reflect.Bool:
		fieldValue.SetBool(value.Bool())
	case reflect.Struct:
		fieldValue.Set(value)
	default:
		return errors.New(fmt.Sprintf("Unknown type: %s", fieldStruct.Name))
	}
	return nil
}

func _setValueWithDiffKind(fieldValue reflect.Value, fieldStruct reflect.StructField, valueKind reflect.Kind, value reflect.Value) (err error) {
	var fieldValueKind = fieldValue.Kind()

	var f float64
	if fieldValueKind != reflect.Bool {
		f, err = floatValue(valueKind, value)
		if err != nil {
			return errors.New(fmt.Sprintln("[" + fieldStruct.Name + "]" + err.Error()))
		}
	}

	switch fieldValueKind {
	case reflect.String:
		fieldValue.SetString(fmt.Sprintf("%f", f))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fieldValue.SetInt(int64(f))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		fieldValue.SetUint(uint64(f))
	case reflect.Float32, reflect.Float64:
		fieldValue.SetFloat(f)
	case reflect.Bool:
		var b = boolValue(valueKind, value)
		fieldValue.SetBool(b)
	default:
		return errors.New(fmt.Sprintf("Unknown type: %s", fieldStruct.Name))
	}
	return nil
}

func boolValue(valueKind reflect.Kind, value reflect.Value) bool {
	switch valueKind {
	case reflect.String:
		var v = value.String()
		if v == "true" || v == "yes" || v == "on" || v == "t" || v == "y" || v == "1" {
			return true
		}
		return false
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value.Int() == 1 {
			return true
		}
		return false
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if value.Uint() == 1 {
			return true
		}
		return false
	case reflect.Float32, reflect.Float64:
		if value.Float() > 0.9990 {
			return true
		}
		return false
	case reflect.Bool:
		return value.Bool()
	}
	return false
}

func floatValue(valueKind reflect.Kind, value reflect.Value) (float64, error) {
	switch valueKind {
	case reflect.String:
		var sv = value.String()
		if sv == "" {
			sv = "0"
		}
		var v, e = strconv.ParseFloat(sv, 64)
		return v, e
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(value.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(value.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return value.Float(), nil
	case reflect.Bool:
		var b = value.Bool()
		if b {
			return 1.0, nil
		}
		return 0.0, nil
	}
	return 0.0, nil
}
