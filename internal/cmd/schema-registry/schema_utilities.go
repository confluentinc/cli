package schemaregistry

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	schema, err := ioutil.ReadFile(*schemaCfg.SchemaPath)
	if err != nil {
		return nil, err
	}

	response, _, err := srClient.DefaultApi.Register(ctx, schemaCfg.Subject, srsdk.RegisterSchemaRequest{Schema: string(schema), SchemaType: schemaCfg.SchemaType, References: schemaCfg.Refs})
	if err != nil {
		return nil, err
	}

	outputFormat, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return nil, err
	}
	if outputFormat == output.Human.String() {
		utils.Printf(cmd, errors.RegisteredSchemaMsg, response.Id)
	} else {
		registerSchemaResponse := &registerSchemaResponse{Id: response.Id}
		err = output.StructuredOutput(outputFormat, registerSchemaResponse)
		if err != nil {
			return nil, err
		}
	}

	metaInfo := GetMetaInfoFromSchemaId(response.Id)
	return metaInfo, nil
}

func ReadSchemaRefs(cmd *cobra.Command) ([]srsdk.SchemaReference, error) {
	var refs []srsdk.SchemaReference
	refPath, err := cmd.Flags().GetString("refs")
	if err != nil {
		return nil, err
	}
	if refPath != "" {
		refBlob, err := ioutil.ReadFile(refPath)
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
			err = ioutil.WriteFile(tempStorePath, []byte(schema.Schema), 0644)
			if err != nil {
				return nil, err
			}
		}
		referencePathMap[ref.Name] = tempStorePath
	}
	return referencePathMap, nil
}

func RequestSchemaWithId(schemaId int32, schemaPath string, subject string, srClient *srsdk.APIClient, ctx context.Context) (string, map[string]string, error) {
	// Create temporary file to store schema retrieved (also for cache). Retry if get error retriving schema or writing temp schema file
	tempStorePath := filepath.Join(schemaPath, fmt.Sprintf("%s-%d.txt", subject, schemaId))
	tempRefStorePath := filepath.Join(schemaPath, fmt.Sprintf("%s-%d.ref", subject, schemaId))
	var references []srsdk.SchemaReference
	if !utils.FileExists(tempStorePath) || !utils.FileExists(tempRefStorePath) {
		// TODO: add handler for writing schema failure
		getSchemaOpts := srsdk.GetSchemaOpts{
			Subject: optional.NewString(subject),
		}
		schemaString, _, err := srClient.DefaultApi.GetSchema(ctx, schemaId, &getSchemaOpts)
		if err != nil {
			return "", nil, err
		}
		err = ioutil.WriteFile(tempStorePath, []byte(schemaString.Schema), 0644)
		if err != nil {
			return "", nil, err
		}

		refBytes, err := json.Marshal(schemaString.References)
		if err != nil {
			return "", nil, err
		}
		err = ioutil.WriteFile(tempRefStorePath, refBytes, 0644)
		if err != nil {
			return "", nil, err
		}
		references = schemaString.References
	} else {
		refBlob, err := ioutil.ReadFile(tempRefStorePath)
		if err != nil {
			return "", nil, err
		}
		err = json.Unmarshal(refBlob, &references)
		if err != nil {
			return "", nil, err
		}
	}

	// Store the references in temporary files
	referencePathMap, err := StoreSchemaReferences(schemaPath, references, srClient, ctx)
	if err != nil {
		return "", nil, err
	}

	return tempStorePath, referencePathMap, nil
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
