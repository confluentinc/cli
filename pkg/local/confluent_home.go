//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../../../mock/confluent_home.go --pkg mock --selfpkg github.com/confluentinc/cli/v3 confluent_home.go ConfluentHome

package local

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hashicorp/go-version"
)

/*
Directory Structure:

CONFLUENT_HOME/
	bin/
	etc/
	examples/
	share/
*/

var (
	scripts = map[string]string{
		"connect":         "connect-distributed",
		"control-center":  "control-center-%s",
		"kafka":           "kafka-server-%s",
		"kafka-rest":      "kafka-rest-%s",
		"ksql-server":     "ksql-server-%s",
		"schema-registry": "schema-registry-%s",
		"zookeeper":       "zookeeper-server-%s",
	}
	serviceConfigs = map[string]string{
		"connect":         "schema-registry/connect-avro-distributed.properties",
		"control-center":  "confluent-control-center/control-center-dev.properties",
		"kafka":           "kafka/server.properties",
		"kafka-rest":      "kafka-rest/kafka-rest.properties",
		"ksql-server":     "ksqldb/ksql-server.properties",
		"schema-registry": "schema-registry/schema-registry.properties",
		"zookeeper":       "kafka/zookeeper.properties",
	}
	servicePortKeys = map[string]string{
		"connect":         "rest.port",
		"control-center":  "listeners",
		"kafka":           "listeners",
		"kafka-rest":      "listeners",
		"ksql-server":     "listeners",
		"schema-registry": "listeners",
		"zookeeper":       "clientPort",
	}
	connectorConfigs = map[string]string{
		"elasticsearch-sink": "kafka-connect-elasticsearch/quickstart-elasticsearch.properties",
		"file-sink":          "kafka/connect-file-sink.properties",
		"file-source":        "kafka/connect-file-source.properties",
		"hdfs-sink":          "kafka-connect-hdfs/quickstart-hdfs.properties",
		"jdbc-sink":          "kafka-connect-jdbc/sink-quickstart-sqlite.properties",
		"jdbc-source":        "kafka-connect-jdbc/source-quickstart-sqlite.properties",
		"s3-sink":            "kafka-connect-s3/quickstart-s3.properties",
	}
	versionFiles = map[string]string{
		"Confluent Platform":           "share/java/kafka-connect-replicator/connect-replicator-*.jar",
		"Confluent Community Software": "share/java/confluent-common/common-config-*.jar",
		"kafka":                        "share/java/kafka/kafka-clients-*.jar",
		"zookeeper":                    "share/java/kafka/zookeeper-*.jar",
	}
)

type ConfluentHome interface {
	GetFile(path ...string) (string, error)
	HasFile(path ...string) (bool, error)
	FindFile(pattern string) ([]string, error)

	IsConfluentPlatform() (bool, error)
	GetConfluentVersion() (string, error)
	IsAtLeastVersion(targetVersion string) (bool, error)

	GetServiceScript(action, service string) (string, error)
	ReadServiceConfig(service string) ([]byte, error)
	ReadServicePort(service string) (int, error)
	GetVersion(service string) (string, error)

	GetConnectorConfigFile(connector string) (string, error)
	GetKafkaScript(mode, format string) (string, error)
}

type ConfluentHomeManager struct{}

func NewConfluentHomeManager() *ConfluentHomeManager {
	return new(ConfluentHomeManager)
}

func (ch *ConfluentHomeManager) getRootDir() (string, error) {
	if dir := os.Getenv("CONFLUENT_HOME"); dir != "" {
		return dir, nil
	}

	return "", fmt.Errorf("set environment variable CONFLUENT_HOME")
}

func (ch *ConfluentHomeManager) GetFile(path ...string) (string, error) {
	dir, err := ch.getRootDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, filepath.Join(path...)), nil
}

func (ch *ConfluentHomeManager) HasFile(path ...string) (bool, error) {
	file, err := ch.GetFile(path...)
	if err != nil {
		return false, err
	}

	return exists(file), nil
}

func (ch *ConfluentHomeManager) FindFile(pattern string) ([]string, error) {
	dir, err := ch.getRootDir()
	if err != nil {
		return []string{}, err
	}

	path := filepath.Join(dir, pattern)

	matches, err := filepath.Glob(path)
	if err != nil {
		return []string{}, err
	}

	for i := range matches {
		matches[i], err = filepath.Rel(dir, matches[i])
		if err != nil {
			return []string{}, err
		}
	}
	return matches, nil
}

func (ch *ConfluentHomeManager) IsConfluentPlatform() (bool, error) {
	controlCenter := "share/java/confluent-control-center/control-center-*.jar"
	files, err := ch.FindFile(controlCenter)
	if err != nil {
		return false, err
	}
	return len(files) > 0, nil
}

func (ch *ConfluentHomeManager) GetConfluentVersion() (string, error) {
	isCP, err := ch.IsConfluentPlatform()
	if err != nil {
		return "", err
	}

	if isCP {
		return ch.GetVersion("Confluent Platform")
	} else {
		return ch.GetVersion("Confluent Community Software")
	}
}

func (ch *ConfluentHomeManager) GetServiceScript(action, service string) (string, error) {
	if service == "connect" {
		if action == "start" {
			return ch.GetFile("bin", scripts[service])
		} else {
			return "", nil
		}
	}
	return ch.GetFile("bin", fmt.Sprintf(scripts[service], action))
}

func (ch *ConfluentHomeManager) ReadServiceConfig(service string) ([]byte, error) {
	file, err := ch.GetFile("etc", serviceConfigs[service])
	if err != nil {
		return []byte{}, err
	}

	if service == "ksql-server" {
		isKsqlDB, err := ch.IsAtLeastVersion("5.5")
		if err != nil {
			return []byte{}, err
		}
		if !isKsqlDB {
			file, err = ch.GetFile("etc/ksql/ksql-server.properties")
			if err != nil {
				return []byte{}, err
			}
		}
	}

	return os.ReadFile(file)
}

func (ch *ConfluentHomeManager) ReadServicePort(service string) (int, error) {
	data, err := ch.ReadServiceConfig(service)
	if err != nil {
		return 0, err
	}

	config := ExtractConfig(data)

	key := servicePortKeys[service]
	val, ok := config[key]
	if !ok {
		return 0, fmt.Errorf("no port specified")
	}

	if key == "listeners" {
		x := strings.Split(val.(string), ":")
		val = x[len(x)-1]
	}

	port, err := strconv.Atoi(val.(string))
	if err != nil {
		return 0, err
	}

	return port, nil
}

func (ch *ConfluentHomeManager) GetVersion(service string) (string, error) {
	pattern, ok := versionFiles[service]
	if !ok {
		return ch.GetConfluentVersion()
	}

	matches, err := ch.FindFile(pattern)
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("could not find %s in CONFLUENT_HOME", pattern)
	}

	versionFile := matches[0]
	x := strings.Split(pattern, "*")
	prefix, suffix := x[0], x[1]
	return versionFile[len(prefix) : len(versionFile)-len(suffix)], nil
}

func (ch *ConfluentHomeManager) GetConnectorConfigFile(connector string) (string, error) {
	return ch.GetFile("etc", connectorConfigs[connector])
}

func (ch *ConfluentHomeManager) GetKafkaScript(format, mode string) (string, error) {
	var script string

	switch format {
	case "":
		script = fmt.Sprintf("kafka-console-%s", mode)
	case "avro":
		script = fmt.Sprintf("kafka-avro-console-%s", mode)
	case "json":
		script = fmt.Sprintf("kafka-json-schema-console-%s", mode)
	case "protobuf":
		script = fmt.Sprintf("kafka-protobuf-console-%s", mode)
	default:
		return "", fmt.Errorf("invalid format: %s", format)
	}

	return ch.GetFile("bin", script)
}

func (ch *ConfluentHomeManager) IsAtLeastVersion(targetVersion string) (bool, error) {
	confluentVersion, err := ch.GetConfluentVersion()
	if err != nil {
		return false, err
	}

	a, err := version.NewSemver(confluentVersion)
	if err != nil {
		return false, err
	}

	b, err := version.NewSemver(targetVersion)
	if err != nil {
		return false, err
	}

	return a.Compare(b) >= 0, nil
}
