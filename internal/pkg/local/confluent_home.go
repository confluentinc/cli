//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../../../mock/confluent_home.go --pkg mock --selfpkg github.com/confluentinc/cli confluent_home.go ConfluentHome

package local

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-version"
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
	versionFiles = map[string]string{
		"Confluent Platform":           "share/java/kafka-connect-replicator/connect-replicator-*.jar",
		"Confluent Community Software": "share/java/confluent-common/common-config-*.jar",
		"kafka":                        "share/java/kafka/kafka-clients-*.jar",
		"zookeeper":                    "share/java/kafka/zookeeper-*.jar",
	}
)

type ConfluentHome interface {
	FindFile(pattern string) ([]string, error)
	GetConfig(service string) ([]byte, error)
	GetConnectPluginPath() (string, error)
	GetConnectorConfigFile(connector string) (string, error)
	GetScriptFile(service string) (string, error)
	GetKafkaScriptFile(mode, format string) (string, error)
	GetACLCLIFile() (string, error)
	GetVersion(service string) (string, error)
	IsConfluentPlatform() (bool, error)
}

type ConfluentHomeManager struct{}

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

func (ch *ConfluentHomeManager) getRootDir() (string, error) {
	if dir := os.Getenv("CONFLUENT_HOME"); dir != "" {
		return dir, nil
	}

	return "", fmt.Errorf("set environment variable CONFLUENT_HOME")
}

func (ch *ConfluentHomeManager) getConfigFile(service string) (string, error) {
	if service == "ksql-server" {
		isKsqlDB, err := ch.isAboveVersion("5.5")
		if err != nil {
			return "", err
		}
		if !isKsqlDB {
			return "etc/ksql/ksql-server.properties", nil
		}
	}

	return ch.getFile(filepath.Join("etc", serviceConfigs[service]))
}

func (ch *ConfluentHomeManager) GetConfig(service string) ([]byte, error) {
	file, err := ch.getConfigFile(service)
	if err != nil {
		return []byte{}, err
	}

	return ioutil.ReadFile(file)
}

func (ch *ConfluentHomeManager) GetConnectPluginPath() (string, error) {
	dir, err := ch.getRootDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(dir, "/share/java")
	return path, nil
}

func (ch *ConfluentHomeManager) GetConnectorConfigFile(connector string) (string, error) {
	return ch.getFile(filepath.Join("etc", connectorConfigs[connector]))
}

func (ch *ConfluentHomeManager) GetScriptFile(service string) (string, error) {
	return ch.getFile(filepath.Join("bin", scripts[service]))
}

func (ch *ConfluentHomeManager) GetKafkaScriptFile(format, mode string) (string, error) {
	var script string

	if format == "json" || format == "protobuf" {
		supported, err := ch.isAboveVersion("5.5")
		if err != nil {
			return "", err
		}
		if !supported {
			return "", fmt.Errorf("format %s is not supported in this version", format)
		}
	}

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

	return ch.getFile(filepath.Join("bin", script))
}

func (ch *ConfluentHomeManager) GetACLCLIFile() (string, error) {
	dir, err := ch.getRootDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "bin", "sr-acl-cli"), nil
}

func (ch *ConfluentHomeManager) GetVersion(service string) (string, error) {
	pattern, ok := versionFiles[service]
	if !ok {
		return ch.getConfluentVersion()
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

func (ch *ConfluentHomeManager) getConfluentVersion() (string, error) {
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

func (ch *ConfluentHomeManager) isAboveVersion(targetVersion string) (bool, error) {
	confluentVersion, err := ch.getConfluentVersion()
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

func (ch *ConfluentHomeManager) getFile(file string) (string, error) {
	dir, err := ch.getRootDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, file), nil
}
