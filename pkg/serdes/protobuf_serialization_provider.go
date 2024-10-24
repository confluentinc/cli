package serdes

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bufbuild/protocompile"
	"google.golang.org/protobuf/encoding/protojson"
	gproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/protobuf"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

type ProtobufSerializationProvider struct {
	ser     *protobuf.Serializer
	message gproto.Message
}

func (p *ProtobufSerializationProvider) InitSerializer(srClientUrl, srClusterId, mode, srApiKey, srApiSecret, token string, schemaId int) error {
	var serdeClientConfig *schemaregistry.Config

	if srClientUrl == mockClientUrl {
		serdeClientConfig = schemaregistry.NewConfig(srClientUrl)
	} else if srApiKey != "" && srApiSecret != "" {
		serdeClientConfig = schemaregistry.NewConfigWithBasicAuthentication(srClientUrl, srApiKey, srApiSecret)
	} else if token != "" {
		serdeClientConfig = schemaregistry.NewConfigWithBearerAuthentication(srClientUrl, token, srClusterId, "")
	} else {
		return fmt.Errorf("schema registry client authentication should be provider to initialize serializer")
	}
	serdeClient, err := schemaregistry.NewClient(serdeClientConfig)

	if err != nil {
		return fmt.Errorf("failed to create serializer-specific Schema Registry client: %w", err)
	}

	// Configure the serde settings
	// If schemaId > 0 then use the intended schema ID
	// otherwise use the latest schema ID
	serdeConfig := protobuf.NewSerializerConfig()
	serdeConfig.AutoRegisterSchemas = false
	serdeConfig.UseLatestVersion = true
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

func (p *ProtobufSerializationProvider) Serialize(topic, message string) ([]byte, error) {
	// Need to materialize the message into the schema of p.message
	if err := protojson.Unmarshal([]byte(message), p.message); err != nil {
		return nil, fmt.Errorf(errors.ProtoDocumentInvalidErrorMsg)
	}

	payload, err := p.ser.Serialize(topic, p.message)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}
	return payload, nil
}

func parseMessage(schemaPath string, referencePathMap map[string]string) (gproto.Message, error) {
	// Collect import paths
	importPaths := []string{filepath.Dir(schemaPath)}
	for _, path := range referencePathMap {
		importPaths = append(importPaths, strings.SplitAfter(path, "ccloud-schema")[0])
	}

	resolver := &protocompile.SourceResolver{
		ImportPaths: importPaths,
	}

	// Create the compiler
	compiler := protocompile.Compiler{
		Resolver: resolver,
	}

	// Parse and compile the .proto files
	compiledFiles, err := compiler.Compile(context.Background(), filepath.Base(schemaPath))
	if err != nil {
		return nil, fmt.Errorf("error compiling or finding .proto files")
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
