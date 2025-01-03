package exec

import (
	"fmt"
	"strings"

	"github.com/cloudposse/atmos/pkg/config"
	"github.com/cloudposse/atmos/pkg/schema"
)

func ProcessCustomYamlTags(cliConfig schema.CliConfiguration, input schema.AtmosSectionMapType) (schema.AtmosSectionMapType, error) {
	return processNodes(cliConfig, input), nil
}

func processNodes(cliConfig schema.CliConfiguration, data map[string]any) map[string]any {
	newMap := make(map[string]any)
	var recurse func(any) any

	recurse = func(node any) any {
		switch v := node.(type) {
		case string:
			return processCustomTags(cliConfig, v)

		case map[string]any:
			newNestedMap := make(map[string]any)
			for k, val := range v {
				newNestedMap[k] = recurse(val)
			}
			return newNestedMap

		case []any:
			newSlice := make([]any, len(v))
			for i, val := range v {
				newSlice[i] = recurse(val)
			}
			return newSlice

		default:
			return v
		}
	}

	for k, v := range data {
		newMap[k] = recurse(v)
	}

	return newMap
}

func processCustomTags(cliConfig schema.CliConfiguration, input string) any {
	if strings.HasPrefix(input, config.AtmosYamlFuncTemplate) {
		return processTagTemplate(cliConfig, input)
	} else if strings.HasPrefix(input, config.AtmosYamlFuncExec) {
		return processTagExec(cliConfig, input)
	} else if strings.HasPrefix(input, config.AtmosYamlFuncTerraformOutput) {
		return processTagTerraformOutput(cliConfig, input)
	}

	// If any other YAML explicit type (not currently supported by Atmos) is used, return it w/o processing
	return input
}

func getStringAfterTag(cliConfig schema.CliConfiguration, input string, tag string) (string, error) {
	str := strings.TrimPrefix(input, tag)
	str = strings.TrimSpace(str)

	if str == "" {
		err := fmt.Errorf("invalid Atmos YAML function: %s", input)
		return "", err
	}

	return str, nil
}
