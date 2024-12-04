package entities

import "time"

// https://developer.valvesoftware.com/wiki/DEM_(file_format)
type CSReplayFileHeader struct {
	Filestamp       string        `json:"filestamp" bson:"filestamp"`
	Protocol        int           `json:"protocol" bson:"protocol"`
	NetworkProtocol int           `json:"network_protocol" bson:"network_protocol"`
	ServerName      string        `json:"server_name" bson:"server_name"`
	ClientName      string        `json:"client_name" bson:"client_name"`
	MapName         string        `json:"map_name" bson:"map_name"`
	Length          time.Duration `json:"length" bson:"length"`
	Ticks           int           `json:"ticks" bson:"ticks"`
	Frames          int           `json:"frames" bson:"frames"`
}
