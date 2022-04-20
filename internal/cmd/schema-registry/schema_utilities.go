package schemaregistry

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"
)

type registerSchemaResponse struct {
	Id int32 `json:"id" yaml:"id"`
}

func RegisterSchemaWithAuth(cmd *cobra.Command, subject, schemaType, schemaPath string, refs []srsdk.SchemaReference, srClient *srsdk.APIClient, ctx context.Context) ([]byte, error) {
	schema, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return nil, err
	}

	response, _, err := srClient.DefaultApi.Register(ctx, subject, srsdk.RegisterSchemaRequest{Schema: string(schema), SchemaType: schemaType, References: refs})
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

	metaInfo := getMetaInfoFromSchemaId(response.Id)
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

func getMetaInfoFromSchemaId(id int32) []byte {
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
