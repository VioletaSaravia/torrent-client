package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

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
