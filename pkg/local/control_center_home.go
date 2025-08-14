package local

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	scriptsC3 = map[string]string{
		"prometheus":     "prometheus-%s",
		"alertmanager":   "alertmanager-%s",
		"control-center": "control-center-%s",
	}
	serviceConfigsC3 = map[string]string{
		"control-center": "confluent-control-center/control-center-local.properties",
		"prometheus":     "confluent-control-center/prometheus-generated-local.yml",
		"alertmanager":   "confluent-control-center/alertmanager-generated-local.yml",
	}
	servicePortKeysC3 = map[string]string{
		"control-center": "listeners",
		"prometheus":     "listeners",
		"alertmanager":   "listeners",
	}
)

type ConfluentControlCenter interface {
	GetC3File(path ...string) (string, error)
	GetServiceScriptC3(action, service string) (string, error)
	ReadServiceConfigC3(service string) ([]byte, error)
	ReadServicePortC3(service string, zookeeperMode bool) (int, error)
}

type ControlCenterHomeManager struct{}

func NewControlCenterHomeManager() *ControlCenterHomeManager {
	return new(ControlCenterHomeManager)
}

func (c3h *ControlCenterHomeManager) getRootDir() (string, error) {
	if dir := os.Getenv("CONTROL_CENTER_HOME"); dir != "" {
		return dir, nil
	}

	return "", fmt.Errorf("set environment variable CONTROL_CENTER_HOME")
}

func (c3h *ControlCenterHomeManager) GetC3File(path ...string) (string, error) {
	dir, err := c3h.getRootDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, filepath.Join(path...)), nil
}

func (c3h *ControlCenterHomeManager) GetServiceScriptC3(action, service string) (string, error) {
	return c3h.GetC3File("bin", fmt.Sprintf(scriptsC3[service], action))
}

func (c3h *ControlCenterHomeManager) ReadServiceConfigC3(service string) ([]byte, error) {
	file, err := c3h.GetC3File("etc", serviceConfigsC3[service])
	if err != nil {
		return []byte{}, err
	}

	return os.ReadFile(file)
}

func (c3h *ControlCenterHomeManager) ReadServicePortC3(service string, zookeeperMode bool) (int, error) {
	data, err := c3h.ReadServiceConfigC3(service)
	if err != nil {
		return 0, err
	}

	config := ExtractConfig(data)
	key := servicePortKeysC3[service]
	val, ok := config[key]
	if !ok {
		return 0, fmt.Errorf("no port specified")
	}

	x := strings.Split(val.(string), ":")
	val = x[len(x)-1]

	port, err := strconv.Atoi(val.(string))
	if err != nil {
		return 0, err
	}

	return port, nil
}
