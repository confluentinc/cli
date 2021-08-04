package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type promoCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func NewPromoCommand(prerunner pcmd.PreRunner) *cobra.Command {
	c := &promoCommand{
		pcmd.NewAuthenticatedCLICommand(
			&cobra.Command{
				Use:   "promo",
				Short: "Manage promo codes.",
				Args:  cobra.NoArgs,
			},
			prerunner,
		),
	}

	c.AddCommand(c.newAddCommand())
	c.AddCommand(c.newListCommand())

	return c.Command
}

func (c *promoCommand) newAddCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add <code>",
		Short: "Add a new promo code.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.addRunE),
	}
}

func (c *promoCommand) addRunE(cmd *cobra.Command, args []string) error {
	org := &orgv1.Organization{Id: c.State.Auth.Organization.Id}
	code := args[0]

	if err := c.Client.Billing.ClaimPromoCode(context.Background(), org, code); err != nil {
		return err
	}

	utils.Println(cmd, "Your promo code was successfully added.")
	return nil
}

func (c *promoCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List claimed promo codes.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.listRunE),
	}

	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().SortFlags = false

	return cmd
}

type humanRow struct {
	code       string
	balance    string
	expiration string
}

type structuredRow struct {
	code       string
	balance    float64
	expiration int64
}

func (c *promoCommand) listRunE(cmd *cobra.Command, _ []string) error {
	org := &orgv1.Organization{Id: c.State.Auth.User.OrganizationId}

	codes, err := c.Client.Billing.GetClaimedPromoCodes(context.Background(), org, true)
	if err != nil {
		return err
	}

	o, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}

	if len(codes) == 0 && o == "human" {
		utils.Println(cmd, "No promo codes found.")
		return nil
	}

	var (
		listFields       = []string{"code", "balance", "expiration"}
		humanLabels      = []string{"Code", "Balance", "Expiration"}
		structuredLabels = []string{"code", "balance", "expiration"}
	)

	w, err := output.NewListOutputWriter(cmd, listFields, humanLabels, structuredLabels)
	if err != nil {
		return err
	}

	for _, code := range codes {
		switch o {
		case "human":
			w.AddElement(&humanRow{
				code:       code.Code,
				balance:    formatBalance(code.Balance, code.Amount),
				expiration: formatExpiration(code.CreditExpirationDate.Seconds),
			})
		case "json", "yaml":
			w.AddElement(&structuredRow{
				code:       code.Code,
				balance:    convertToUSD(code.Balance),
				expiration: code.CreditExpirationDate.Seconds,
			})
		}
	}
	w.StableSort()

	return w.Out()
}

func formatBalance(balance int64, amount int64) string {
	return fmt.Sprintf("$%.2f/%.2f USD", convertToUSD(balance), convertToUSD(amount))
}

func convertToUSD(balance int64) float64 {
	// The backend represents money in hundredths of cents
	return float64(balance) / 10000
}

func formatExpiration(seconds int64) string {
	t := time.Unix(seconds, 0)
	return t.Format("Jan 2, 2006")
}
