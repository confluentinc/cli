package serdes

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/golang/protobuf/jsonpb" //nolint:staticcheck // deprecated module cannot be removed due to https://github.com/jhump/protoreflect/issues/301
	"github.com/golang/protobuf/proto"  //nolint:staticcheck // deprecated module cannot be removed due to https://github.com/jhump/protoreflect/issues/301
	parse "github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

type ProtobufSerializationProvider struct {
	message proto.Message
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
	return ProtobufSchemaBackendName
}

func (p *ProtobufSerializationProvider) Serialize(str string) ([]byte, error) {
	// Index array indicates which message in the file we're referring to.
	// In our case, index array is always [0].
	indexBytes := []byte{0x0}

	// Convert from JSON string to proto message type.
	if err := jsonpb.UnmarshalString(str, p.message); err != nil {
		return nil, fmt.Errorf(errors.ProtoDocumentInvalidErrorMsg)
	}

	// Serialize proto message type to binary format.
	data, err := proto.Marshal(p.message)
	if err != nil {
		return nil, err
	}
	data = append(indexBytes, data...)
	return data, nil
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
