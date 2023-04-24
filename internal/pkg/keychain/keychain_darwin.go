//go:build darwin

package keychain

import (
	"fmt"
	"strings"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"github.com/keybase/go-keychain"
)

const (
	accessGroup = "cli"
	separator   = "?"
)

func Write(isCloud bool, ctxName, url, username, password string) error {
	service := netrc.GetLocalCredentialName(isCloud, false, ctxName)

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
	service := netrc.GetLocalCredentialName(isCloud, false, ctxName)

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
			err := keychain.DeleteGenericPasswordItem(service, entry.Account)
			if err != nil {
				return err
			}

			log.CliLogger.Warnf(errors.RemoveKeychainCredentialsMsg, username)
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
		service := netrc.GetLocalCredentialName(isCloud, false, ctxName)
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
		return "", "", errors.New(errors.ParseKeychainCredentialsErrorMsg)
	}
	return substrings[0], substrings[1], nil
}
