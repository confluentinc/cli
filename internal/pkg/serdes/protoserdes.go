package serdes

import (
	"github.com/confluentinc/cli/internal/pkg/errors"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	parse "github.com/jhump/protoreflect/desc/protoparse"
	dynamic "github.com/jhump/protoreflect/dynamic"
)

type ProtoSerializationProvider struct {
	message proto.Message
}

func (protoProvider *ProtoSerializationProvider) LoadSchema(schemaPath string) error {
	parser := parse.Parser{}
	fileDescriptors, err := parser.ParseFiles(schemaPath)
	if err != nil {
		return errors.New(errors.ProtoSchemaInvalidErrorMsg)
	}
	if len(fileDescriptors) == 0 {
		return errors.New(errors.ProtoSchemaInvalidErrorMsg)
	}
	fileDescriptor := fileDescriptors[0]

	messageDescriptors := fileDescriptor.GetMessageTypes()
	if len(messageDescriptors) == 0 {
		return errors.New(errors.ProtoSchemaInvalidErrorMsg)
	}
	// We're always using the outermost first message.
	messageDescriptor := messageDescriptors[0]
	messageFactory := dynamic.NewMessageFactoryWithDefaults()
	message := messageFactory.NewMessage(messageDescriptor)
	protoProvider.message = message
	return nil
}

func (protoProvider *ProtoSerializationProvider) GetSchemaName() string {
	return PROTOBUFSCHEMANAME
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

func (protoProvider *ProtoDeserializationProvider) LoadSchema(schemaPath string) error {
	parser := parse.Parser{}
	fileDescriptors, err := parser.ParseFiles(schemaPath)
	if err != nil {
		return errors.New(errors.ProtoSchemaInvalidErrorMsg)
	}
	if len(fileDescriptors) == 0 {
		return errors.New(errors.ProtoSchemaInvalidErrorMsg)
	}
	fileDescriptor := fileDescriptors[0]

	messageDescriptors := fileDescriptor.GetMessageTypes()
	if len(messageDescriptors) == 0 {
		return errors.New(errors.ProtoSchemaInvalidErrorMsg)
	}
	// We're always using the outermost first message.
	messageDescriptor := messageDescriptors[0]
	messageFactory := dynamic.NewMessageFactoryWithDefaults()
	message := messageFactory.NewMessage(messageDescriptor)
	protoProvider.message = message
	return nil
}

func (protoProvider *ProtoDeserializationProvider) GetSchemaName() string {
	return PROTOBUFSCHEMANAME
}

func (protoProvider *ProtoDeserializationProvider) decode(data []byte) (string, error) {
	// Index array indicates which message in the file we're referring to.
	// In our case, we simply ignore the index array [0].
	data = data[1:]

	// Convert from binary format to proto message type.
	err := proto.Unmarshal(data, protoProvider.message)
	if err != nil {
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
