package schemaregistry

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	schemaregistry "github.com/confluentinc/cli/internal/pkg/schema-registry"
	"github.com/confluentinc/cli/internal/pkg/utils"
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
	Metadata   *srsdk.Metadata
	Ruleset    *srsdk.RuleSet
	Normalize  bool
}

func RegisterSchemaWithAuth(cmd *cobra.Command, schemaCfg *RegisterSchemaConfigs, client *schemaregistry.Client) (int32, error) {
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

	if output.GetFormat(cmd).IsSerialized() {
		if err := output.SerializedOutput(cmd, &registerSchemaResponse{Id: response.Id}); err != nil {
			return 0, err
		}
	} else {
		output.Printf(errors.RegisteredSchemaMsg, response.Id)
	}

	return response.Id, nil
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

func GetMetaInfoFromSchemaId(id int32) []byte {
	metaInfo := []byte{0x0}
	schemaIdBuffer := make([]byte, 4)
	binary.BigEndian.PutUint32(schemaIdBuffer, uint32(id))
	return append(metaInfo, schemaIdBuffer...)
}

func CreateTempDir() (string, error) {
	dir := filepath.Join(os.TempDir(), "ccloud-schema")
	err := os.MkdirAll(dir, 0755)
	return dir, err
}

func RequestSchemaWithId(id int32, subject string, client *schemaregistry.Client) (srsdk.SchemaString, error) {
	opts := &srsdk.GetSchemaOpts{Subject: optional.NewString(subject)}
	return client.GetSchema(id, opts)
}

func SetSchemaPathRef(schemaString srsdk.SchemaString, dir, subject string, schemaId int32, client *schemaregistry.Client) (string, map[string]string, error) {
	// Create temporary file to store schema retrieved (also for cache). Retry if get error retrieving schema or writing temp schema file
	tempStorePath := filepath.Join(dir, fmt.Sprintf("%s-%d.txt", subject, schemaId))
	tempRefStorePath := filepath.Join(dir, fmt.Sprintf("%s-%d.ref", subject, schemaId))
	var references []srsdk.SchemaReference

	if !utils.FileExists(tempStorePath) || !utils.FileExists(tempRefStorePath) {
		// TODO: add handler for writing schema failure
		if err := os.WriteFile(tempStorePath, []byte(schemaString.Schema), 0644); err != nil {
			return "", nil, err
		}

		refBytes, err := json.Marshal(schemaString.References)
		if err != nil {
			return "", nil, err
		}
		if err := os.WriteFile(tempRefStorePath, refBytes, 0644); err != nil {
			return "", nil, err
		}
		references = schemaString.References
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
