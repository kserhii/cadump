package cadump

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jlaffaye/ftp"
)

// UploadFileToFTP upload file to FTP server
func UploadFileToFTP(filePath string, host string, user string, password string) (err error) {
	log.Infof("Connecting to FTP %s ...", host)

	conn, err := ftp.Connect(fmt.Sprintf("%s:%s", host, "21"))
	if err != nil {
		return fmt.Errorf("FTP '%s' open connection error: %s", host, err)
	}
	defer func() {
		if cerr := conn.Quit(); cerr != nil {
			err = fmt.Errorf("FTP '%s' close connection error: %s", host, cerr)
		}
	}()

	err = conn.Login(user, password)
	if err != nil {
		return fmt.Errorf("FTP '%s' login error: %s", host, err)
	}

	log.Infof("Got FTP connection to '%s'", host)

	inFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file '%s' error: %s", filePath, err)
	}

	defer func() {
		if ferr := inFile.Close(); ferr != nil {
			err = fmt.Errorf("close file '%s' error: %s", filePath, ferr)
		}
	}()

	fileOnFTP := filepath.Base(filePath)
	log.Infof("Uploading file '%s' to FTP '%s'", filePath, fileOnFTP)
	err = conn.Stor(fileOnFTP, inFile)
	if err != nil {
		return fmt.Errorf("upload file on FTP '%s' error: %s", host, err)
	}

	log.Infof("File '%s' saved on FTP", filePath)
	return nil
}
