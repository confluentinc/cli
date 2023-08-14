package serdes

import (
	"github.com/golang/protobuf/jsonpb" //nolint:staticcheck // deprecated module cannot be removed due to https://github.com/jhump/protoreflect/issues/301
	"github.com/golang/protobuf/proto"  //nolint:staticcheck // deprecated module cannot be removed due to https://github.com/jhump/protoreflect/issues/301

	"github.com/confluentinc/cli/internal/pkg/errors"
)

type ProtobufDeserializationProvider struct {
	message proto.Message
}

func (p *ProtobufDeserializationProvider) LoadSchema(schemaPath string, referencePathMap map[string]string) error {
	message, err := parseMessage(schemaPath, referencePathMap)
	if err != nil {
		return err
	}
	p.message = message
	return nil
}

func (p *ProtobufDeserializationProvider) Deserialize(data []byte) (string, error) {
	// Index array indicates which message in the file we're referring to.
	// In our case, we simply ignore the index array [0].
	data = data[1:]

	// Convert from binary format to proto message type.
	if err := proto.Unmarshal(data, p.message); err != nil {
		return "", errors.New(errors.ProtoDocumentInvalidErrorMsg)
	}

	// Convert from proto message type to JSON string.
	marshaler := &jsonpb.Marshaler{}
	str, err := marshaler.MarshalToString(p.message)
	if err != nil {
		return "", err
	}

	return str, nil
}
