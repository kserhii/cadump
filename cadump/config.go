package cadump

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// ----- Config -----

const configExample = `
TMP_FOLDER: /tmp
REMOVE_TMP_FILES: true
COMPRESS_CSV: true

CASSANDRA:
    hosts:
      - cassandra-host1
      - cassandra-host2
    keyspace: some_key_space

FTP: 
    host: files.net
    user: user
    password: pass 
`

type Config struct {
	TMPFolder      string `yaml:"TMP_FOLDER"`
	RemoveTMPFiles bool   `yaml:"REMOVE_TMP_FILES"`
	CompressCSV    bool   `yaml:"COMPRESS_CSV"`

	Cassandra struct {
		Hosts    []string `yaml:"hosts"`
		Keyspace string   `yaml:"keyspace"`
	} `yaml:"CASSANDRA"`

	FTP struct {
		Host     string `yaml:"host"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"FTP"`
}

func LoadConfig(cnfFile string) (Config, error) {
	config := Config{}
	errHelp := fmt.Sprintf(
		"Please create configuration YAML file according to this template:\n%s",
		configExample)

	data, err := ioutil.ReadFile(cnfFile)
	if err != nil {
		return config, fmt.Errorf("read config error: %s\n%s", err, errHelp)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("parse config error: %s\n%s", err, errHelp)
	}

	if len(config.Cassandra.Hosts) == 0 || config.Cassandra.Keyspace == "" {
		return config, fmt.Errorf("missing required CASSANDRA fields\n%s", errHelp)
	}

	return config, nil
}
