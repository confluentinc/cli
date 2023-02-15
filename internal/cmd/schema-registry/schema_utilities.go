package schemaregistry

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/antihax/optional"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type registerSchemaResponse struct {
	Id int32 `json:"id" yaml:"id"`
}

type RegisterSchemaConfigs struct {
	SchemaDir   string
	Subject     string
	ValueFormat string
	SchemaType  string
	SchemaPath  *string
	Refs        []srsdk.SchemaReference
}

func RegisterSchemaWithAuth(cmd *cobra.Command, schemaCfg *RegisterSchemaConfigs, srClient *srsdk.APIClient, ctx context.Context) ([]byte, error) {
	schema, err := os.ReadFile(*schemaCfg.SchemaPath)
	if err != nil {
		return nil, err
	}

	response, _, err := srClient.DefaultApi.Register(ctx, schemaCfg.Subject, srsdk.RegisterSchemaRequest{Schema: string(schema), SchemaType: schemaCfg.SchemaType, References: schemaCfg.Refs})
	if err != nil {
		return nil, err
	}

	if output.GetFormat(cmd).IsSerialized() {
		err = output.SerializedOutput(cmd, &registerSchemaResponse{Id: response.Id})
		if err != nil {
			return nil, err
		}
	} else {
		utils.Printf(cmd, errors.RegisteredSchemaMsg, response.Id)
	}

	return GetMetaInfoFromSchemaId(response.Id), nil
}

func ReadSchemaRefs(cmd *cobra.Command) ([]srsdk.SchemaReference, error) {
	var refs []srsdk.SchemaReference
	references, err := cmd.Flags().GetString("references")
	if err != nil {
		return nil, err
	}
	if references != "" {
		refBlob, err := os.ReadFile(references)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(refBlob, &refs)
		if err != nil {
			return nil, err
		}
	}
	return refs, nil
}

func StoreSchemaReferences(schemaDir string, refs []srsdk.SchemaReference, srClient *srsdk.APIClient, ctx context.Context) (map[string]string, error) {
	referencePathMap := map[string]string{}
	for _, ref := range refs {
		tempStorePath := filepath.Join(schemaDir, ref.Name)
		if !utils.FileExists(tempStorePath) {
			schema, _, err := srClient.DefaultApi.GetSchemaByVersion(ctx, ref.Subject, strconv.Itoa(int(ref.Version)), &srsdk.GetSchemaByVersionOpts{})
			if err != nil {
				return nil, err
			}
			err = os.MkdirAll(filepath.Dir(tempStorePath), 0755)
			if err != nil {
				return nil, err
			}
			err = os.WriteFile(tempStorePath, []byte(schema.Schema), 0644)
			if err != nil {
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

func RequestSchemaWithId(schemaId int32, subject string, srClient *srsdk.APIClient, ctx context.Context) (srsdk.SchemaString, error) {
	opts := &srsdk.GetSchemaOpts{Subject: optional.NewString(subject)}
	schemaString, _, err := srClient.DefaultApi.GetSchema(ctx, schemaId, opts)
	return schemaString, err
}

func SetSchemaPathRef(schemaString srsdk.SchemaString, dir string, subject string, schemaId int32, srClient *srsdk.APIClient, ctx context.Context) (string, map[string]string, error) {
	// Create temporary file to store schema retrieved (also for cache). Retry if get error retriving schema or writing temp schema file
	tempStorePath := filepath.Join(dir, fmt.Sprintf("%s-%d.txt", subject, schemaId))
	tempRefStorePath := filepath.Join(dir, fmt.Sprintf("%s-%d.ref", subject, schemaId))
	var references []srsdk.SchemaReference

	if !utils.FileExists(tempStorePath) || !utils.FileExists(tempRefStorePath) {
		// TODO: add handler for writing schema failure
		err := os.WriteFile(tempStorePath, []byte(schemaString.Schema), 0644)
		if err != nil {
			return "", nil, err
		}

		refBytes, err := json.Marshal(schemaString.References)
		if err != nil {
			return "", nil, err
		}
		err = os.WriteFile(tempRefStorePath, refBytes, 0644)
		if err != nil {
			return "", nil, err
		}
		references = schemaString.References
	} else {
		refBlob, err := os.ReadFile(tempRefStorePath)
		if err != nil {
			return "", nil, err
		}
		err = json.Unmarshal(refBlob, &references)
		if err != nil {
			return "", nil, err
		}
	}
	referencePathMap, err := StoreSchemaReferences(dir, references, srClient, ctx)
	if err != nil {
		return "", nil, err
	}
	return tempStorePath, referencePathMap, nil
}
