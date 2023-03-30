package auditlog

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	mds "github.com/confluentinc/mds-sdk-go-public/mdsv1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "audit-log",
		Aliases:     []string{"al"},
		Short:       "Manage audit log configuration.",
		Long:        "Manage which auditable events are logged, and where the event logs are sent.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

	cmd.AddCommand(newDescribeCommand(prerunner))
	cmd.AddCommand(newConfigCommand(prerunner))
	cmd.AddCommand(newRouteCommand(prerunner))

	return cmd
}

type errorMessage struct {
	ErrorCode uint32 `json:"error_code" yaml:"error_code"`
	Message   string `json:"message" yaml:"message"`
}

func HandleMdsAuditLogApiError(cmd *cobra.Command, err error, response *http.Response) error {
	if response != nil {
		switch status := response.StatusCode; status {
		case http.StatusNotFound:
			cmd.SilenceUsage = true
			return errors.NewWrapErrorWithSuggestions(err, errors.UnableToAccessEndpointErrorMsg, errors.UnableToAccessEndpointSuggestions)
		case http.StatusForbidden:
			switch e := err.(type) {
			case mds.GenericOpenAPIError:
				cmd.SilenceUsage = true
				em := errorMessage{}
				if err = json.Unmarshal(e.Body(), &em); err != nil {
					return err
				}
				return fmt.Errorf("%s\n%s", e.Error(), em.Message)
			}
		}
	}
	return err
}
