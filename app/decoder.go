package main

import (
	"fmt"
	"strconv"
)

type Decoder struct {
	input string
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
		return "", fmt.Errorf("unsupported format: %c", d.input[0])
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

		result[key] = val
	}

	d.cur += 1
	return
}

func (d *Decoder) parseStr() (result string, err error) {
	div := d.cur

	for i := d.cur; i < len(d.input); i++ {
		if d.input[i] == ':' {
			div = i
			break
		}
	}

	lengthStr := d.input[d.cur:div]

	length, e := strconv.Atoi(lengthStr)
	if e != nil {
		return "", e
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

	result, err = strconv.Atoi(d.input[d.cur+1 : div])
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
