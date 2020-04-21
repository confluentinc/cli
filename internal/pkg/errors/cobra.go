package errors

import (
	"crypto/x509"
	"fmt"
	"net/url"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	"github.com/confluentinc/ccloud-sdk-go"
	corev1 "github.com/confluentinc/ccloudapis/core/v1"
	"github.com/confluentinc/mds-sdk-go"
)

var messages = map[error]string{
	ErrNoContext:      "You must log in to run that command.",
	ErrNotLoggedIn:    "You must log in to run that command.",
	ErrNotImplemented: "Sorry, this functionality is not yet available in the CLI.",
	ErrNoKafkaContext: "You must pass --cluster or set an active kafka in your context with 'kafka cluster use'",
	ErrNoKSQL:         "Could not find KSQL cluster with Resource ID specified.",
}

// HandleCommon provides standard error messaging for common errors.
func HandleCommon(err error, cmd *cobra.Command) error {
	// Give an indication of successful completion
	if err == nil {
		return nil
	}
	cmd.SilenceUsage = true

	if msg, ok := messages[err]; ok {
		return fmt.Errorf(msg)
	}

	switch e := err.(type) {
	case mds.GenericOpenAPIError:
		return fmt.Errorf(e.Error() + ": " + string(e.Body()))
	case *corev1.Error:
		var result error
		result = multierror.Append(result, e)
		for name, msg := range e.GetNestedErrors() {
			result = multierror.Append(result, fmt.Errorf("%s: %s", name, msg))
		}
		return result
	case *url.Error:
		if certErr, ok := e.Err.(x509.CertificateInvalidError); ok {
			return fmt.Errorf("%s. Check the system keystore or login again with the --ca-cert-path option to add custom certs", certErr.Error())
		}
		return e
	case *UnspecifiedAPIKeyError:
		return fmt.Errorf("no API key selected for %s, please select an api-key first (e.g., with `api-key use`)", e.ClusterID)
	case *UnspecifiedCredentialError:
		// TODO: Add more context to credential error messages (add variable error).
		return fmt.Errorf("context \"%s\" has corrupted credentials. To fix, please remove the config file, "+
			"and run `login` or `init`", e.ContextName)
	case *UnspecifiedPlatformError:

		// TODO: Add more context to platform error messages (add variable error).
		return fmt.Errorf("context \"%s\" has a corrupted platform. To fix, please remove the config file, "+
			"and run `login` or `init`", e.ContextName)
	case *ccloud.InvalidLoginError:
		return fmt.Errorf("You have entered an incorrect username or password. Please try again.")
	case *ccloud.InvalidTokenError:
		return fmt.Errorf("Your auth token has been corrupted. Please login again.")
	}

	return err
}
