package main

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
