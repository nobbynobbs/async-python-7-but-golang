package internal

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type WebclientsServer struct {
	BusStorage BusStorage
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *WebclientsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("cannot upgrade connection", err)
		return
	}
	defer ws.Close()

	log.Println("connection accepted")
	defer log.Println("connection closed")

	bbox := &BBox{
		BaseBBox: BaseBBox{
			EastLng:  -180,
			WestLng:  -180,
			NorthLat: -90,
			SouthLat: -90,
		},
	}

	// accept messages to update bounds
	go func() {
		log.Println("receiving messages...")
		defer log.Println("receiver stopped...")
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				break
			}

			newBbox := &BBoxMessage{}
			err = json.Unmarshal(message, newBbox)
			if err != nil {
				resp := ErrorMessage{
					BaseMessage: BaseMessage{MsgType: "Error"},
					Errors:      []string{"invalid json"},
				}
				_ = ws.WriteJSON(resp)
				continue
			}
			log.Printf("reseived new message %+v", *newBbox.Data)
			bbox.Update(newBbox.Data)
		}
	}()

	// send responses
	for {
		buses := s.BusStorage.GetList(bbox)
		resp := BusesListMessage{
			BaseMessage: BaseMessage{"Buses"},
			Buses:       buses,
		}
		err := ws.WriteJSON(resp)
		if err != nil {
			break
		}
		time.Sleep(300 * time.Millisecond)
	}
}

type BusesServer struct {
	BusStorage BusStorage
}

func (s *BusesServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("cannot upgrade connection", err)
		return
	}
	defer ws.Close()

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			break
		}
		busInfo := &Bus{}
		err = json.Unmarshal(message, busInfo)
		if err != nil {
			log.Println("unable to parse JSON")
			continue
		}
		//log.Printf("received bus info: %+v", busInfo)
		s.BusStorage.Add(busInfo)
	}
}
