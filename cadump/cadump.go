package cadump

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/integrii/flaggy"
	"github.com/op/go-logging"
)

const version = "1.0.0"
const description = `is Cassandra scan results processing script. 
It extracts the results of the scan from Cassandra 'scan_data' table with specified scan_id.
Script splits those results by rooms and also counts the number of rooms for each hotel and provider.
All results saved in temp files and uploaded on FTP.
`

const logLevel = "INFO"
const logFormat = `%{color}%{time:2006-01-02 15:04:05.000} %{level:.4s} ▶ %{color:reset}%{message}`

var scanTimestamp = time.Now().Format("2006_01_02-15_04_05")

// ----- Logger -----

var log = logging.MustGetLogger("cadump")

func initLogger(level string) {
	logLev, err := logging.LogLevel(level)
	if err != nil {
		logLev = logging.INFO
	}

	backend := logging.AddModuleLevel(
		logging.NewBackendFormatter(
			logging.NewLogBackend(os.Stderr, "", 0),
			logging.MustStringFormatter(logFormat)))
	backend.SetLevel(logLev, "")

	log.SetBackend(backend)
}

// ----- Parse args -----

func parseArgs() (configFile string, scanIDs []uint, err error) {
	flaggy.SetName("cadump")
	flaggy.SetDescription(description)
	flaggy.SetVersion(version)

	flaggy.String(&configFile, "c", "config", "Project YAML configuration file")
	flaggy.UIntSlice(&scanIDs, "s", "sid", "Scan ID to process (can to set multiple values)")

	flaggy.Parse()

	if configFile == "" {
		err = fmt.Errorf("configuration YAML file not set")
	}
	if len(scanIDs) == 0 {
		err = fmt.Errorf("scan id not set")
	}

	return
}

// ----- Process data -----

// ProcessScan is a main function of the project
func ProcessScan() {
	var roomFiles []string
	initLogger(logLevel)

	cnfFile, scanIDs, err := parseArgs()
	checkFatalError("Arguments parse error", err)

	config, err := LoadConfig(cnfFile)
	checkFatalError("Load config error", err)

	csvSaver := SaveToCSV
	if config.CompressCSV {
		csvSaver = SaveToCSVZipped
	}

	aggregator := NewAggregator()
	db := NewCassandraReader(config.Cassandra.Hosts, config.Cassandra.Keyspace)

	log.Infof("Start Scan Data processing for Scan IDs [%s]\n", scanIDsStr(scanIDs, ", "))
	for _, scanID := range scanIDs {
		rooms, err := processScanData(scanID, db, aggregator)
		checkFatalError(fmt.Sprintf("Process Scan Data [%d] error", scanID), err)
		if len(rooms) == 0 {
			continue
		}

		sort.Slice(rooms, roomsSortFn(rooms))

		channel := rooms[0].Channel
		log.Infof("Saving rooms to CSV file (scan id: %d, channel: %s)", scanID, channel)
		roomFileName := filepath.Join(
			config.TMPFolder, fmt.Sprintf("rooms-%s-%s-%d.csv", scanTimestamp, channel, scanID))
		roomFileName, err = csvSaver(roomFileName, rooms)
		checkFatalError("Save rooms error", err)
		roomFiles = append(roomFiles, roomFileName)
		log.Infof("Rooms with channel %s(%d) saved to '%s'", channel, scanID, roomFileName)

		if config.RemoveTMPFiles {
			defer removeFile(roomFileName)
		}
	}

	log.Infof("Saving hotels counters to CSV file")

	hotelsCountsFileName := filepath.Join(
		config.TMPFolder, fmt.Sprintf("hotels_counts-%s-%s.csv", scanTimestamp, scanIDsStr(scanIDs, "_")))
	hotelsCountsFileName, err = csvSaver(hotelsCountsFileName, aggregator.HotelsCounts())
	checkFatalError("Save hotels counters error", err)
	log.Infof("Hotels counts saved to '%s'", hotelsCountsFileName)

	if config.RemoveTMPFiles {
		defer removeFile(hotelsCountsFileName)
	}

	if config.FTP.Host != "" {
		for _, roomFileName := range roomFiles {
			err = UploadFileToFTP(roomFileName, config.FTP.Host, config.FTP.User, config.FTP.Password)
			checkFatalError("Upload file error", err)
		}

		err = UploadFileToFTP(hotelsCountsFileName, config.FTP.Host, config.FTP.User, config.FTP.Password)
		checkFatalError("Upload file error", err)
	}
}

func processScanData(scanID uint, db *CassandraReader, aggregator *Aggregator) ([]Room, error) {
	var count uint = 0
	var tableRow ScanDataTable
	var allRooms []Room

	iter, err := db.SelectScanData(scanID, &tableRow)
	if err != nil {
		return allRooms, fmt.Errorf("select scan_data error: %s", err)
	}

	defer func(i SelectIter) {
		err := i.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(iter)

	for iter.Next() {
		rooms, err := ExtractRooms(tableRow)
		if err != nil {
			return allRooms, fmt.Errorf("parse rooms error: %s", err)
		}
		if len(rooms) == 0 {
			continue
		}
		if len(rooms) == 1 && rooms[0].Rate == "" {
			// skip unavailable hotels
			continue
		}
		allRooms = append(allRooms, rooms...)
		aggregator.AddRooms(rooms)

		count++
		if count%100 == 0 {
			log.Infof("[%d] => processed %d rows", scanID, count)
		}
	}
	log.Infof("[ScanID: %d] Processed %d rows. Extracted %d rooms",
		scanID, count, len(allRooms))

	return allRooms, nil
}

// ----- Helpers -----

func checkFatalError(prefix string, err error) {
	if err != nil {
		log.Fatalf("%s: %s", prefix, err)
	}
}

func scanIDsStr(sids []uint, sep string) string {
	var sidsStr []string
	for _, sid := range sids {
		sidsStr = append(sidsStr, fmt.Sprint(sid))
	}
	return strings.Join(sidsStr, sep)
}

func removeFile(file string) {
	log.Infof("Removing file '%s'", file)
	err := os.Remove(file)
	checkFatalError(fmt.Sprintf("Remove file '%s' error", file), err)
}

// cmpDate revert date for sort (31/12/2018 -> 20181231)
func cmpDate(date string) string {
	n := strings.Split(date, "/")
	if len(n) == 3 {
		return fmt.Sprintf("%s%s%s", n[2], n[1], n[0])
	}
	return date
}
