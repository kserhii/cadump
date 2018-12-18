package cadump

import (
	"archive/zip"
	"fmt"
	"os"
	"strings"

	"github.com/gocarina/gocsv/v2"
)

// ----- CSV Writer -----

func SaveToCSV(filePath string, rows interface{}) (savedFile string, err error) {
	savedFile = filePath

	outFile, err := os.Create(filePath)
	if err != nil {
		err = fmt.Errorf("create CSV file '%s' error: %s", filePath, err)
		return
	}

	defer func() {
		if ferr := outFile.Close(); ferr != nil {
			err = fmt.Errorf("close CSV file '%s' error: %s", filePath, ferr)
		}
	}()

	err = gocsv.MarshalFile(rows, outFile)
	if err != nil {
		err = fmt.Errorf("serilize rooms into CSV file '%s' error: %s", filePath, err)
		return
	}

	return
}

func SaveToCSVZipped(filePath string, rows interface{}) (savedFile string, err error) {
	var originFileName, zipFileName string

	if strings.HasSuffix(filePath, ".zip") {
		originFileName = filePath[:len(filePath)-4]
		zipFileName = filePath
	} else {
		originFileName = filePath
		zipFileName = fmt.Sprintf("%s.zip", filePath)
	}
	savedFile = zipFileName

	outFile, err := os.Create(zipFileName)
	if err != nil {
		err = fmt.Errorf("create CSV ZIP file '%s' error: %s", zipFileName, err)
		return
	}

	defer func() {
		if ferr := outFile.Close(); ferr != nil {
			err = fmt.Errorf("close CSV ZIP file '%s' error: %s", zipFileName, ferr)
		}
	}()

	zipWriter := zip.NewWriter(outFile)
	defer func() {
		if zwerr := zipWriter.Close(); zwerr != nil {
			err = fmt.Errorf("close ZIP writer error: %s", zwerr)
		}
	}()

	fileWriter, err := zipWriter.Create(originFileName)
	if err != nil {
		err = fmt.Errorf("write CSV ZIP file error: %s", err)
		return
	}

	err = gocsv.Marshal(rows, fileWriter)
	if err != nil {
		err = fmt.Errorf("serilize rooms into CSV ZIP file '%s' error: %s", zipFileName, err)
		return
	}

	return
}
