package cadump_test

import (
	"testing"

	"cadump/cadump"
)

// ----- Test vars ---

var testRoom1 = cadump.Room{HotelName: "Beverly Hills", HotelCode: "BH-19210", CIDate: "31/12/2018"}
var testRoom2 = cadump.Room{HotelName: "Hotel California", HotelCode: "HC1980", CIDate: "10/11/2018"}
var testHCount1 = cadump.HotelCounts{
	HotelName: "Beverly Hills", HotelCode: "BH-19210", CIDate: "31/12/2018",
	Marriott: 0, Booking: 0, Expedia: 0, Ctrip: 0, Priceline: 0}
var testHCount2 = cadump.HotelCounts{
	HotelName: "Hotel California", HotelCode: "HC1980", CIDate: "10/11/2018",
	Marriott: 0, Booking: 0, Expedia: 0, Ctrip: 0, Priceline: 0}

func newRoom(room cadump.Room, channel string) cadump.Room {
	room.Channel = channel
	return room
}

func Room1(channel string) cadump.Room {
	return newRoom(testRoom1, channel)
}

func Room2(channel string) cadump.Room {
	return newRoom(testRoom2, channel)
}

// ----- Tests -----

func TestNewAggregator(t *testing.T) {
	agg := cadump.NewAggregator()
	counts := agg.HotelsCounts()
	equals(t, make([]cadump.HotelCounts, 0), counts)
}

func TestAggregator_AddRoom_NewOrExist(t *testing.T) {
	hCount1, hCount2 := testHCount1, testHCount2

	agg := cadump.NewAggregator()
	equals(t, []cadump.HotelCounts{}, agg.HotelsCounts())

	agg.AddRoom(Room1("Booking"))
	hCount1.Booking++
	equals(t, []cadump.HotelCounts{hCount1}, agg.HotelsCounts())

	agg.AddRoom(Room1("Booking"))
	hCount1.Booking++
	equals(t, []cadump.HotelCounts{hCount1}, agg.HotelsCounts())

	agg.AddRoom(Room2("Booking"))
	hCount2.Booking++
	equals(t, []cadump.HotelCounts{hCount1, hCount2}, agg.HotelsCounts())
}

func TestAggregator_AddRoom_AllChannels(t *testing.T) {
	hCounts1, hCounts2 := testHCount1, testHCount2
	agg := cadump.NewAggregator()

	for _, chName := range []string{"Marriott", "Booking", "Expedia", "Ctrip", "Priceline"} {
		agg.AddRoom(Room1(chName))
	}

	hCounts1.Marriott++
	hCounts1.Booking++
	hCounts1.Expedia++
	hCounts1.Ctrip++
	hCounts1.Priceline++

	equals(t, []cadump.HotelCounts{hCounts1}, agg.HotelsCounts())

	for _, chName := range []string{"Marriott", "Booking", "Expedia"} {
		agg.AddRoom(Room2(chName))
	}

	hCounts2.Marriott++
	hCounts2.Booking++
	hCounts2.Expedia++

	equals(t, []cadump.HotelCounts{hCounts1, hCounts2}, agg.HotelsCounts())

	for i := 0; i < 10; i++ {
		agg.AddRoom(Room2("Priceline"))
		hCounts2.Priceline++
	}

	equals(t, []cadump.HotelCounts{hCounts1, hCounts2}, agg.HotelsCounts())

	agg.AddRoom(Room1("unknown"))

	equals(t, []cadump.HotelCounts{hCounts1, hCounts2}, agg.HotelsCounts())
}

func TestAggregator_AddRooms(t *testing.T) {
	var rooms []cadump.Room
	hCounts1, hCounts2 := testHCount1, testHCount2
	agg := cadump.NewAggregator()

	for _, chName := range []string{"Marriott", "Booking", "Expedia", "Ctrip", "Priceline", "Unknown"} {
		rooms = append(rooms, Room1(chName), Room2(chName))
	}

	hCounts1.Marriott++
	hCounts1.Booking++
	hCounts1.Expedia++
	hCounts1.Ctrip++
	hCounts1.Priceline++

	hCounts2.Marriott++
	hCounts2.Booking++
	hCounts2.Expedia++
	hCounts2.Ctrip++
	hCounts2.Priceline++

	agg.AddRooms(rooms)

	equals(t, []cadump.HotelCounts{hCounts1, hCounts2}, agg.HotelsCounts())
}

func TestAggregator_HotelsCounts_Sort(t *testing.T) {
	agg := cadump.NewAggregator()

	rooms := []cadump.Room{
		{HotelName: "Dubrova house", HotelCode: "AGG", CIDate: "20/01/2018", Channel: "Marriott"},
		{HotelName: "AbuDabi hotel", HotelCode: "ZA-42", CIDate: "31/02/2018", Channel: "Marriott"},
		{HotelName: "AbuDabi hotel", HotelCode: "ZA-42", CIDate: "01/05/2018", Channel: "Marriott"},
		{HotelName: "Bee house", HotelCode: "BZB-ksv", CIDate: "31/02/2018", Channel: "Marriott"},
	}
	agg.AddRooms(rooms)
	counts := agg.HotelsCounts()

	equals(t, []cadump.HotelCounts{
		{HotelName: "AbuDabi hotel", HotelCode: "ZA-42", CIDate: "31/02/2018", Marriott: 1},
		{HotelName: "AbuDabi hotel", HotelCode: "ZA-42", CIDate: "01/05/2018", Marriott: 1},
		{HotelName: "Bee house", HotelCode: "BZB-ksv", CIDate: "31/02/2018", Marriott: 1},
		{HotelName: "Dubrova house", HotelCode: "AGG", CIDate: "20/01/2018", Marriott: 1},
	}, counts)
}
