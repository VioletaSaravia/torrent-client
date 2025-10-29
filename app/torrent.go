package main

type TorrentInfo struct {
	length       int
	name         string
	piece_length int
	pieces       []byte
}

type MetaInfo struct {
	announce   string
	created_by string
	info       []TorrentInfo
}

func NewMetaInfo(metainfo map[string]any) (MetaInfo, bool) {
	info := metainfo["info"].(map[string]any)

	name, nameOk := info["name"].([]byte)
	length, lengthOk := info["length"].(int)
	piece_length, pieceOk := info["piece length"].(int)
	pieces_str, piecesOk := info["pieces"].([]byte)
	pieces := []byte(pieces_str)

	if !(lengthOk && nameOk && pieceOk && piecesOk) {
		return MetaInfo{}, false
	}

	return MetaInfo{
		announce:   string(metainfo["announce"].([]byte)),
		created_by: string(metainfo["created by"].([]byte)),
		info: []TorrentInfo{{
			length:       length,
			name:         string(name),
			piece_length: piece_length,
			pieces:       pieces,
		}},
	}, true
}
