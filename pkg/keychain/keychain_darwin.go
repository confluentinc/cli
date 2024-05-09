//go:build darwin

package keychain

import (
	"fmt"
	"strings"

	"github.com/keybase/go-keychain"

	"github.com/confluentinc/cli/v3/pkg/log"
)

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
	accessGroup                  = "cli"
	separator                    = "?"
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

func Write(isCloud bool, ctxName, url, username, password string) error {
	service := GetLocalCredentialName(isCloud, ctxName)

	item := keychain.NewGenericPassword(service, url, fmt.Sprintf("%s-%s", username, url), []byte(username+separator+password), accessGroup)
	item.SetSynchronizable(keychain.SynchronizableNo)
	item.SetAccessible(keychain.AccessibleWhenUnlocked)

	err := keychain.AddItem(item)
	if err == keychain.ErrorDuplicateItem {
		return nil
	}

	return err
}

func Delete(isCloud bool, ctxName string) error {
	service := GetLocalCredentialName(isCloud, ctxName)

	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetService(service)
	item.SetAccessGroup(accessGroup)
	item.SetReturnData(true)

	results, err := keychain.QueryItem(item)
	if err != nil {
		return err
	}

	for _, entry := range results {
		if strings.Contains(ctxName, entry.Account) {
			username, _, _ := parseCredentialsFromKeychain(entry.Data)
			if err := keychain.DeleteGenericPasswordItem(service, entry.Account); err != nil {
				return err
			}

			log.CliLogger.Warnf(`Removed credentials for user "%s" from keychain`, username)
			break
		}
	}
	return nil
}

func Read(isCloud bool, ctxName, url string) (string, string, error) {
	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetAccount(url)
	item.SetAccessGroup(accessGroup)
	item.SetReturnData(true)

	if ctxName != "" {
		service := GetLocalCredentialName(isCloud, ctxName)
		item.SetService(service)
	}

	results, err := keychain.QueryItem(item)
	if err == nil {
		if len(results) > 0 {
			return parseCredentialsFromKeychain(results[0].Data)
		}
	}

	return "", "", err
}

func parseCredentialsFromKeychain(data []byte) (string, string, error) {
	substrings := strings.Split(string(data), separator)
	if len(substrings) < 2 {
		return "", "", fmt.Errorf("unable to parse credentials in keychain access")
	}
	return substrings[0], substrings[1], nil
}

func GetLocalCredentialName(isCloud bool, ctxName string) string {
	credType := mdsUsernamePassword
	if isCloud {
		credType = ccloudUsernamePassword
	}
	return fmt.Sprintf(localCredentialStringFormat, credType.String(), ctxName)
}
