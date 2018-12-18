package cadump

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ----- Room row -----

// Room is single room row structure
type Room struct {
	HotelName   string `csv:"Hotel name"`
	HotelCode   string `csv:"Hotel Code"`
	CIDate      string `csv:"CI date"`
	LOS         uint   `csv:"LOS"`
	Channel     string `csv:"Channel"`
	RoomName    string `csv:"Room name"`
	ProductNum  *uint  `csv:"Product #"`
	Rate        string `csv:"Rate"`
	Currency    string `csv:"Currency"`
	Description string `csv:"Description"`
	TabName     string `csv:"Tab name"`
	Snapshot    string `csv:"Snapshot"`
}

func roomsSortFn(rooms []Room) func(int, int) bool {
	// sort by: HotelName, CIDate, LOS, Channel, ProductNum
	return func(i, j int) bool {
		r1, r2 := rooms[i], rooms[j]

		if r1.HotelName == r2.HotelName {
			if r1.CIDate == r2.CIDate {
				if r1.LOS == r2.LOS {
					if r1.Channel == r2.Channel {
						return *r1.ProductNum < *r2.ProductNum
					}
					return r1.Channel < r2.Channel
				}
				return r1.LOS < r2.LOS
			}
			return cmpDate(r1.CIDate) < cmpDate(r2.CIDate)
		}
		return r1.HotelName < r2.HotelName
	}
}

// ----- Rooms extractor -----

// ExtractRooms return array of the rooms from single DB scan row
func ExtractRooms(scanData ScanDataTable) ([]Room, error) {
	var rooms []Room

	snapshot := ""
	if scanData.SnapshotURL != nil && len(scanData.SnapshotURL) > 0 {
		snapshot = scanData.SnapshotURL[0]
	}

	hotel := Room{
		HotelName: scanData.AuxDataName,
		HotelCode: scanData.ExtData["aux_data_customer_hotel_id"],
		CIDate:    scanData.CIDate.Format("02/01/2006"),
		LOS:       uint(scanData.CODate.Sub(scanData.CIDate).Hours() / 24),
		Channel:   strings.Title(scanData.AuxDataProvider),
		Currency:  strings.ToUpper(scanData.Currency),
		Snapshot:  snapshot}

	if scanData.Availability == "Not available" {
		rooms = append(rooms, hotel)
		return rooms, nil
	}

	roomName, err := unpackExtDataField(scanData.ExtData, "room_name", false)
	if err != nil {
		return rooms, fmt.Errorf("[%s] ExtData 'room_name' parse error: %s",
			scanData.AuxDataFuid.String(), err)
	}

	description, err := unpackExtDataField(scanData.ExtData, "rate_name", true)
	if err != nil {
		description, err = unpackExtDataField(scanData.ExtData, "description", false)
		if err != nil {
			return rooms, fmt.Errorf("[%s] ExtData 'rate_name' or 'description' parse error: %s",
				scanData.AuxDataFuid.String(), err)
		}
	}

	tabName, err := unpackExtDataField(scanData.ExtData, "tab_name", true)
	if err != nil {
		return rooms, fmt.Errorf("[%s] ExtData 'tab_name' parse error: %s",
			scanData.AuxDataFuid.String(), err)
	}

	for numKey := range scanData.ShownPrice {
		prodNum, err := strToUInt(numKey)
		if err != nil {
			return rooms, fmt.Errorf("product number \"%s\" parse error: %s", numKey, err)
		}

		room := hotel
		room.ProductNum = prodNum
		room.Rate = scanData.ShownPrice[numKey]
		room.RoomName = roomName[numKey]
		room.Description = description[numKey]
		room.TabName = tabName[numKey]

		rooms = append(rooms, room)
	}

	return rooms, nil
}

// ----- Helpers -----

func unpackExtDataField(extData map[string]string, fieldName string, optional bool) (map[string]string, error) {
	var fieldValue map[string]string

	value, ok := extData[fieldName]
	if !ok {
		if optional {
			return fieldValue, nil
		}
		return fieldValue, fmt.Errorf("field not found")
	}

	if value == "" {
		return fieldValue, fmt.Errorf("field is empty")
	}

	err := json.Unmarshal([]byte(value), &fieldValue)
	if err != nil {
		return fieldValue, fmt.Errorf("field JSON unmarshalling error: %s", err)
	}

	return fieldValue, nil
}

func strToUInt(value string) (*uint, error) {
	numInt, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}
	numUInt := uint(numInt)
	return &numUInt, nil
}
