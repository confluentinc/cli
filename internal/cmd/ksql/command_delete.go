package ksql

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/dghubble/sling"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *ksqlCommand) newDeleteCommand(isApp bool) *cobra.Command {
	shortText := "Delete a ksqlDB cluster."
	var longText string
	runCommand := c.deleteCluster
	if isApp {
		// DEPRECATED: this should be removed before CLI v3, this work is tracked in https://confluentinc.atlassian.net/browse/KCI-1411
		shortText = "DEPRECATED: Delete a ksqlDB app."
		longText = "DEPRECATED: Delete a ksqlDB app. " + errors.KSQLAppDeprecateWarning
		runCommand = c.deleteApp
	}

	cmd := &cobra.Command{
		Use:               "delete <id>",
		Short:             shortText,
		Long:              longText,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              runCommand,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *ksqlCommand) deleteCluster(cmd *cobra.Command, args []string) error {
	return c.delete(cmd, args, false)
}

func (c *ksqlCommand) deleteApp(cmd *cobra.Command, args []string) error {
	return c.delete(cmd, args, true)
}

func (c *ksqlCommand) delete(cmd *cobra.Command, args []string, isApp bool) error {
	id := args[0]
	environmentId := c.EnvironmentId()

	// Check KSQL exists
	cluster, err := c.V2Client.DescribeKsqlCluster(id, environmentId)
	if err != nil {
		return errors.CatchKSQLNotFoundError(err, id)
	}

	// Terminated cluster needs to also be sent to KSQL cluster to clean up internal topics of the KSQL
	if cluster.Status.Phase == "PROVISIONED" {
		ctx := c.Config.Context()
		state, err := ctx.AuthenticatedState()
		if err != nil {
			return err
		}

		bearerToken, err := pauth.GetBearerToken(state, ctx.Platform.Server, *cluster.Id)
		if err != nil {
			return err
		}

		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: bearerToken})
		client := sling.New().Client(oauth2.NewClient(context.Background(), ts)).Base(*cluster.Status.HttpEndpoint)
		request := make(map[string][]string)
		request["deleteTopicList"] = []string{".*"}
		response, err := client.Post("/ksql/terminate").BodyJSON(&request).ReceiveSuccess(nil)
		if err != nil {
			return err
		}

		if response.StatusCode != http.StatusOK {
			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				return err
			}
			return errors.Errorf(errors.KsqlDBTerminateClusterMsg, args[0], string(body))
		}
	}

	if err := c.V2Client.DeleteKsqlCluster(id, c.EnvironmentId()); err != nil {
		return err
	}

	if isApp {
		fmt.Fprintln(os.Stderr, errors.KSQLAppDeprecateWarning)
	}
	utils.Printf(cmd, errors.KsqlDBDeletedMsg, args[0])
	return nil
}
