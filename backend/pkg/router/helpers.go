package router

import (
	"errors"
	"fmt"
	"http-mqtt-boilerplate/backend/pkg/generate"
	"slices"
	"strings"
)

// validateRouteSpec validates a RouteSpec.
func validateRouteSpec(spec RouteSpec) error {
	if spec.OperationID == "" {
		return errors.New("field OperationID required")
	}

	if spec.Summary == "" {
		return errors.New("field Summary required")
	}

	if spec.Description == "" {
		return errors.New("field Description required")
	}

	if spec.Group == "" {
		return errors.New("field Group required")
	}

	if spec.Handler == nil {
		return errors.New("field Handler required")
	}

	return nil
}

func generateParameters(spec RouteSpec) ([]generate.ParameterInfo, error) {
	var parameters []generate.ParameterInfo

	// Validate path parameters and collect metadata
	paramsInPath := map[string]struct{}{}
	documentedPathParams := map[string]struct{}{}

	// Extract param names from path
	for section := range strings.SplitSeq(spec.fullPath, "/") {
		paramsName, err := generate.ExtractParamName(section)
		if err != nil {
			return nil, fmt.Errorf("invalid path %s: %w", spec.fullPath, err)
		}

		for _, paramName := range paramsName {
			if !generate.IsValidParameterName(paramName) {
				return nil, fmt.Errorf("invalid parameter name %s in path %s", paramName, spec.fullPath)
			}
			paramsInPath[paramName] = struct{}{}
		}
	}

	// For each documented parameter, validate and collect metadata
	for name, paramSpec := range spec.Parameters {
		if name == "" {
			return nil, fmt.Errorf("parameter name required for %s %s", spec.method, spec.fullPath)
		}

		if paramSpec.Description == "" {
			return nil, fmt.Errorf("parameter Description required for %s %s", spec.method, spec.fullPath)
		}

		if paramSpec.Type == nil {
			return nil, fmt.Errorf("parameter Type required for %s %s", spec.method, spec.fullPath)
		}

		validInValues := []ParameterIn{ParameterInPath, ParameterInQuery, ParameterInHeader}
		if !slices.Contains(validInValues, paramSpec.In) {
			return nil, fmt.Errorf("parameter In must be one of %v for %s %s", validInValues, spec.method, spec.fullPath)
		}

		parameters = append(parameters, generate.ParameterInfo{
			Name:        name,
			In:          string(paramSpec.In),
			TypeValue:   paramSpec.Type,
			Description: paramSpec.Description,
			Required:    paramSpec.Required,
		})

		if paramSpec.In == ParameterInPath {
			if _, exists := paramsInPath[name]; !exists {
				return nil, fmt.Errorf("documented path parameter %s not found in path", name)
			}

			if !paramSpec.Required {
				return nil, fmt.Errorf("path parameter %s must be required", name)
			}

			documentedPathParams[name] = struct{}{}
		}

	}

	// Now go over all discovered path parameters and validate that they are documented
	for name := range paramsInPath {
		if _, exists := documentedPathParams[name]; !exists {
			return nil, fmt.Errorf("path parameter %s not documented", name)
		}
	}

	return parameters, nil
}
