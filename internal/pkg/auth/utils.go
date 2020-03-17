package auth

import (
	"fmt"

	"github.com/atrox/homedir"
	"github.com/bgentry/go-netrc/netrc"

	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

var (
	netrcfile = "~/.netrc"
	netrcErrorString = "Unable to get credentials from Netrc file: %s"
)

type netrcHandler struct {
	fileName string
}

func (n *netrcHandler) getNetrcCredentials(ctxName string) (email string, password string, err error){
	filename, err := homedir.Expand(n.fileName)
	if err != nil {
		err = fmt.Errorf("an error resolving the Netrc filepath at %s has occurred. ", filename)
		return "", "", err
	}
	machine, err := netrc.FindMachine(filename, ctxName)
	if err != nil {
		return "", "", err
	}
	if machine == nil {
		return "", "", errors.Errorf("Login credential not in netrc file.")
	}
	return machine.Login, machine.Password, nil
}

func updateContext(ctx *v3.Context, token string) error {
	ctx.State.AuthToken = token
	err := ctx.Save()
	if err != nil {
		return err
	}
	return nil
}
