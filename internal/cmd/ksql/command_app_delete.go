package ksql

import (
	"context"
	"io/ioutil"
	"net/http"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/dghubble/sling"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *appCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id>",
		Short:             "Delete a ksqlDB app.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              pcmd.NewCLIRunE(c.delete),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *appCommand) delete(cmd *cobra.Command, args []string) error {
	id := args[0]

	req := &schedv1.KSQLCluster{
		AccountId: c.EnvironmentId(),
		Id:        id,
	}

	// Check KSQL exists
	cluster, err := c.Client.KSQL.Describe(context.Background(), req)
	if err != nil {
		return errors.CatchKSQLNotFoundError(err, id)
	}

	// Terminated cluster needs to also be sent to KSQL cluster to clean up internal topics of the KSQL
	if cluster.Status == schedv1.ClusterStatus_UP {
		ctx := c.Config.Context()
		state, err := ctx.AuthenticatedState()
		if err != nil {
			return err
		}

		bearerToken, err := pauth.GetBearerToken(state, ctx.Platform.Server, cluster.Id)
		if err != nil {
			return err
		}

		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: bearerToken})
		client := sling.New().Client(oauth2.NewClient(context.Background(), ts)).Base(cluster.Endpoint)
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

	if err := c.Client.KSQL.Delete(context.Background(), req); err != nil {
		return err
	}

	c.analyticsClient.SetSpecialProperty(analytics.ResourceIDPropertiesKey, id)
	utils.Printf(cmd, errors.KsqlDBDeletedMsg, args[0])
	return nil
}
