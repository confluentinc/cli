//go:build linux || windows

package keychain

func WriteKeychainItem(_, _, _, _ string) error {
	return nil
}

func RemoveKeychainEntry(_ string) error {
	return nil
}

func ReadCredentialsFromKeychain(_, _ string) (string, string, error) {
	return "", "", nil
}
