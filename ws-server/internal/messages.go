package internal

type Bus struct {
	Id    string  `json:"busId"`
	Route string  `json:"route"`
	Lat   float64 `json:"lat"`
	Lng   float64 `json:"lng"`
}

type BaseMessage struct {
	MsgType string `json:"msgType"`
}

type BusesListMessage struct {
	BaseMessage
	Buses []*Bus `json:"buses"`
}

type ErrorMessage struct {
	BaseMessage
	Errors []string
}

type BBoxMessage struct {
	BaseMessage
	Data *BaseBBox `json:"data"`
}
