package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"unicode"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

// Ensures gofmt doesn't remove the "os" encoding/json import (feel free to remove this!)
var _ = json.Marshal

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345

type Decoder struct {
	input string
	cur   int
}

func (d *Decoder) Parse() (result any, err error) {
	if unicode.IsDigit(rune(d.input[d.cur])) {
		return d.parseStr()
	} else if d.input[d.cur] == 'd' {
		return d.parseDict()
	} else if d.input[d.cur] == 'i' {
		return d.parseInt()
	} else if d.input[d.cur] == 'l' {
		return d.parseList()
	} else {
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

type TorrentInfo struct {
	length       int
	name         string
	piece_length int
	pieces       []byte
}

type MetaInfo struct {
	announce string
	info     []TorrentInfo
}

func NewMetaInfo(info map[string]any) MetaInfo {
	return MetaInfo{
		announce: info["announce"].(string),
		info: []TorrentInfo{{
			length:       info["length"].(int),
			name:         info["name"].(string),
			piece_length: info["piece_length"].(int),
			pieces:       info["pieces"].([]byte),
		}},
	}
}

func main() {
	command := os.Args[1]

	switch command {
	case "file":
		if len(os.Args) != 3 {
			log.Fatal("Usage: ./decoder file [path]")
			return
		}
		path := os.Args[2]
		bencoded, err := os.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}

		d := Decoder{input: string(bencoded)}
		decoded, err := d.Parse()
		if err != nil {
			log.Fatal(err)
		}

		info, _ := json.MarshalIndent(decoded, "", "  ")
		fmt.Println(string(info))

	case "decode":
		if len(os.Args) != 3 {
			log.Fatal("Usage: ./decoder decode [bencoded value]")
		}
		value := os.Args[2]

		d := Decoder{input: value}
		decoded, err := d.Parse()
		if err != nil {
			log.Fatal(err)
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))

	default:
		log.Fatal("Unknown command: " + command)
	}
}
