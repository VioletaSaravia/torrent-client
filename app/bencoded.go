package main

import (
	"fmt"
	"log"
	"os"
	r "reflect"
	"regexp"
	"strconv"
	"strings"
)

func BencodeStruct(val any) []byte {
	result := []byte{}
	t := r.TypeOf(val)
	if t.Kind() != r.Struct {
		return nil
	}
	v := r.ValueOf(val)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		value := v.Field(i)

		result = append(result, EncodeString(field.Name)...)

		switch field.Type.Kind() {
		case r.String:
			result = append(result, EncodeString(value.String())...)
		case r.Int:
			result = append(result, EncodeSigned(value.Int())...)
		case r.Uint8:
			result = append(result, EncodeUnsigned(uint8(value.Uint()))...)
		case r.Slice:
			if field.Type.Elem().Kind() != r.Uint8 {
				result = append(result, EncodeSlice(field.Type, value)...)
			} else {
				result = append(result, EncodeString(string(value.Interface().([]uint8)))...)
			}
		case r.Struct:
			result = append(result, BencodeStruct(value.Interface())...)
		default:
			fmt.Printf("field not encoded: %s\n", field.Type.Kind())
		}
	}

	return result
}

func EncodeSlice(info r.Type, value r.Value) []byte {
	result := []byte{'l'}
	switch info.Elem().Kind() {
	case r.String:
		val := value.Interface().([]string)
		for _, i := range val {
			result = append(result, EncodeString(i)...)
		}
	case r.Int:
		val := value.Interface().([]int)
		for _, i := range val {
			result = append(result, EncodeSigned(i)...)
		}
	case r.Uint8:
		val := value.Interface().([]uint8)
		result = append(result, EncodeString(string(val))...)
	case r.Slice:
		val := value.Interface()
		sliceType := r.TypeOf(val)

		if sliceType.Elem().Kind() != r.Uint8 {
			result = append(result, EncodeSlice(sliceType, r.ValueOf(val))...)
		} else {
			for _, i := range val.([][]uint8) {
				result = append(result, EncodeString(string(i))...)
			}
		}
	case r.Struct:
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

	result, err := Decode(bencoded)
	if err != nil {
		log.Fatal(err)
	}
	return result.(map[string]any)
}

func Decode(input []byte) (any, error) {
	d := Decoder{input: input}
	return d.Parse()
}

var cases *regexp.Regexp = regexp.MustCompile(`([a-z])([A-Z])|([A-Z])([A-Z][a-z])`)

func titleToWords(s string) string {
	return strings.ToLower(cases.ReplaceAllString(s, `${1}${3} ${2}${4}`))
}

func DecodeInto[T any](input map[string]any) (T, bool) {
	var result T
	t := r.TypeFor[T]()
	v := r.ValueOf(&result).Elem()
	for i := range t.NumField() {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}

		name := t.Field(i).Tag.Get("bencoded")
		if name == "" {
			name = titleToWords(t.Field(i).Name)
		}

		switch field.Kind() {
		case r.Invalid:
			continue
		case r.Bool:
			if val, ok := input[name].(bool); ok {
				field.SetBool(val)
			}
		case r.Int, r.Int8, r.Int16, r.Int32, r.Int64:
			if val, ok := input[name].(int64); ok {
				field.SetInt(val)
			}
		case r.Uint, r.Uint8, r.Uint16, r.Uint32, r.Uint64:
			if val, ok := input[name].(uint64); ok {
				field.SetUint(val)
			}
		case r.Uintptr:
			continue
		case r.Float32, r.Float64:
			if val, ok := input[name].(float64); ok {
				field.SetFloat(val)
			}
		case r.Array:
			continue
		case r.Chan:
			continue
		case r.Func:
			continue
		case r.Interface:
			continue
		case r.Map:
			continue
		case r.Pointer:
			continue
		case r.Slice:
			if val, ok := input[name].([]any); ok {
				slice := r.MakeSlice(field.Type(), len(val), len(val))
			}
		case r.String:
			if val, ok := input[name].([]byte); ok {
				field.SetString(string(val))
			}
		case r.Struct:
			if val, ok := input[name].(map[string]any); ok {
				decodeInnerStruct(field, val)
			}
			continue
		case r.UnsafePointer:
			continue
		}
	}

	return result, true
}

func decodeInnerStruct(v r.Value, input map[string]any) {
	if v.Kind() != r.Struct {
		return
	}
	t := v.Type()

	for i := range t.NumField() {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}
		name := t.Field(i).Tag.Get("bencoded")
		if name == "" {
			name = titleToWords(t.Field(i).Name)
		}

		switch field.Kind() {
		case r.Invalid:
			continue
		case r.Bool:
			if val, ok := input[name].(bool); ok {
				field.SetBool(val)
			}
		case r.Int, r.Int8, r.Int16, r.Int32, r.Int64:
			if val, ok := input[name].(int64); ok {
				field.SetInt(val)
			}
		case r.Uint, r.Uint8, r.Uint16, r.Uint32, r.Uint64:
			if val, ok := input[name].(uint64); ok {
				field.SetUint(val)
			}
		case r.Uintptr:
			continue
		case r.Float32, r.Float64:
			if val, ok := input[name].(float64); ok {
				field.SetFloat(val)
			}
		case r.Array:
			continue
		case r.Chan:
			continue
		case r.Func:
			continue
		case r.Interface:
			continue
		case r.Map:
			continue
		case r.Pointer:
			continue
		case r.Slice:
			continue
		case r.String:
			if val, ok := input[name].([]byte); ok {
				field.SetString(string(val))
			}
		case r.Struct:
			if val, ok := input[name].(map[string]any); ok {
				decodeInnerStruct(field, val)
			}
			continue
		case r.UnsafePointer:
			continue
		}
	}
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
