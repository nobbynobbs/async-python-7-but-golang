package internal

import (
	"sync"
)

type Container interface {
	Contains(lat, lng float64) bool
}

type BaseBBox struct {
	EastLng  float64 `json:"east_lng"`
	WestLng  float64 `json:"west_lng"`
	NorthLat float64 `json:"north_lat"`
	SouthLat float64 `json:"south_lat"`
}

type BBox struct {
	BaseBBox
	mu sync.RWMutex
}

func (b *BBox) Update(newBbox *BaseBBox) {
	b.mu.Lock()
	b.NorthLat = newBbox.NorthLat
	b.SouthLat = newBbox.SouthLat
	b.EastLng = newBbox.EastLng
	b.WestLng = newBbox.WestLng
	b.mu.Unlock()
}

func (b *BBox) Contains(lat, lng float64) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	// it works for north eastern hemisphere only
	return b.WestLng < lng && lng < b.EastLng && b.SouthLat < lat && lat < b.NorthLat
}

type BusStorage interface {
	Add(*Bus)
	GetList(Container) []*Bus
}

type MapBasedBusStorage struct {
	mu    sync.RWMutex
	Buses map[string]*Bus
}

func (s *MapBasedBusStorage) Add(b *Bus) {
	s.mu.Lock()
	s.Buses[b.Id] = b
	s.mu.Unlock()
}

func (s *MapBasedBusStorage) GetList(c Container) []*Bus {
	res := make([]*Bus, 0)
	s.mu.RLock()
	for _, bus := range s.Buses {
		if c.Contains(bus.Lat, bus.Lng) {
			res = append(res, bus)
		}
	}
	s.mu.RUnlock()
	return res
}
