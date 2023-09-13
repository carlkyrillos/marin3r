package envoy

import (
	"github.com/3scale-ops/marin3r/pkg/apishelper"
	envoy_serializer "github.com/3scale-ops/marin3r/pkg/apishelper/serializer"
)

func Validate(resource string, encoding envoy_serializer.Serialization, version apishelper.APIVersion, rType apishelper.Type) error {
	decoder := envoy_serializer.NewResourceUnmarshaller(encoding, version)
	generator := NewGenerator(version)
	res := generator.New(rType)
	if err := decoder.Unmarshal(resource, res); err != nil {
		return err
	}

	return nil
}
