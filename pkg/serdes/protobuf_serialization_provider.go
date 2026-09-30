package serdes

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/bufbuild/protocompile"
	"google.golang.org/protobuf/encoding/protojson"
	gproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/rules/cel"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/rules/encryption"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/rules/encryption/awskms"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/rules/encryption/azurekms"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/rules/encryption/gcpkms"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/rules/encryption/hcvault"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/rules/encryption/localkms"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/rules/jsonata"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/protobuf"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

// Embed all .proto built-in schema files in both folders
//
//go:embed google/protobuf/*.proto
//go:embed google/type/*.proto
//go:embed confluent/*.proto
//go:embed confluent/type/*.proto
var builtInSchemas embed.FS

type ProtobufSerializationProvider struct {
	ser     *protobuf.Serializer
	message gproto.Message
}

func (p *ProtobufSerializationProvider) InitSerializer(srClientUrl, srClusterId, mode string, schemaId int, srAuth SchemaRegistryAuth) error {
	serdeClient, err := initSchemaRegistryClient(srClientUrl, srClusterId, srAuth, nil)
	if err != nil {
		return fmt.Errorf("failed to create serializer-specific Schema Registry client: %w", err)
	}

	// Register the KMS drivers and the field-level encryption executor
	awskms.Register()
	azurekms.Register()
	gcpkms.Register()
	hcvault.Register()
	localkms.Register()
	encryption.Register()
	cel.Register()
	jsonata.Register()

	if err != nil {
		return fmt.Errorf("failed to create serializer-specific Schema Registry client: %w", err)
	}

	// Configure the serde settings
	// If schemaId > 0 then use the intended schema ID
	// otherwise use the latest schema ID
	serdeConfig := protobuf.NewSerializerConfig()
	serdeConfig.AutoRegisterSchemas = false
	serdeConfig.UseLatestVersion = true

	// local KMS secret is only set and used during local testing with ruleSet
	if localKmsSecretValue := os.Getenv(localKmsSecretMacro); srClientUrl == mockClientUrl && localKmsSecretValue != "" {
		serdeConfig.RuleConfig = map[string]string{
			localKmsSecretKey: localKmsSecretValue,
		}
	}

	if schemaId > 0 {
		serdeConfig.UseSchemaID = schemaId
		serdeConfig.UseLatestVersion = false
	}

	var serdeType serde.Type
	if mode == "key" {
		serdeType = serde.KeySerde
	} else if mode == "value" {
		serdeType = serde.ValueSerde
	} else {
		return fmt.Errorf("unknown serialization mode: %s", mode)
	}

	ser, err := protobuf.NewSerializer(serdeClient, serdeType, serdeConfig)

	if err != nil {
		return fmt.Errorf("failed to initialize PROTOBUF serializer: %w", err)
	}

	p.ser = ser
	return nil
}

func (p *ProtobufSerializationProvider) LoadSchema(schemaPath string, referencePathMap map[string]string) error {
	message, err := parseMessage(schemaPath, referencePathMap)
	if err != nil {
		return err
	}
	p.message = message
	return nil
}

func (p *ProtobufSerializationProvider) GetSchemaName() string {
	return protobufSchemaBackendName
}

func (p *ProtobufSerializationProvider) Serialize(topic, message string) ([]kafka.Header, []byte, error) {
	// Need to materialize the message into the schema of p.message
	if err := protojson.Unmarshal([]byte(message), p.message); err != nil {
		return nil, nil, fmt.Errorf(errors.ProtoDocumentInvalidErrorMsg)
	}

	headers, payload, err := p.ser.SerializeWithHeaders(topic, p.message)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to serialize message: %w", err)
	}
	return headers, payload, nil
}

func parseMessage(schemaPath string, referencePathMap map[string]string) (gproto.Message, error) {
	if schemaPath == "" {
		return nil, fmt.Errorf("schema path is empty")
	}

	// Collect import paths
	importPath := filepath.Dir(schemaPath)
	importPaths := []string{importPath}

	for _, path := range referencePathMap {
		importPaths = append(importPaths, strings.SplitAfter(path, "ccloud-schema")[0])
	}

	resolver := &protocompile.SourceResolver{
		ImportPaths: importPaths,
	}

	// Extract and copy embedded builtin proto files schemas needed for CSFLE to a temp destination directory
	if err := copyBuiltInProtoFiles(importPaths[0]); err != nil {
		return nil, fmt.Errorf("failed to copy built-in proto files to the temp folder: %w", err)
	}

	// Create the compiler
	compiler := protocompile.Compiler{
		Resolver: resolver,
	}

	// Parse and compile the .proto files
	compiledFiles, err := compiler.Compile(context.Background(), filepath.Base(schemaPath))
	if err != nil {
		return nil, fmt.Errorf("error compiling .proto files: %w\n", err)
	}
	if len(compiledFiles) == 0 {
		return nil, fmt.Errorf("error fetching valid compiled files")
	}

	// Get the first compiled file descriptor
	fileDescriptor := compiledFiles[0]

	// Get the message descriptors
	messageDescriptors := fileDescriptor.Messages()
	if messageDescriptors.Len() == 0 {
		return nil, fmt.Errorf("proto schema invalid: no message descriptors found")
	}

	// Always use the outermost first message
	messageDescriptor := messageDescriptors.Get(0)

	// Create a dynamic message from the descriptor
	dynamicMessage := dynamicpb.NewMessage(messageDescriptor)
	return dynamicMessage, nil
}

// GetSchemaRegistryClient This getter function is used in mock testing
// as serializer and deserializer have to share the same SR client instance
func (p *ProtobufSerializationProvider) GetSchemaRegistryClient() schemaregistry.Client {
	return p.ser.Client
}

func copyBuiltInProtoFiles(destinationDir string) error {
	return fs.WalkDir(builtInSchemas, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Read file content from the embedded filesystem
		content, err := builtInSchemas.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		// Determine the destination path
		destPath := filepath.Join(destinationDir, path)

		// Ensure the destination directory exists
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", destPath, err)
		}

		// Write the built-in schema files to the destination
		if err := os.WriteFile(destPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", destPath, err)
		}

		return nil
	})
}

func (p *ProtobufSerializationProvider) SetSchemaIDSerializer(headerSerializer serde.SchemaIDSerializerFunc) {
	p.ser.SchemaIDSerializer = headerSerializer
}
