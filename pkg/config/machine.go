package config

import "fmt"

type Machine struct {
	Name     string
	User     string
	Password string
}

type MachineParams struct {
	IgnoreCert bool
	IsCloud    bool
	Name       string
	URL        string
}

const (
	localCredentialsPrefix       = "confluent-cli"
	localCredentialStringFormat  = localCredentialsPrefix + ":%s:%s"
	mdsUsernamePasswordString    = "mds-username-password"
	ccloudUsernamePasswordString = "ccloud-username-password"
)

type credentialType int

const (
	mdsUsernamePassword credentialType = iota
	ccloudUsernamePassword
)

func (c credentialType) String() string {
	return []string{mdsUsernamePasswordString, ccloudUsernamePasswordString}[c]
}

func GetLocalCredentialName(isCloud bool, ctxName string) string {
	credType := mdsUsernamePassword
	if isCloud {
		credType = ccloudUsernamePassword
	}
	return fmt.Sprintf(localCredentialStringFormat, credType.String(), ctxName)
}
