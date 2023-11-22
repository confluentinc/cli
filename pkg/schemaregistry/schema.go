package schemaregistry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"

	"github.com/antihax/optional"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	"github.com/confluentinc/cli/v3/pkg/utils"
)

type RegisterSchemaConfigs struct {
	SchemaDir  string
	Subject    string
	Format     string
	SchemaType string
	SchemaPath string
	Refs       []srsdk.SchemaReference
	Metadata   *srsdk.Metadata
	Ruleset    *srsdk.RuleSet
	Normalize  bool
}

type RegisterSchemaResponse struct {
	Id int32 `json:"id" yaml:"id"`
}

func ReadSchemaReferences(references string) ([]srsdk.SchemaReference, error) {
	if references == "" {
		return []srsdk.SchemaReference{}, nil
	}

	data, err := os.ReadFile(references)
	if err != nil {
		return nil, err
	}

	var refs []srsdk.SchemaReference
	if err := json.Unmarshal(data, &refs); err != nil {
		return nil, err
	}

	return refs, nil
}

func StoreSchemaReferences(schemaDir string, refs []srsdk.SchemaReference, client *Client) (map[string]string, error) {
	referencePathMap := map[string]string{}
	for _, ref := range refs {
		tempStorePath := filepath.Join(schemaDir, ref.Name)
		if !utils.FileExists(tempStorePath) {
			schema, err := client.GetSchemaByVersion(ref.Subject, strconv.Itoa(int(ref.Version)), nil)
			if err != nil {
				return nil, err
			}
			if err := os.MkdirAll(filepath.Dir(tempStorePath), 0755); err != nil {
				return nil, err
			}
			if err := os.WriteFile(tempStorePath, []byte(schema.Schema), 0644); err != nil {
				return nil, err
			}
		}
		referencePathMap[ref.Name] = tempStorePath
	}
	return referencePathMap, nil
}

func RegisterSchemaWithAuth(schemaCfg *RegisterSchemaConfigs, client *Client) (int32, error) {
	schema, err := os.ReadFile(schemaCfg.SchemaPath)
	if err != nil {
		return 0, err
	}

	request := srsdk.RegisterSchemaRequest{
		Schema:     string(schema),
		SchemaType: schemaCfg.SchemaType,
		References: schemaCfg.Refs,
		Metadata:   schemaCfg.Metadata,
		RuleSet:    schemaCfg.Ruleset,
	}

	opts := &srsdk.RegisterOpts{}
	if schemaCfg.Normalize {
		opts.Normalize = optional.NewBool(true)
	}

	response, err := client.Register(schemaCfg.Subject, request, opts)
	if err != nil {
		return 0, err
	}

	return response.Id, nil
}
