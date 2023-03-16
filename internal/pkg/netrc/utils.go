package netrc

import (
	"strings"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

type MachineContextInfo struct {
	CredentialType string
	Username       string
	URL            string
	CaCertPath     string
}

func ParseNetrcMachineName(machineName string) (*MachineContextInfo, error) {
	if !strings.HasPrefix(machineName, localCredentialsPrefix) {
		return nil, errors.New("Incorrect machine name format")
	}

	// example: machinename = confluent-cli:ccloud-username-password:login-caas-team+integ-cli@confluent.io-https://devel.cpdev.cloud
	credTypeAndContextNameString := suffixFromIndex(machineName, len(localCredentialsPrefix)+1)

	// credTypeAndContextName = ccloud-username-password:login-caas-team+integ-cli@confluent.io-https://devel.cpdev.cloud
	credType, contextNameString, err := extractCredentialType(credTypeAndContextNameString)
	if err != nil {
		return nil, err
	}

	// contextNameString = login-caas-team+integ-cli@confluent.io-https://devel.cpdev.cloud
	username, url, caCertPath, err := parseContextName(contextNameString)
	if err != nil {
		return nil, err
	}

	// username = caas-team+integ-cli@confluent.io
	// url = https://devel.cpdev.cloud
	return &MachineContextInfo{
		CredentialType: credType,
		Username:       username,
		URL:            url,
		CaCertPath:     caCertPath,
	}, nil
}

func extractCredentialType(nameSubstring string) (string, string, error) {
	var credType string
	if strings.HasPrefix(nameSubstring, mdsUsernamePasswordString) {
		credType = mdsUsernamePasswordString
	} else if strings.HasPrefix(nameSubstring, ccloudUsernamePasswordString) {
		credType = ccloudUsernamePasswordString
	} else {
		return "", "", errors.New("incorrect machine name format")
	}
	// +1 to remove the character ":"
	rest := suffixFromIndex(nameSubstring, len(credType)+1)
	return credType, rest, nil
}

func parseContextName(nameSubstring string) (string, string, string, error) {
	contextNamePrefix := "login-"
	if !strings.HasPrefix(nameSubstring, contextNamePrefix) {
		return "", "", "", errors.New("incorrect context name format")
	}

	contextName := suffixFromIndex(nameSubstring, len(contextNamePrefix))

	urlIndex := strings.Index(contextName, "http")

	// -1 to exclude "-"
	username := prefixToIndex(contextName, urlIndex-1)

	// +1 to exclude "-"
	rest := suffixFromIndex(contextName, len(username)+1)

	url := rest
	caCertPath := ""

	if idx := strings.Index(rest, "?"); idx != -1 {
		url = prefixToIndex(rest, idx)
		caCertPath = suffixFromIndex(rest, idx+len("cacertpath")+2)
	}

	return username, url, caCertPath, nil
}

func suffixFromIndex(s string, index int) string {
	runes := []rune(s)
	return string(runes[index:])
}

func prefixToIndex(s string, index int) string {
	runes := []rune(s)
	return string(runes[:index])
}
