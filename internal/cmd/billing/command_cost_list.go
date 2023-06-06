package billing

import (
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/billing"
	"time"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type costOut struct {
	StartDate           string  `human:"Start Date" serialized:"start_date"`
	EndDate             string  `human:"End Date" serialized:"end_date"`
	Granularity         string  `human:"Granularity" serialized:"granularity"`
	LineType            string  `human:"Line Type" serialized:"line_type"`
	Product             string  `human:"Product" serialized:"product,omitempty"`
	ResourceId          string  `human:"Resource ID" serialized:"resource_id,omitempty"`
	ResourceDisplayName string  `human:"Resource Display Name" serialized:"resource_display_name,omitempty"`
	EnvironmentId       string  `human:"Environment ID" serialized:"environment_id,omitempty"`
	NetworkAccessType   string  `human:"Network Access Type" serialized:"network_access_type,omitempty"`
	Price               string  `human:"Price" serialized:"price,omitempty"`
	OriginalAmount      float64 `human:"Original Amount" serialized:"original_amount"`
	DiscountAmount      float64 `human:"Discount Amount" serialized:"discount_amount,omitempty"`
	Amount              float64 `human:"Amount" serialized:"amount,omitempty"`
}

func (c *command) newCostListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list <start-date> <end-date>",
		Example: "list 2023-01-01 2023-01-10",
		Short:   "List Confluent Cloud billing costs.",
		Long:    "List Confluent Cloud daily aggregated costs for a specific range of dates.",
		Args:    cobra.ExactArgs(2),
		RunE:    c.list,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) checkDateFormat(date string) error {
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return fmt.Errorf("expected format should look like: 2022-01-01")
	}

	return nil
}

func (c *command) list(cmd *cobra.Command, args []string) error {
	startDate := args[0]
	endDate := args[1]

	err := c.checkDateFormat(startDate)
	if err != nil {
		return fmt.Errorf("invalid start date: %s", err.Error())
	}

	err = c.checkDateFormat(endDate)
	if err != nil {
		return fmt.Errorf("invalid end date: %s", err.Error())
	}

	costs, err := c.V2Client.ListBillingCosts(startDate, endDate)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, cost := range costs {
		out := &costOut{
			StartDate:         cost.GetStartDate(),
			EndDate:           cost.GetEndDate(),
			Granularity:       cost.GetGranularity(),
			LineType:          cost.GetLineType(),
			Product:           cost.GetProduct(),
			NetworkAccessType: cost.GetNetworkAccessType(),
		}

		// These fields may be empty depending on the line type, so casting floats as strings as to avoid zero-value
		if price, ok := cost.GetPriceOk(); ok {
			out.Price = billing.FormatPrice(*price, cost.GetUnit())
		}

		if originalAmount, ok := cost.GetOriginalAmountOk(); ok {
			out.OriginalAmount = *originalAmount
		}

		if discountAmount, ok := cost.GetDiscountAmountOk(); ok {
			out.DiscountAmount = *discountAmount
		}
		if amount, ok := cost.GetAmountOk(); ok {
			out.Amount = *amount
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
