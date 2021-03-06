package entities

type Point [2]float32

type RouteInfo struct {
	Name             string  `json:"name"`
	FirstStationName string  `json:"station_start_name"`
	LastStationName  string  `json:"station_stop_name"`
	Coordinates      []Point `json:"coordinates"`
}

type BusInfo struct {
	Id    string  `json:"busId"`
	Route string  `json:"route"`
	Lat   float32 `json:"lat"`
	Lng   float32 `json:"lng"`
}
