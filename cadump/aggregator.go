package cadump

import (
	"fmt"
	"sort"
)

type HotelCounts struct {
	HotelName string `csv:"Hotel name"`
	HotelCode string `csv:"Hotel Code"`
	CIDate    string `csv:"CI date"`
	Marriott  uint   `csv:"Marriott"`
	Booking   uint   `csv:"Booking"`
	Expedia   uint   `csv:"Expedia"`
	Ctrip     uint   `csv:"Ctrip"`
	Priceline uint   `csv:"Priceline"`
}

func hotelsCountsSortFn(counts []HotelCounts) func(int, int) bool {
	// sort by: HotelName, CIDate
	return func(i, j int) bool {
		hc1, hc2 := counts[i], counts[j]

		if hc1.HotelName == hc2.HotelName {
			return cmpDate(hc1.CIDate) < cmpDate(hc2.CIDate)
		}
		return hc1.HotelName < hc2.HotelName
	}
}

type Aggregator struct {
	hotels map[string]*HotelCounts
}

func NewAggregator() *Aggregator {
	return &Aggregator{hotels: make(map[string]*HotelCounts)}
}

func (agg *Aggregator) AddRoom(room Room) {
	var hotel *HotelCounts

	key := fmt.Sprintf("%s-%s", room.HotelCode, room.CIDate)

	hotel, exist := agg.hotels[key]
	if !exist {
		hotel = &HotelCounts{
			HotelName: room.HotelName,
			HotelCode: room.HotelCode,
			CIDate:    room.CIDate}
		agg.hotels[key] = hotel
	}

	switch room.Channel {
	case "Marriott":
		hotel.Marriott++
	case "Booking":
		hotel.Booking++
	case "Expedia":
		hotel.Expedia++
	case "Ctrip":
		hotel.Ctrip++
	case "Priceline":
		hotel.Priceline++
	default:
		log.Warningf("Unknown chanel '%s' (hotel_code: %s, CI: %s)",
			room.Channel, room.HotelCode, room.CIDate)
	}
}

func (agg *Aggregator) AddRooms(rooms []Room) {
	for _, room := range rooms {
		agg.AddRoom(room)
	}
}

func (agg *Aggregator) HotelsCounts() []HotelCounts {
	counts := make([]HotelCounts, 0, len(agg.hotels))
	for key := range agg.hotels {
		counts = append(counts, *agg.hotels[key])
	}
	sort.Slice(counts, hotelsCountsSortFn(counts))
	return counts
}
