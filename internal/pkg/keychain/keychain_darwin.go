//go:build darwin
// +build darwin

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
	isCloud     = true
	sep         = "?"
)

func WriteKeychainItem(ctxName, url, username, password string) error { // this happens after login, so you can have the ctxName
	service := netrc.GetLocalCredentialName(isCloud, ctxName)

	item := keychain.NewGenericPassword(service, url, username+"-"+url, []byte(username+sep+password), accessGroup) // save username as the label
	item.SetSynchronizable(keychain.SynchronizableNo)
	item.SetAccessible(keychain.AccessibleWhenUnlocked)

	err := keychain.AddItem(item)
	if err == keychain.ErrorDuplicateItem {
		fmt.Println("already exist such entry")
		return nil
	}

	fmt.Println("saved item using:", service, username)
	return err
}

func RemoveKeychainEntry(ctxName string) error { // this also happens before logout, so you have context!
	service := netrc.GetLocalCredentialName(isCloud, ctxName)

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
			err := keychain.DeleteGenericPasswordItem(service, entry.Account)
			if err != nil {
				return err
			}
			log.CliLogger.Warnf(errors.RemoveKeychainCredentialsMsg, service) // more detailed? not the whole account name?
			fmt.Println("removed your entry:", service)
			break
		}
	}
	return nil
}

func ReadCredentialsFromKeychain(ctxName, url string) (string, string, error) { // before log in, there's no context name. How do we look for an entry? // url is all you have...
	// what if after login? when we have a context name?
	// search for multiple and return first?
	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetAccount(url)
	item.SetAccessGroup(accessGroup)
	item.SetReturnData(true)

	if ctxName != "" { // TODO: test what is ctx is nil. can it be deleted correctly? what's the behavior of normal confluent?
		service := netrc.GetLocalCredentialName(isCloud, ctxName)
		item.SetService(service)
	}

	results, err := keychain.QueryItem(item)
	if err != nil {
		return "", "", err
	} else {
		if len(results) > 0 {
			fmt.Println("found entries #", len(results), ". returning the first one:", string(results[0].Data))
			return parseCredentialsFromKeychain(results[0].Data)
		} else {
			fmt.Println("didn't found any useful entry..")
			return "", "", nil
		}
	}
}

func parseCredentialsFromKeychain(data []byte) (string, string, error) {
	substrings := strings.Split(string(data), sep)
	return substrings[0], substrings[1], nil
}
