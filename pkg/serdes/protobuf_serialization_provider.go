package serdes

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/golang/protobuf/proto"
	parse "github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
	gproto "google.golang.org/protobuf/proto"

	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/protobuf"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/golang/protobuf/jsonpb"
)

type ProtobufSerializationProvider struct {
	ser      *protobuf.Serializer
	message  proto.Message
	message2 gproto.Message
}

func (p *ProtobufSerializationProvider) InitSerializer(srClientUrl, mode string, schemaId int) error {
	serdeClientConfig := schemaregistry.NewConfig(srClientUrl)
	serdeClient, err := schemaregistry.NewClient(serdeClientConfig)

	if err != nil {
		return fmt.Errorf("failed to create serializer-specific Schema Registry client: %w", err)
	}

	// Configure the serde settings
	// If schemaId > 0 then use the intended schema ID
	// otherwise use the latest schema ID
	// Configuring this correctly determines the underlying serialize strategy
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
		return fmt.Errorf("failed to create serializer: %w", err)
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
	// Convert the plain string message from customer type-in into proto.Message
	// TODO: replace the message if possible
	if err := jsonpb.UnmarshalString(message, p.message); err != nil {
		return nil, fmt.Errorf(errors.ProtoDocumentInvalidErrorMsg)
	}

	payload, err := p.ser.Serialize(topic, proto.MessageV2(p.message))
	if err != nil {
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}
	return payload, nil
}

func parseMessage(schemaPath string, referencePathMap map[string]string) (proto.Message, error) {
	importPaths := []string{filepath.Dir(schemaPath)}
	for _, path := range referencePathMap {
		importPaths = append(importPaths, strings.SplitAfter(path, "ccloud-schema")[0])
	}
	parser := parse.Parser{ImportPaths: importPaths}
	fileDescriptors, err := parser.ParseFiles(filepath.Base(schemaPath))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errors.ProtoSchemaInvalidErrorMsg, err)
	}
	if len(fileDescriptors) == 0 {
		return nil, fmt.Errorf(errors.ProtoSchemaInvalidErrorMsg)
	}
	fileDescriptor := fileDescriptors[0]

	messageDescriptors := fileDescriptor.GetMessageTypes()
	if len(messageDescriptors) == 0 {
		return nil, fmt.Errorf(errors.ProtoSchemaInvalidErrorMsg)
	}
	// We're always using the outermost first message.
	messageDescriptor := messageDescriptors[0]
	messageFactory := dynamic.NewMessageFactoryWithDefaults()
	return messageFactory.NewMessage(messageDescriptor), nil
}

// GetSchemaRegistryClient This getter function is used in mock testing
// as serializer and deserializer have to share the same SR client instance
func (p *ProtobufSerializationProvider) GetSchemaRegistryClient() schemaregistry.Client {
	return p.ser.Client
}
