package billing

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/billing"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type promoHumanOut struct {
	Code       string `human:"Code"`
	Balance    string `human:"Balance"`
	Expiration string `human:"Expiration"`
}

type promoSerializedOut struct {
	Code       string  `serialized:"code"`
	Balance    float64 `serialized:"balance"`
	Expiration int64   `serialized:"expiration"`
}

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List claimed promo codes.",
		Args:  cobra.NoArgs,
		RunE:  c.promoList,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) promoList(cmd *cobra.Command, _ []string) error {
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
			list.Add(&promoHumanOut{
				Code:       code.GetCode(),
				Balance:    formatBalance(code.GetBalance(), code.GetAmount()),
				Expiration: formatExpiration(code.GetCreditExpirationDate().GetSeconds()),
			})
		} else {
			list.Add(&promoSerializedOut{
				Code:       code.GetCode(),
				Balance:    billing.ConvertToUSD(code.GetBalance()),
				Expiration: code.GetCreditExpirationDate().GetSeconds(),
			})
		}
	}
	return list.Print()
}

func formatBalance(balance, amount int64) string {
	return fmt.Sprintf("$%.2f/%.2f USD", billing.ConvertToUSD(balance), billing.ConvertToUSD(amount))
}

func formatExpiration(seconds int64) string {
	return time.Unix(seconds, 0).Format("Jan 2, 2006")
}
