package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: ./app [decode|download]")
	}
	command := os.Args[1]

	switch command {
	case "decode":
		if len(os.Args) < 3 {
			log.Fatal("Usage: ./app decode [file|value]")
		}

		what := os.Args[2]
		switch what {
		case "file":
			if len(os.Args) != 4 {
				log.Fatal("Usage: ./app decode file [path]")
				return
			}
			result := DecodeFile(os.Args[3])
			fmt.Println(result)

		case "value":
			if len(os.Args) != 4 {
				log.Fatal("Usage: ./app decode value [bencoded value]")
			}
			DecodeString([]byte(os.Args[3]))

		default:
			log.Fatal("Unknown decode command: " + what)
		}
	case "download":
		if len(os.Args) < 3 {
			log.Fatal("Usage: ./app download [path]")
		}

		file := DecodeFile(os.Args[2])
		info, ok := NewMetaInfo(file)
		if ok {
			fmt.Printf("%#v\n", info)
		}
	default:
		log.Fatal("Unknown command: " + command)
	}

}
