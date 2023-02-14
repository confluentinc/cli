//go:build linux || windows

package keychain

func Write(isCloud bool, ctxName, url, username, password string) error {
	return nil
}

func Delete(isCloud bool, ctxName string) error {
	return nil
}

func Read(isCloud bool, ctxName, url string) (string, string, error) {
	return "", "", nil
}
