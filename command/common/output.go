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
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/codyaray/retag"
	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/olekukonko/tablewriter"

	"github.com/confluentinc/cc-structs/kafka/core/v1"
)

// ToRow formats a single row for inclusion in RenderTable output.
func ToRow(obj interface{}, fields []string) []string {
	c := reflect.ValueOf(obj).Elem()
	var data []string
	for _, field := range fields {
		data = append(data, fmt.Sprintf("%v", c.FieldByName(field)))
	}
	return data
}

// RenderTable outputs data in a tabular format.
func RenderTable(data [][]string, labels []string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(labels)
	table.AppendBulk(data)
	table.SetBorder(false)
	table.Render()
}

// RenderDetail outputs a subset of fields of an object, with fields renamed by labels.
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

type enumStringifyingFieldMaker struct{}

func (f *enumStringifyingFieldMaker) MakeField(oldType reflect.Type, newType reflect.Type, oldTag reflect.StructTag, newTag reflect.StructTag) (reflect.Type, reflect.StructTag) {
	if strings.Contains(string(oldTag), "protobuf:") && strings.Contains(string(oldTag), "enum=") {
		newType = reflect.TypeOf("")
	}
	return newType, newTag
}

// Render outputs an object detail in a specified format, optionally with a subset of (renamed) fields.
func Render(obj interface{}, fields []string, labels []string, outputFormat string) error {
	switch outputFormat {
	case "":
		fallthrough
	case "human":
		RenderDetail(obj, fields, labels)
	case "json":
		if msg, ok := obj.(proto.Message); ok {
			obj = prepareProtoStruct(msg, fields)
		}
		obj, err := reTagFields(obj, fields, labels, "json")
		if err != nil {
			return err
		}
		b, err := json.MarshalIndent(obj, "", "  ")
		if err != nil {
			return v1.WrapErr(err, "unable to marshal object to json for rendering")
		}
		fmt.Printf("%v\n", string(b))
	case "yaml":
		if msg, ok := obj.(proto.Message); ok {
			obj = prepareProtoStruct(msg, fields)
		}
		obj, err := reTagFields(obj, fields, labels, "json")
		b, err := yaml.Marshal(obj)
		if err != nil {
			return v1.WrapErr(err, "unable to marshal object to yaml for rendering")
		}
		fmt.Printf("%v\n", string(b))
	}
	return nil
}

// Helper which stringifies protobuf enum fields.
// Implemented by returning an anonymous dynamic struct with string field type in place of enum fields.
func prepareProtoStruct(msg proto.Message, fields []string) interface{} {
	m := jsonpb.Marshaler{EmitDefaults: true, EnumsAsInts: false}
	s, err := m.MarshalToString(msg)
	if err != nil {
		return err
	}
	tagMaker := &viewer{fields, fields, "json"}
	fieldMaker := &enumStringifyingFieldMaker{}
	obj := retag.ConvertFields(msg, tagMaker, fieldMaker)
	err = json.Unmarshal([]byte(s), &obj)
	if err != nil {
		return err
	}
	return obj
}

func reTagFields(obj interface{}, fields []string, labels []string, tagName string) (interface{}, error) {
	if fields == nil {
		return obj, nil
	}
	obj = retag.Convert(obj, &viewer{fields, labels, tagName})
	return obj, nil
}

type viewer struct {
	fields  []string
	labels  []string
	tagName string
}

func (s *viewer) MakeTag(t reflect.Type, fieldIndex int) reflect.StructTag {
	key := string(s.tagName)
	field := t.Field(fieldIndex)
	value := field.Tag.Get(key)
	value = s.updateForView(field.Name, fieldIndex)
	tag := fmt.Sprintf(`%s:"%s"`, key, value)
	return reflect.StructTag(tag)
}

func (s *viewer) updateForView(src string, fieldIndex int) string {
	if i, ok := contains(s.fields, src); ok {
		return s.labels[i]
	} else {
		return "-"
	}
}

func contains(s []string, e string) (int, bool) {
	for i, a := range s {
		if a == e {
			return i, true
		}
	}
	return -1, false
}
