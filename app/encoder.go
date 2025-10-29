package main

import (
	"fmt"
	"reflect"
	"strconv"
)

func EncodeStruct(val any) []byte {
	result := []byte{}
	t := reflect.TypeOf(val)
	v := reflect.ValueOf(val)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		switch field.Type.Name() {
		case "string":
			result = append(result, EncodeString(value.String())...)
		case "int":
			result = append(result, EncodeInt(value.Int())...)
		}
	}

	return result
}

func EncodeField(field reflect.StructField, value reflect.Value) []byte {
	result := []byte{}

	return result
}

func EncodeString(value string) []byte {
	len := strconv.Itoa(len(value))
	return []byte(len + ":" + value)
}

func EncodeInt(value int64) []byte {
	return []byte("i" + fmt.Sprintf("%d", value) + "e")
}
