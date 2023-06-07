package billing

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/billing"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type costOut struct {
	StartDate           string `human:"Start Date" serialized:"start_date"`
	EndDate             string `human:"End Date" serialized:"end_date"`
	Granularity         string `human:"Granularity" serialized:"granularity"`
	LineType            string `human:"Line Type" serialized:"line_type"`
	Product             string `human:"Product,omitempty" serialized:"product,omitempty"`
	ResourceId          string `human:"Resource ID,omitempty" serialized:"resource_id,omitempty"`
	ResourceDisplayName string `human:"Resource Display Name,omitempty" serialized:"resource_display_name,omitempty"`
	EnvironmentId       string `human:"Environment ID,omitempty" serialized:"environment_id,omitempty"`
	NetworkAccessType   string `human:"Network Access Type,omitempty" serialized:"network_access_type,omitempty"`
	Price               string `human:"Price,omitempty" serialized:"price,omitempty"`
	OriginalAmount      string `human:"Original Amount" serialized:"original_amount"`
	DiscountAmount      string `human:"Discount Amount,omitempty" serialized:"discount_amount,omitempty"`
	Amount              string `human:"Amount,omitempty" serialized:"amount,omitempty"`
}

func (c *command) newCostListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list --start-date <start-date> --end-date <end-date>",
		Short: "List Confluent Cloud billing costs.",
		Long:  "List Confluent Cloud daily aggregated costs for a specific range of dates.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List Billing costs from 2023-01-01 to 2023-01-10:`,
				Code: "confluent billing list --start-date 2023-01-01 --end-date 2023-01-10",
			}),
	}

	cmd.Flags().String("start-date", "", "Start Date.")
	cmd.Flags().String("end-date", "", "End Date.")
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("start-date"))
	cobra.CheckErr(cmd.MarkFlagRequired("end-date"))

	return cmd
}

func (c *command) checkDateFormat(date string) error {
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return fmt.Errorf("must be formatted as: YYYY-MM-DD")
	}

	return nil
}

func (c *command) list(cmd *cobra.Command, args []string) error {
	startDate, err := cmd.Flags().GetString("start-date")
	if err != nil {
		return err
	}

	endDate, err := cmd.Flags().GetString("end-date")
	if err != nil {
		return err
	}

	if err = c.checkDateFormat(startDate); err != nil {
		return fmt.Errorf("invalid start date: %v", err)
	}

	if err = c.checkDateFormat(startDate); err != nil {
		return fmt.Errorf("invalid end date: %v", err)
	}

	costs, err := c.V2Client.ListBillingCosts(startDate, endDate)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, cost := range costs {
		out := &costOut{
			Granularity:       cost.GetGranularity(),
			LineType:          cost.GetLineType(),
			Product:           cost.GetProduct(),
			NetworkAccessType: cost.GetNetworkAccessType(),
		}
		if cost.GetGranularity() == "DAILY" {
			out.StartDate = strings.TrimSuffix(cost.GetStartDate(), " 00:00:00")
			out.EndDate = strings.TrimSuffix(cost.GetEndDate(), " 00:00:00")
		}

		// These fields may be empty depending on the line type, so casting floats as strings as to avoid zero-value
		if price, ok := cost.GetPriceOk(); ok {
			out.Price = billing.FormatPrice(*price, cost.GetUnit())
		}

		if originalAmount, ok := cost.GetOriginalAmountOk(); ok {
			out.OriginalAmount = billing.FormatDollars(*originalAmount)
		}
		if discountAmount, ok := cost.GetDiscountAmountOk(); ok {
			out.DiscountAmount = billing.FormatDollars(*discountAmount)
		}
		if amount, ok := cost.GetAmountOk(); ok {
			out.Amount = billing.FormatDollars(*amount)
		}

		if resource, ok := cost.GetResourceOk(); ok {
			out.ResourceId = resource.GetId()
			out.ResourceDisplayName = resource.GetDisplayName()
			if environment, ok := resource.GetEnvironmentOk(); ok {
				out.EnvironmentId = environment.GetId()
			}
		}
		list.Add(out)
	}
	return list.Print()
}
