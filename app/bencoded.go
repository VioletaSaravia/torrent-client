package main

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
)

func BencodeStruct(val any) []byte {
	result := []byte{}
	t := reflect.TypeOf(val)
	if t.Kind() != reflect.Struct {
		return nil
	}
	v := reflect.ValueOf(val)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		value := v.Field(i)

		result = append(result, EncodeString(field.Name)...)

		switch field.Type.Kind() {
		case reflect.String:
			result = append(result, EncodeString(value.String())...)
		case reflect.Int:
			result = append(result, EncodeSigned(value.Int())...)
		case reflect.Uint8:
			result = append(result, EncodeUnsigned(uint8(value.Uint()))...)
		case reflect.Slice:
			if field.Type.Elem().Kind() != reflect.Uint8 {
				result = append(result, EncodeSlice(field.Type, value)...)
			} else {
				result = append(result, EncodeString(string(value.Interface().([]uint8)))...)
			}
		case reflect.Struct:
			result = append(result, BencodeStruct(value.Interface())...)
		default:
			fmt.Printf("field not encoded: %s\n", field.Type.Kind())
		}
	}

	return result
}

func EncodeSlice(info reflect.Type, value reflect.Value) []byte {
	result := []byte{'l'}
	switch info.Elem().Kind() {
	case reflect.String:
		val := value.Interface().([]string)
		for _, i := range val {
			result = append(result, EncodeString(i)...)
		}
	case reflect.Int:
		val := value.Interface().([]int)
		for _, i := range val {
			result = append(result, EncodeSigned(i)...)
		}
	case reflect.Uint8:
		val := value.Interface().([]uint8)
		result = append(result, EncodeString(string(val))...)
	case reflect.Slice:
		val := value.Interface()
		sliceType := reflect.TypeOf(val)

		if sliceType.Elem().Kind() != reflect.Uint8 {
			result = append(result, EncodeSlice(sliceType, reflect.ValueOf(val))...)
		} else {
			for _, i := range val.([][]uint8) {
				result = append(result, EncodeString(string(i))...)
			}
		}
	case reflect.Struct:
		val := value.Interface()
		result = append(result, BencodeStruct(val)...)
	default:
		fmt.Printf("slice of %s not encoded\n", info.Elem())
	}

	result = append(result, 'e')
	return result
}

func EncodeString(value string) []byte {
	len := strconv.Itoa(len(value))
	return []byte(len + ":" + value)
}

func EncodeSigned[T int64 | int32 | int](value T) []byte {
	return []byte("i" + fmt.Sprintf("%d", value) + "e")
}

func EncodeUnsigned[T uint8 | uint64](value T) []byte {
	return []byte("i" + fmt.Sprintf("%d", value) + "e")
}

func DecodeString(value []byte) any {
	d := Decoder{input: value}
	if decoded, err := d.Parse(); err != nil {
		log.Fatal(err)
	} else {
		return decoded
	}

	return nil
}

func DecodeFile(path string) map[string]any {
	bencoded, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	d := Decoder{input: bencoded}
	decoded, err := d.Parse()
	if err != nil {
		log.Fatal(err)
	}

	return decoded.(map[string]any)
}

type Decoder struct {
	input []byte
	cur   int
}

func (d *Decoder) Parse() (result any, err error) {
	switch d.input[d.cur] {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return d.parseStr()
	case 'i':
		return d.parseInt()
	case 'd':
		return d.parseDict()
	case 'l':
		return d.parseList()
	default:
		return "", fmt.Errorf("unidentified value: %c", d.input[0])
	}
}

func (d *Decoder) parseDict() (result map[string]any, err error) {
	d.cur += 1
	result = make(map[string]any)

	for d.input[d.cur] != 'e' {
		key, keyErr := d.parseStr()
		if keyErr != nil {
			return nil, keyErr
		}

		val, e := d.Parse()
		if e != nil {
			return nil, e
		}

		result[string(key)] = val
	}

	d.cur += 1
	return
}

func (d *Decoder) parseStr() (result []byte, err error) {
	div := d.cur

	for i := d.cur; i < len(d.input); i++ {
		if d.input[i] == ':' {
			div = i
			break
		}
	}

	lengthStr := d.input[d.cur:div]

	length, e := strconv.Atoi(string(lengthStr))
	if e != nil {
		return nil, e
	}

	result, err = d.input[div+1:div+1+length], nil
	d.cur = div + length + 1
	return
}

func (d *Decoder) parseInt() (result int, err error) {
	div := d.cur + 1

	for i := d.cur; i < len(d.input); i++ {
		if d.input[i] == 'e' {
			div = i
			break
		}
	}

	result, err = strconv.Atoi(string(d.input[d.cur+1 : div]))
	d.cur = div + 1
	return
}

func (d *Decoder) parseList() (result []any, err error) {
	result = []any{}
	d.cur += 1

	for d.input[d.cur] != 'e' {
		if element, err := d.Parse(); err != nil {
			return nil, err
		} else {
			result = append(result, element)
		}
	}

	d.cur += 1
	return
}
