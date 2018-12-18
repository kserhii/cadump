package cadump

import (
	"fmt"
	"reflect"
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/qb"
)

const (
	pageSize          = 100
	defaultTimeoutSec = 300
	connAttempts      = 5
)

// ----- ScanData table -----

type ScanDataTable struct {
	AuxDataFuid     gocql.UUID        `cql:"aux_data_fuid"`
	AuxDataName     string            `cql:"aux_data_name"`
	AuxDataProvider string            `cql:"aux_data_provider"`
	Availability    string            `cql:"availability"`
	CIDate          time.Time         `cql:"ci_date"`
	CODate          time.Time         `cql:"co_date"`
	ShownPrice      map[string]string `cql:"shown_price"`
	Currency        string            `cql:"currency"`
	SnapshotURL     []string          `cql:"snapshot_url"`
	ExtData         map[string]string `cql:"ext_data"`
}

// ----- Cassandra Reader -----

// CassandraReader is Cassandra database connection config
//
// Example:
//
//	db := NewCassandraReader("cassandra-1", "keyspace")
//
//	var table ScanDataTable
//	iter, err := db.SelectScanData(90210, &table)
//	checkFatalError(err)
//
//	defer func(i SelectIter) {
//		checkFatalError(i.Close())
//	}(iter)
//
//	for iter.Next() {
//		fmt.Println(table.AuxDataName)
//	}
//
type CassandraReader struct {
	conn *gocql.ClusterConfig
}

// NewCassandraReader is CassandraReader constructor
func NewCassandraReader(hosts []string, keyspace string) *CassandraReader {
	conn := gocql.NewCluster(hosts...)
	conn.Keyspace = keyspace
	conn.Consistency = gocql.One
	conn.PageSize = pageSize
	conn.Timeout = time.Duration(defaultTimeoutSec) * time.Second

	return &CassandraReader{conn: conn}
}

// createSession make connection to the Cassandra DB with retries
func (reader *CassandraReader) createSession() (*gocql.Session, error) {
	var err error
	connErrTimeout := 1 * time.Second

	log.Infof("Connecting to Cassandra %v (keyspace: %s)...",
		reader.conn.Hosts, reader.conn.Keyspace)

	for attempt := 0; attempt <= connAttempts; attempt++ {
		session, err := reader.conn.CreateSession()
		if err == nil {
			log.Info("Got Cassandra connection")
			return session, nil
		}
		log.Warningf("Error connect to Cassandra: %s (attempt %d of %d)",
			err, attempt, connAttempts)

		if attempt < connAttempts {
			connErrTimeout *= 2
			time.Sleep(connErrTimeout)
		}
	}

	return nil, fmt.Errorf("can't connect to Cassandra %v: %s", reader.conn.Hosts, err)
}

// SelectScanData load rows from "scan_data" table without limit
func (reader *CassandraReader) SelectScanData(scanID uint, dest *ScanDataTable) (SelectIter, error) {
	return reader.SelectScanDataLimit(scanID, dest, 0)
}

// SelectScanDataLimit make query to select data from "scan_data" table with limit and map it to the dest struct
func (reader *CassandraReader) SelectScanDataLimit(scanID uint, dest *ScanDataTable, limit uint) (SelectIter, error) {
	session, err := reader.createSession()
	if err != nil {
		return SelectIter{}, err
	}

	columns := getTags(*dest, "cql")
	query := qb.Select("scan_data").Where(qb.Eq("aux_data_scan_id")).Columns(columns...)
	if limit > 0 {
		query = query.Limit(limit)
	}
	queryStr, names := query.ToCql()
	queryParams := qb.M{"aux_data_scan_id": scanID}

	log.Debugf("%s (aux_data_scan_id: %d)", queryStr, scanID)

	iterx := gocqlx.Query(session.Query(queryStr), names).BindMap(queryParams).Iter().Unsafe()

	selectIter := SelectIter{
		session: session,
		dest:    dest,
		iterx:   iterx}

	return selectIter, nil
}

// ----- Select Query Iterator -----

// SelectIter is DB Select query iterator
type SelectIter struct {
	session *gocql.Session
	dest    *ScanDataTable

	iterx *gocqlx.Iterx
}

// Next is used to iterate over query results
func (iter *SelectIter) Next() bool {
	return iter.iterx.StructScan(iter.dest)
}

// Close iteration and connect session and return error is exists
func (iter *SelectIter) Close() error {
	defer iter.session.Close()
	return iter.iterx.Close()
}

// ----- Helper -----

// getTags return list of specified tags in struct
func getTags(s interface{}, tag string) []string {
	var columns []string
	tblStructType := reflect.TypeOf(s)

	for fnum := 0; fnum < tblStructType.NumField(); fnum++ {
		field := tblStructType.Field(fnum)
		if fieldName, ok := field.Tag.Lookup(tag); ok {
			if fieldName != "" {
				columns = append(columns, fieldName)
			}
		}
	}
	return columns
}
