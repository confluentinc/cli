package common

/* Output Usage

List View:

  Code:

	var data [][]string
	for _, cluster := range clusters {
		data = append(data, common.ToRow(cluster, []string{"Name", "ServiceProvider", "Region", "Durability", "Status"}))
	}
	common.RenderTable(data, []string{"Name", "Provider", "Region", "Durability", "Status"})

  Output:

	      NAME      | PROVIDER |  REGION   | DURABILITY | STATUS
	+---------------+----------+-----------+------------+---------+
	  sr-test       | aws      | us-east-1 | LOW        | UP
	  asdf          | aws      | us-east-1 | LOW        | UP
	  thisdaone     | aws      | us-east-1 | LOW        | UP

Describe View:

  Code:

	fields := []string{"Name", "NetworkIngress", "NetworkEgress", "Storage", "ServiceProvider", "Region", "Status", "Endpoint", "PricePerHour"}
	labels := []string{"Name", "Ingress", "Egress", "Storage", "Provider", "Region", "Status", "Endpoint", "PricePerHour"}
	common.RenderDetail(cluster, fields, labels)

  Output:

	          Name : sr-test
	       Ingress : 1
	        Egress : 1
	       Storage : 500
	      Provider : aws
	        Region : us-east-1
	        Status : UP
	      Endpoint : SASL_SSL://r0.kafka-mt-1.us-east-1.aws.stag.cpdev.cloud:9092,r0.kafka-mt-1.us-east-1.aws.stag.cpdev.cloud:9093,r0.kafka-mt-1.us-east-1.aws.stag.cpdev.cloud:9094
	  PricePerHour : 6849
*/

import (
	"fmt"
	"os"
	"reflect"

	"github.com/olekukonko/tablewriter"
)

func ToRow(obj interface{}, fields []string) []string {
	c := reflect.ValueOf(obj).Elem()
	var data []string
	for _, field := range fields {
		data = append(data, fmt.Sprintf("%v", c.FieldByName(field)))
	}
	return data
}

func RenderTable(data [][]string, labels []string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(labels)
	table.AppendBulk(data)
	table.SetBorder(false)
	table.Render()
}

func RenderDetail(obj interface{}, fields []string, labels []string) {
	c := reflect.ValueOf(obj).Elem()
	var data [][]string
	if fields == nil {
		for i := 0; i < c.NumField(); i++ {
			field := c.Field(i)
			fieldType := c.Type().Field(i)
			data = append(data, []string{fieldType.Name, fmt.Sprintf("%v", field)})
		}
	} else {
		for i, field := range fields {
			data = append(data, []string{labels[i], fmt.Sprintf("%v", c.FieldByName(field))})
		}
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.AppendBulk(data)
	table.SetColumnSeparator(":")
	table.SetColumnAlignment([]int{tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_LEFT})
	table.SetBorder(false)
	table.Render()
}
