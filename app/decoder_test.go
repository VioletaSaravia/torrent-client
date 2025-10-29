package main

import (
	_ "embed"
	"reflect"
	"testing"
)

func TestBencodedInt(t *testing.T) {
	cases := []struct {
		input    []byte
		expected any
	}{
		{[]byte("i42e"), 42},
		{[]byte("i-999e"), -999},
		{[]byte("i0000e"), 0},
		{[]byte("i-0e"), 0},
		{[]byte("i999999999999999999999999e"), nil},
		{[]byte("i0.9e"), nil},
		{[]byte("i42.0e"), nil},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			d := Decoder{input: c.input}
			result, err := d.Parse()

			if c.expected == nil {
				if err == nil {
					t.Error(result)
				}
			} else {
				if c.expected != result {
					t.Error(err)
				}
			}
		})
	}
}

func TestBencodedString(t *testing.T) {
	cases := []struct {
		input    []byte
		expected any
	}{
		{[]byte("2:la"), "la"},
		{[]byte("4:blab"), "blab"},
		{[]byte("0:asd"), ""},
		{[]byte("2:helloimtoolong"), "he"},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			d := Decoder{input: c.input}
			result, err := d.Parse()

			if c.expected == nil {
				if err == nil {
					t.Error(result)
				}
			} else {
				if c.expected != result {
					t.Error(err)
				}
			}
		})
	}
}

func TestBencodedList(t *testing.T) {
	cases := []struct {
		name     string
		input    []byte
		expected []any
	}{
		{"example", []byte("l5:helloi52ee"), []any{"hello", 52}},
		{"empty list", []byte("le"), []any{}},
		{"list of strings", []byte("l1:a1:be"), []any{"a", "b"}},
		{"list of ints", []byte("li1ei2ei3ee"), []any{1, 2, 3}},
		{"nested list", []byte("ll5:helloi42eee"), []any{[]any{"hello", 42}}},
		{"list with list", []byte("l4:spaml1:a1:bee"), []any{"spam", []any{"a", "b"}}},
		{"list with empty string", []byte("l0:1:ae"), []any{"", "a"}},
		// {"deeply nested list", "lllee", []any{[]any{[]any{}}}},
		// {"unterminated list", "l5:helloi52e", nil},
		{"list with bad integer", []byte("li0.4ee"), nil},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := Decoder{input: c.input}
			result, err := d.parseList()

			if c.expected == nil {
				if err == nil {
					t.Error(result)
				}
			} else {
				if !reflect.DeepEqual(c.expected, result) {
					t.Error(err)
				}
			}
		})
	}
}

//go:embed tests/sample.torrent
var sample_torrent []byte

func TestBencodedDictionary(t *testing.T) {
	cases := []struct {
		name     string
		input    []byte
		expected map[string]any
	}{
		{"example", []byte("d3:foo3:bar5:helloi52ee"), map[string]any{"foo": "bar", "hello": 52}},
		{"sample.torrent",
			sample_torrent,
			map[string]any{
				"announce":   "http://bittorrent-test-tracker.codecrafters.io/announce",
				"created by": "mktorrent 1.1",
				"info": map[string]any{
					"name":         "sample.txt",
					"length":       92063,
					"piece length": 32768,
					"pieces":       "\xe8v\xf6z*\x88\x86\xe8\xf3k\x13g&\xc3\x0f\xa2\x97\x03\x02-n\"u\xe6\x04\xa0vfVsn\x81\xff\x10\xb5R\x04\xad\x8d5\xf0\r\x93z\x02\x13\xdf\x19\x82\xbc\x8d\tr'\xad\x9e\x90\x9a\xcc\x17"},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := Decoder{input: c.input}
			result, err := d.parseDict()

			if c.expected == nil {
				if err == nil {
					t.Error(result)
				}
			} else {
				if !reflect.DeepEqual(c.expected, result) {
					t.Error(err)
				}
			}
		})
	}
}
