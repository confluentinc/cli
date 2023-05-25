package serdes

import (
	"path/filepath"
	"strings"

	"github.com/golang/protobuf/jsonpb" //nolint:staticcheck // deprecated module cannot be removed due to https://github.com/jhump/protoreflect/issues/301
	"github.com/golang/protobuf/proto"  //nolint:staticcheck // deprecated module cannot be removed due to https://github.com/jhump/protoreflect/issues/301
	parse "github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

type ProtoSerializationProvider struct {
	message proto.Message
}

func (protoProvider *ProtoSerializationProvider) LoadSchema(schemaPath string, referencePathMap map[string]string) error {
	message, err := parseMessage(schemaPath, referencePathMap)
	if err != nil {
		return err
	}
	protoProvider.message = message
	return nil
}

func (protoProvider *ProtoSerializationProvider) GetSchemaName() string {
	return PROTOBUFSCHEMABACKEND
}

func (protoProvider *ProtoSerializationProvider) encode(str string) ([]byte, error) {
	// Index array indicates which message in the file we're referring to.
	// In our case, index array is always [0].
	indexBytes := []byte{0x0}

	// Convert from Json string to proto message type.
	if err := jsonpb.UnmarshalString(str, protoProvider.message); err != nil {
		return nil, errors.New(errors.ProtoDocumentInvalidErrorMsg)
	}

	// Serialize proto message type to binary format.
	data, err := proto.Marshal(protoProvider.message)
	if err != nil {
		return nil, err
	}
	data = append(indexBytes, data...)
	return data, nil
}

type ProtoDeserializationProvider struct {
	message proto.Message
}

func (protoProvider *ProtoDeserializationProvider) LoadSchema(schemaPath string, referencePathMap map[string]string) error {
	message, err := parseMessage(schemaPath, referencePathMap)
	if err != nil {
		return err
	}
	protoProvider.message = message
	return nil
}

func (protoProvider *ProtoDeserializationProvider) GetSchemaName() string {
	return PROTOBUFSCHEMABACKEND
}

func (protoProvider *ProtoDeserializationProvider) decode(data []byte) (string, error) {
	// Index array indicates which message in the file we're referring to.
	// In our case, we simply ignore the index array [0].
	data = data[1:]

	// Convert from binary format to proto message type.
	if err := proto.Unmarshal(data, protoProvider.message); err != nil {
		return "", errors.New(errors.ProtoDocumentInvalidErrorMsg)
	}

	// Convert from proto message type to Json string.
	marshaler := &jsonpb.Marshaler{}
	str, err := marshaler.MarshalToString(protoProvider.message)
	if err != nil {
		return "", err
	}

	return str, nil
}

func parseMessage(schemaPath string, referencePathMap map[string]string) (proto.Message, error) {
	importPaths := []string{filepath.Dir(schemaPath)}
	for _, path := range referencePathMap {
		importPaths = append(importPaths, strings.SplitAfter(path, "ccloud-schema")[0])
	}
	parser := parse.Parser{ImportPaths: importPaths}
	fileDescriptors, err := parser.ParseFiles(filepath.Base(schemaPath))
	if err != nil {
		return nil, errors.Wrap(err, errors.ProtoSchemaInvalidErrorMsg)
	}
	if len(fileDescriptors) == 0 {
		return nil, errors.New(errors.ProtoSchemaInvalidErrorMsg)
	}
	fileDescriptor := fileDescriptors[0]

	messageDescriptors := fileDescriptor.GetMessageTypes()
	if len(messageDescriptors) == 0 {
		return nil, errors.New(errors.ProtoSchemaInvalidErrorMsg)
	}
	// We're always using the outermost first message.
	messageDescriptor := messageDescriptors[0]
	messageFactory := dynamic.NewMessageFactoryWithDefaults()
	return messageFactory.NewMessage(messageDescriptor), nil
}
