package apishelper

import (
	"github.com/3scale-ops/marin3r/pkg/apishelper"
	envoy_serializer_v3 "github.com/3scale-ops/marin3r/pkg/apishelper/serializer/v3"
)

// Serialization represents a serialization encoding for envoy.Resource structs.
type Serialization string

const (
	// YAML represents yaml serialization of envoy.Resource structs.
	YAML Serialization = "yaml"
	// JSON represents json serialization of envoy.Resource structs.
	JSON Serialization = "json"
	// B64JSON represents yaml base64 encpded json serizalization of envoy.Resource structs.
	B64JSON Serialization = "b64json"
)

// ResourceMarshaller serialize a protobuf struct into json
type ResourceMarshaller interface {
	Marshal(apishelper.Resource) (string, error)
}

// ResourceUnmarshaller deserialize from json into a protobuf struct
type ResourceUnmarshaller interface {
	Unmarshal(string, apishelper.Resource) error
}

// NewResourceMarshaller returns a ResourceMarshaller for the given API version and encoding
func NewResourceMarshaller(encoding Serialization, version apishelper.APIVersion) ResourceMarshaller {
	return envoy_serializer_v3.JSON{}

}

// NewResourceUnmarshaller returns a ResourceUnmarshaller for the given api version and encoding
func NewResourceUnmarshaller(encoding Serialization, version apishelper.APIVersion) ResourceUnmarshaller {
	switch encoding {
	case JSON:
		return envoy_serializer_v3.JSON{}
	case YAML:
		return envoy_serializer_v3.YAML{}
	case B64JSON:
		return envoy_serializer_v3.B64JSON{}
	}
	return nil
}
