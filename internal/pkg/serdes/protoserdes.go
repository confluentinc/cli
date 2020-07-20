package serdes

import (
	"errors"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	//"google.golang.org/protobuf/reflect/protoreflect"
	parse "github.com/jhump/protoreflect/desc/protoparse"
	dynamic "github.com/jhump/protoreflect/dynamic"
)

type ProtoSerializationProvider uint32

func (protoProvider *ProtoSerializationProvider) GetSchemaName() string {
	return "PROTOBUF"
}

func (protoProvider *ProtoSerializationProvider) encode(str string, schemaPath string) ([]byte, error) {
	// Index array indicates which message in the file we're referring to.
	// In our case, index array is always [0].
	indexBytes := []byte{0x0}

	parser := parse.Parser{}
	fileDescriptors, err := parser.ParseFiles(schemaPath)
	if err != nil {
		return nil, err
	}
	if len(fileDescriptors) == 0 {
		return nil, errors.New("No valid file entered.")
	}
	fileDescriptor := fileDescriptors[0]

	messageDescriptors := fileDescriptor.GetMessageTypes()
	if len(messageDescriptors) == 0 {
		return nil, errors.New("No valid message in the file.")
	}
	// We're always using the outermost first message.
	messageDescriptor := messageDescriptors[0]
	messageFactory := dynamic.NewMessageFactoryWithDefaults()
	message := messageFactory.NewMessage(messageDescriptor)

	// Convert from Json string to proto message type.
	if err := jsonpb.UnmarshalString(str, message); err != nil {
		return nil, err
	}

	// Serialize proto message type to binary format.
	data, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}
	data = append(indexBytes, data...)
	return data, nil
}

type ProtoDeserializationProvider uint32

func (protoProvider *ProtoDeserializationProvider) GetSchemaName() string {
	return "PROTOBUF"
}

func (protoProvider *ProtoDeserializationProvider) decode(data []byte, schemaPath string) (string, error) {
	// Index array indicates which message in the file we're referring to.
	// In our case, we simply ignore the index array [0].
	data = data[1:]
	parser := parse.Parser{}
	fileDescriptors, err := parser.ParseFiles(schemaPath)
	if err != nil {
		return "", err
	}
	if len(fileDescriptors) == 0 {
		return "", errors.New("No valid file entered.")
	}
	fileDescriptor := fileDescriptors[0]

	messageDescriptors := fileDescriptor.GetMessageTypes()
	if len(messageDescriptors) == 0 {
		return "", errors.New("No valid message in the file.")
	}
	// We're always using the outermost first message.
	messageDescriptor := messageDescriptors[0]
	messageFactory := dynamic.NewMessageFactoryWithDefaults()
	message := messageFactory.NewMessage(messageDescriptor)

	// Convert from binary format to proto message type.
	err = proto.Unmarshal(data, message)
	if err != nil {
		return "", err
	}

	// Convert from proto message type to Json string.
	marshaler := &jsonpb.Marshaler{}
	str, err := marshaler.MarshalToString(message)
	if err != nil {
		return "", err
	}

	return str, nil
}
