package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"

	"emulator/entities"
)

type Emulator struct {
	Ctx              context.Context
	WG               *sync.WaitGroup
	ServerURL        string
	Buses            chan entities.BusInfo
	RoutesDir        string
	RoutesFiles      []os.FileInfo
	ConnectionsCount int
	BusesPerRoute    int
}

func (e *Emulator) Run() {
	// run server clients
	for i := 0; i < e.ConnectionsCount; i++ {
		e.WG.Add(1)
		go e.sendBusInfo()
	}

	// spawning "buses"
	for i := 0; i < e.BusesPerRoute; i++ {
		for routeIndex, fileInfo := range e.RoutesFiles {
			fullPath := path.Join(e.RoutesDir, fileInfo.Name())

			f, err := os.Open(fullPath)
			if err != nil {
				log.Printf("unable to open file: %s\n", err)
				return
			}

			fileContent, err := ioutil.ReadAll(f)
			if err != nil {
				log.Printf("unable to open file: %s\n", err)
				return
			}
			_ = f.Close()

			route := entities.RouteInfo{}

			err = route.UnmarshalJSON(fileContent)
			if err != nil {
				log.Printf("unable to unmarshal json: %s\n", err)
				return
			}

			randOffset := rand.Intn(len(route.Coordinates) / 2)
			busIndex := i*e.BusesPerRoute + routeIndex
			busId := route.Name + strconv.Itoa(busIndex)

			e.WG.Add(1)
			go e.spawnBus(busId, route.Name, route.Coordinates[randOffset:])
		}
	}

	fmt.Println("waiting for all goroutines")
	e.WG.Wait()
}

// sendBusInfo - read data from channel, send data to socket
func (e *Emulator) sendBusInfo() {
	log.Println("start sender")
	defer func() {
		log.Println("stop sender")
		e.WG.Done()
	}()

	for {
		select {
		case <-e.Ctx.Done():
			return
		default:
			func() { // use closure for correct defer scope
				ws, _, err := websocket.DefaultDialer.Dial(e.ServerURL, nil)
				if err != nil {
					log.Printf("unable connect to server %s\n", err)
					return
				}
				defer ws.Close()
				e.sendForever(ws)
			}()
		}
		time.Sleep(1 * time.Second) // pause on error
	}
}

func (e *Emulator) sendForever(ws *websocket.Conn) {
	for {
		select {
		case <-e.Ctx.Done():
			return
		case bus := <-e.Buses:
			msg, err := bus.MarshalJSON()
			if err != nil {
				log.Printf("unable to serialize message %s\n", err)
				continue
			}
			if err := ws.WriteMessage(websocket.BinaryMessage, msg); err != nil {
				log.Printf("unable to send message to server %s\n", err)
				return
			}
		}
	}
}

func (e *Emulator) spawnBus(
	busId,
	route string,
	coordinates []entities.Point,
) {
	defer e.WG.Done()
	bus := entities.BusInfo{}
	for {
		for _, point := range coordinates {
			bus.Id = busId
			bus.Route = route
			bus.Lat = point[0]
			bus.Lng = point[1]

			select {
			case <-e.Ctx.Done():
				return
			case e.Buses <- bus:
				time.Sleep(300 * time.Millisecond)
			}
		}
	}
}

func main() {

	// controlling shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sig := <-interrupt
		log.Printf("caught shutdown signal: %s", sig)
		cancel()
	}()

	// get routes files info
	dir, err := os.Getwd()
	if err != nil {
		log.Printf("unable to get current directory: %s\n", err)
		return
	}

	routesDir := path.Join(dir, "routes") // hardcoded
	files, err := ioutil.ReadDir(routesDir)
	if err != nil {
		log.Printf("unable to read routes directory: %s\n", err)
		return
	}

	wg := &sync.WaitGroup{}

	app := Emulator{
		Ctx:              ctx,
		WG:               wg,
		ServerURL:        getEnv("EMULATOR_SERVER_URL", "ws://127.0.0.1:8080"),
		Buses:            make(chan entities.BusInfo),
		RoutesDir:        routesDir,
		RoutesFiles:      files,
		ConnectionsCount: getIntEnv("EMULATOR_CONNECTIONS_COUNT", 5),
		BusesPerRoute:    getIntEnv("EMULATOR_BUSES_PER_ROUTE", 20),
	}

	go http.ListenAndServe(":8088", nil) // for profiler endpoints

	app.Run()
}

// getEnv read env variable value
func getEnv(name, deflt string) string {
	val := os.Getenv(name)
	if val == "" {
		return deflt
	}
	return val
}

// getIntEnv read integer value from env variable
func getIntEnv(name string, deflt int) int {
	strVal := os.Getenv(name)
	if strVal == "" {
		return deflt
	}

	intVal, err := strconv.Atoi(strVal)
	if err != nil {
		return deflt
	}

	return intVal
}
