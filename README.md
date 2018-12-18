# Cassandra Dumper

### Description

Cassandra Dumper (CaDump) is Cassandra scan results processing script.
It extracts the results of the scan from Cassandra 'scan_data' table with specified scan_id.
Script splits those results by rooms and also counts the number of rooms for each hotel and provider.
All results saved in temp files and uploaded on FTP.

Project programming language is [Go 1.1](https://blog.golang.org/go1.11).

### How to run

##### Configuration

CaDump script requires configuration [YAML](http://yaml.org/) file with all settings.

*Example of config file*:

```yaml
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
```

Field `CASSANDRA` is required. All other fields are optional.
Default `TMP_FOLDER` is a folder where the script is placed.
Default `REMOVE_TMP_FILES` and `COMPRESS_CSV` values are false.
If `FTP.host` not set, files will not be uploaded to FTP.

##### Run commands

Usage:

```bash
./cadump [-h] [--config cnf.yaml] [--sid 42] [--sid 43]
```

You can specify as many scan ids (sid) as you need.

Run dev scan example (used flags shortcut):

```bash
./cadump -c dev.yaml -s 229261 -s 229262 -s 229263 
```
 
Script version:
```bash
./cadump --version 
```

### Build and deploy

CaDump distributed as single script file that can be copied to the server and ready to execute.
`Makefile` contains commands to compile and build `cabump` script.
Results of the script generation will be in the `build` folder.

```bash
make build
```

### Local development and testing

##### Install environment

To make an update you need to install Go lang v1.11 first (see [GoLang install](https://golang.org/doc/install)).
Then download project out of GOPATH dir (project uses go modules).

##### Local run

```bash
go run cmd/cadump/main.go --config local.yaml --sid 90210 --sid 90211
```

##### Tests

Run tests:

```bash
make test
```

Run tests with coverage:

```bash
make test-cov
```

Clean after tests:

```bash
make clean
```
