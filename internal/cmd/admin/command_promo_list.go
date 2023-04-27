package admin

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type humanOut struct {
	Code       string `human:"Code"`
	Balance    string `human:"Balance"`
	Expiration string `human:"Expiration"`
}

type serializedOut struct {
	Code       string  `serialized:"code"`
	Balance    float64 `serialized:"balance"`
	Expiration int64   `serialized:"expiration"`
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
	user, err := c.Client.Auth.User()
	if err != nil {
		return err
	}

	codes, err := c.Client.Billing.GetClaimedPromoCodes(user.GetOrganization(), true)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, code := range codes {
		if output.GetFormat(cmd) == output.Human {
			list.Add(&humanOut{
				Code:       code.GetCode(),
				Balance:    formatBalance(code.GetBalance(), code.GetAmount()),
				Expiration: formatExpiration(code.GetCreditExpirationDate().GetSeconds()),
			})
		} else {
			list.Add(&serializedOut{
				Code:       code.GetCode(),
				Balance:    ConvertToUSD(code.GetBalance()),
				Expiration: code.GetCreditExpirationDate().GetSeconds(),
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
