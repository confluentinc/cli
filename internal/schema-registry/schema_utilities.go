package schemaregistry

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	"github.com/confluentinc/cli/v3/pkg/output"
	schemaregistry "github.com/confluentinc/cli/v3/pkg/schema-registry"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

type registerSchemaResponse struct {
	Id int32 `json:"id" yaml:"id"`
}

type RegisterSchemaConfigs struct {
	SchemaDir  string
	Subject    string
	Format     string
	SchemaType string
	SchemaPath string
	Refs       []srsdk.SchemaReference
	Metadata   srsdk.NullableMetadata
	Ruleset    srsdk.NullableRuleSet
	Normalize  bool
}

func RegisterSchemaWithAuth(cmd *cobra.Command, schemaCfg *RegisterSchemaConfigs, client *schemaregistry.Client) (int32, error) {
	schema, err := os.ReadFile(schemaCfg.SchemaPath)
	if err != nil {
		return 0, err
	}

	request := srsdk.RegisterSchemaRequest{
		Schema:     srsdk.PtrString(string(schema)),
		SchemaType: srsdk.PtrString(schemaCfg.SchemaType),
		References: &schemaCfg.Refs,
		Metadata:   schemaCfg.Metadata,
		RuleSet:    schemaCfg.Ruleset,
	}

	response, err := client.Register(schemaCfg.Subject, request, schemaCfg.Normalize)
	if err != nil {
		return 0, err
	}

	if output.GetFormat(cmd).IsSerialized() {
		if err := output.SerializedOutput(cmd, &registerSchemaResponse{Id: response.GetId()}); err != nil {
			return 0, err
		}
	} else {
		output.Printf(false, "Successfully registered schema with ID \"%d\".\n", response.GetId())
	}

	return response.GetId(), nil
}

func ReadSchemaReferences(cmd *cobra.Command, isKey bool) ([]srsdk.SchemaReference, error) {
	name := "references"
	if isKey {
		name = "key-references"
	}

	references, err := cmd.Flags().GetString(name)
	if err != nil {
		return nil, err
	}

	var refs []srsdk.SchemaReference
	if references != "" {
		refBlob, err := os.ReadFile(references)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(refBlob, &refs); err != nil {
			return nil, err
		}
	}

	return refs, nil
}

func StoreSchemaReferences(schemaDir string, refs []srsdk.SchemaReference, client *schemaregistry.Client) (map[string]string, error) {
	referencePathMap := map[string]string{}
	for _, ref := range refs {
		tempStorePath := filepath.Join(schemaDir, ref.GetName())
		if !utils.FileExists(tempStorePath) {
			schema, err := client.GetSchemaByVersion(ref.GetSubject(), strconv.Itoa(int(ref.GetVersion())), false)
			if err != nil {
				return nil, err
			}
			if err := os.MkdirAll(filepath.Dir(tempStorePath), 0755); err != nil {
				return nil, err
			}
			if err := os.WriteFile(tempStorePath, []byte(schema.GetSchema()), 0644); err != nil {
				return nil, err
			}
		}
		referencePathMap[ref.GetName()] = tempStorePath
	}
	return referencePathMap, nil
}

func GetMetaInfoFromSchemaId(id int32) []byte {
	metaInfo := []byte{0x0}
	schemaIdBuffer := make([]byte, 4)
	binary.BigEndian.PutUint32(schemaIdBuffer, uint32(id))
	return append(metaInfo, schemaIdBuffer...)
}

func createTempDir() (string, error) {
	dir := filepath.Join(os.TempDir(), "ccloud-schema")
	err := os.MkdirAll(dir, 0755)
	return dir, err
}

func SetSchemaPathRef(schemaString srsdk.SchemaString, dir, subject string, schemaId int32, client *schemaregistry.Client) (string, map[string]string, error) {
	// Create temporary file to store schema retrieved (also for cache). Retry if get error retrieving schema or writing temp schema file
	tempStorePath := filepath.Join(dir, fmt.Sprintf("%s-%d.txt", subject, schemaId))
	tempRefStorePath := filepath.Join(dir, fmt.Sprintf("%s-%d.ref", subject, schemaId))
	var references []srsdk.SchemaReference

	if !utils.FileExists(tempStorePath) || !utils.FileExists(tempRefStorePath) {
		// TODO: add handler for writing schema failure
		if err := os.WriteFile(tempStorePath, []byte(schemaString.GetSchema()), 0644); err != nil {
			return "", nil, err
		}

		refBytes, err := json.Marshal(schemaString.References)
		if err != nil {
			return "", nil, err
		}
		if err := os.WriteFile(tempRefStorePath, refBytes, 0644); err != nil {
			return "", nil, err
		}
		references = schemaString.GetReferences()
	} else {
		refBlob, err := os.ReadFile(tempRefStorePath)
		if err != nil {
			return "", nil, err
		}
		if err := json.Unmarshal(refBlob, &references); err != nil {
			return "", nil, err
		}
	}
	referencePathMap, err := StoreSchemaReferences(dir, references, client)
	if err != nil {
		return "", nil, err
	}
	return tempStorePath, referencePathMap, nil
}
