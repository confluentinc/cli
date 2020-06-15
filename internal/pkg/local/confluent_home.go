//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../../../mock/confluent_home.go --pkg mock --selfpkg github.com/confluentinc/cli confluent_home.go ConfluentHome

package local

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

/*
Directory Structure:

CONFLUENT_HOME/
	bin/
	etc/
	share/
*/

var (
	scripts = map[string]string{
		"connect":         "connect-distributed",
		"control-center":  "control-center-start",
		"kafka":           "kafka-server-start",
		"kafka-rest":      "kafka-rest-start",
		"ksql-server":     "ksql-server-start",
		"schema-registry": "schema-registry-start",
		"zookeeper":       "zookeeper-server-start",
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
	connectorConfigs = map[string]string{
		"elasticsearch-sink": "kafka-connect-elasticsearch/quickstart-elasticsearch.properties",
		"file-sink":          "kafka/connect-file-sink.properties",
		"file-source":        "kafka/connect-file-source.properties",
		"hdfs-sink":          "kafka-connect-hdfs/quickstart-hdfs.properties",
		"jdbc-sink":          "kafka-connect-jdbc/sink-quickstart-sqlite.properties",
		"jdbc-source":        "kafka-connect-jdbc/source-quickstart-sqlite.properties",
		"s3-sink":            "kafka-connect-s3/quickstart-s3.properties",
	}
)

type ConfluentHome interface {
	IsConfluentPlatform() (bool, error)
	FindFile(pattern string) ([]string, error)
	GetConfig(service string) ([]byte, error)
	GetConnectorConfigFile(connector string) (string, error)
	GetScriptFile(service string) (string, error)
}

type ConfluentHomeManager struct {}

func NewConfluentHomeManager() *ConfluentHomeManager {
	return new(ConfluentHomeManager)
}

func (ch *ConfluentHomeManager) IsConfluentPlatform() (bool, error) {
	controlCenter := "share/java/confluent-control-center/control-center-*.jar"
	files, err := ch.FindFile(controlCenter)
	if err != nil {
		return false, err
	}
	return len(files) > 0, nil
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

func (ch *ConfluentHomeManager) getConfigFile(service string) (string, error) {
	// TODO: Return ksql/ksql-server.properties for older versions
	return ch.getFile(filepath.Join("etc", serviceConfigs[service]))
}

func (ch *ConfluentHomeManager) GetConfig(service string) ([]byte, error) {
	file, err := ch.getConfigFile(service)
	if err != nil {
		return []byte{}, err
	}

	return ioutil.ReadFile(file)
}

func (ch *ConfluentHomeManager) GetConnectorConfigFile(connector string) (string, error) {
	return ch.getFile(filepath.Join("etc", connectorConfigs[connector]))
}

func (ch *ConfluentHomeManager) GetScriptFile(service string) (string, error) {
	return ch.getFile(filepath.Join("bin", scripts[service]))
}

func (ch *ConfluentHomeManager) getRootDir() (string, error) {
	if dir := os.Getenv("CONFLUENT_HOME"); dir != "" {
		return dir, nil
	}

	return "", fmt.Errorf("set environment variable CONFLUENT_HOME")
}

func (ch *ConfluentHomeManager) getFile(file string) (string, error) {
	dir, err := ch.getRootDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, file), nil
}
