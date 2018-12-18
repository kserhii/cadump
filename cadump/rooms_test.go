package cadump_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gocql/gocql"

	"cadump/cadump"
)

// ----- Test vars ---

func str2uuid(uuid string) gocql.UUID {
	cqlUUID, _ := gocql.ParseUUID(uuid)
	return cqlUUID
}

func str2date(date string) time.Time {
	goDate, _ := time.Parse("2006-01-02", date)
	return goDate
}

func json2str(data map[string]string) string {
	strJson, _ := json.Marshal(data)
	return string(strJson)
}

func scanDataRow() cadump.ScanDataTable {
	return cadump.ScanDataTable{
		AuxDataFuid:     str2uuid("00000000-1111-2222-3333-444444444444"),
		AuxDataName:     "FPBS Kolasin",
		AuxDataProvider: "marriott",
		Availability:    "",
		CIDate:          str2date("2019-01-18"),
		CODate:          str2date("2019-01-20"),
		ShownPrice:      map[string]string{"1": "100", "2": "101", "3": "102"},
		Currency:        "eur",
		SnapshotURL:     []string{"https://s3.amazonaws.com/img/fpbs_test.png"},
		ExtData: map[string]string{
			"aux_data_customer_hotel_id": "TGDFP",
			"room_name":                  json2str(map[string]string{"1": "Standard Room", "2": "Twin Room", "3": "Queen Room"}),
			"rate_name":                  json2str(map[string]string{"1": "No breakfast", "2": "Breakfast", "3": "Member"}),
			"tab_name":                   json2str(map[string]string{"1": "Standard Rates", "2": "Standard Rates", "3": "Prepay and Save"}),
		},
	}

}

func rooms() []cadump.Room {
	one, two, three := uint(1), uint(2), uint(3)
	return []cadump.Room{
		{
			HotelName:   "FPBS Kolasin",
			HotelCode:   "TGDFP",
			CIDate:      "18/01/2019",
			LOS:         2,
			Channel:     "Marriott",
			RoomName:    "Standard Room",
			ProductNum:  &one,
			Rate:        "100",
			Currency:    "EUR",
			Description: "No breakfast",
			TabName:     "Standard Rates",
			Snapshot:    "https://s3.amazonaws.com/img/fpbs_test.png"},
		{
			HotelName:   "FPBS Kolasin",
			HotelCode:   "TGDFP",
			CIDate:      "18/01/2019",
			LOS:         2,
			Channel:     "Marriott",
			RoomName:    "Twin Room",
			ProductNum:  &two,
			Rate:        "101",
			Currency:    "EUR",
			Description: "Breakfast",
			TabName:     "Standard Rates",
			Snapshot:    "https://s3.amazonaws.com/img/fpbs_test.png"},
		{
			HotelName:   "FPBS Kolasin",
			HotelCode:   "TGDFP",
			CIDate:      "18/01/2019",
			LOS:         2,
			Channel:     "Marriott",
			RoomName:    "Queen Room",
			ProductNum:  &three,
			Rate:        "102",
			Currency:    "EUR",
			Description: "Member",
			TabName:     "Prepay and Save",
			Snapshot:    "https://s3.amazonaws.com/img/fpbs_test.png"},
	}

}

// ----- Tests -----

func TestExtractRooms(t *testing.T) {
	expRooms := rooms()

	sd := scanDataRow()
	rooms, err := cadump.ExtractRooms(sd)

	ok(t, err)

	for i := 0; i < 3; i++ {
		equals(t, *rooms[i].ProductNum, *expRooms[i].ProductNum)
		rooms[i].ProductNum = expRooms[i].ProductNum
	}

	equals(t, expRooms, rooms)
}

func TestExtractRooms_NotAvailable(t *testing.T) {
	sd := scanDataRow()
	sd.Availability = "Not available"
	expRooms := []cadump.Room{
		{
			HotelName: "FPBS Kolasin",
			HotelCode: "TGDFP",
			CIDate:    "18/01/2019",
			LOS:       2,
			Channel:   "Marriott",
			Currency:  "EUR",
			Snapshot:  "https://s3.amazonaws.com/img/fpbs_test.png"},
	}

	rooms, err := cadump.ExtractRooms(sd)
	ok(t, err)
	equals(t, expRooms, rooms)
}
