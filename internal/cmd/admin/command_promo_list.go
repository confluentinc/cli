package admin

import (
	"context"
	"fmt"
	"time"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type out struct {
	Code       string `human:"Code" serialized:"code"`
	Balance    string `human:"Balance" serialized:"balance"`
	Expiration string `human:"Expiration" serialized:"expiration"`
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
	org := &orgv1.Organization{Id: c.Context.GetOrganization().GetId()}

	codes, err := c.Client.Billing.GetClaimedPromoCodes(context.Background(), org, true)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, code := range codes {
		if output.GetFormat(cmd).IsSerialized() {
			list.Add(&out{
				Code:       code.GetCode(),
				Balance:    fmt.Sprint(ConvertToUSD(code.GetBalance())),
				Expiration: fmt.Sprint(code.GetCreditExpirationDate().GetSeconds()),
			})
		} else {
			list.Add(&out{
				Code:       code.GetCode(),
				Balance:    formatBalance(code.GetBalance(), code.GetAmount()),
				Expiration: formatExpiration(code.GetCreditExpirationDate().GetSeconds()),
			})
		}
	}
	return list.Print()
}

func formatBalance(balance, amount int64) string {
	return fmt.Sprintf("$%.2f/%.2f USD", ConvertToUSD(balance), ConvertToUSD(amount))
}

func ConvertToUSD(balance int64) float64 {
	// The backend represents money in hundredths of cents
	return float64(balance) / 10000
}

func formatExpiration(seconds int64) string {
	return time.Unix(seconds, 0).Format("Jan 2, 2006")
}
