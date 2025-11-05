package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

type TorrentInfo struct {
	Length      int
	Name        string
	PieceLength int
	Pieces      []byte
}

type MetaInfo struct {
	Announce  string
	CreatedBy string
	Info      []TorrentInfo
}

func NewMetaInfo(metainfo map[string]any) (MetaInfo, bool) {
	info := metainfo["info"].(map[string]any)

	name, nameOk := info["name"].([]byte)
	length, lengthOk := info["length"].(int)
	piece_length, pieceOk := info["piece length"].(int)
	pieces, piecesOk := info["pieces"].([]byte)

	if !(lengthOk && nameOk && pieceOk && piecesOk) {
		return MetaInfo{}, false
	}

	return MetaInfo{
		Announce:  string(metainfo["announce"].([]byte)),
		CreatedBy: string(metainfo["created by"].([]byte)),
		Info: []TorrentInfo{{
			Length:      length,
			Name:        string(name),
			PieceLength: piece_length,
			Pieces:      pieces,
		}},
	}, true
}

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
		params := url.Values{}

		hashed := sha1.Sum(BencodeStruct(info.Info[0]))
		params.Add("info_hash", string(hashed[:]))
		params.Add("peer_id", "12345678901234567890")
		params.Add("port", "6881")
		params.Add("uploaded", "0")
		params.Add("downloaded", "0")
		params.Add("compact", "1")
		params.Add("left", fmt.Sprintf("%d", info.Info[0].Length))

		requestUrl := info.Announce + "?" + params.Encode()
		if resp, err := http.Get(requestUrl); err != nil {
			log.Fatal(err)
		} else {
			body := make([]byte, resp.ContentLength)
			if _, err := resp.Body.Read(body); err != io.EOF {
				log.Fatal(err)
			}
			fmt.Println("Response:", string(body))
		}
	default:
		log.Fatal("Unknown command: " + command)
	}
}
