package admin

import (
	"context"
	"fmt"
	"time"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

var (
	listFields       = []string{"code", "balance", "expiration"}
	humanLabels      = []string{"Code", "Balance", "Expiration"}
	structuredLabels = []string{"code", "balance", "expiration"}
)

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

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List claimed promo codes.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	org := &orgv1.Organization{Id: c.State.Auth.Account.OrganizationId}

	codes, err := c.Client.Billing.GetClaimedPromoCodes(context.Background(), org, true)
	if err != nil {
		return err
	}

	o, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}

	if len(codes) == 0 && o == output.Human.String() {
		utils.Println(cmd, "No promo codes found.")
		return nil
	}

	w, err := output.NewListOutputWriter(cmd, listFields, humanLabels, structuredLabels)
	if err != nil {
		return err
	}

	for _, code := range codes {
		switch o {
		case output.Human.String():
			w.AddElement(&humanRow{
				code:       code.Code,
				balance:    formatBalance(code.Balance, code.Amount),
				expiration: formatExpiration(code.CreditExpirationDate.Seconds),
			})
		case output.JSON.String(), output.YAML.String():
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
