package providers

import (
	"encoding/json"
	"fmt"

	"github.com/obot-platform/obot/apiclient/types"
	v1 "github.com/obot-platform/obot/pkg/storage/apis/obot.obot.ai/v1"
)

func ConvertProviderToolRef(toolRef v1.ToolReference, cred map[string]string) (*types.CommonProviderStatus, error) {
	var (
		providerMeta   ProviderMeta
		missingEnvVars []string
	)
	if toolRef.Status.Tool != nil {
		if toolRef.Status.Tool.Metadata["providerMeta"] != "" {
			if err := json.Unmarshal([]byte(toolRef.Status.Tool.Metadata["providerMeta"]), &providerMeta); err != nil {
				return nil, fmt.Errorf("failed to unmarshal provider meta for %s: %v", toolRef.Name, err)
			}
		}

		for _, envVar := range providerMeta.EnvVars {
			if _, ok := cred[envVar.Name]; !ok {
				missingEnvVars = append(missingEnvVars, envVar.Name)
			}
		}
	}

	return &types.CommonProviderStatus{
		CommonProviderMetadata:          providerMeta.CommonProviderMetadata,
		Configured:                      toolRef.Status.Tool != nil && len(missingEnvVars) == 0,
		RequiredConfigurationParameters: providerMeta.EnvVars,
		OptionalConfigurationParameters: providerMeta.OptionalEnvVars,
		MissingConfigurationParameters:  missingEnvVars,
		Error:                           toolRef.Status.Error,
	}, nil
}
