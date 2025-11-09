// Package bazel provides functionality for parsing BUILD.bazel files to extract
// GAPIC generation configuration.
package bazel

import (
	"fmt"
	"os"
	"strings"

	"github.com/bazelbuild/buildtools/build"
	"github.com/julieqiu/exp/librarian/internal/state"
)

// ParseBuildFile reads a BUILD.bazel file and extracts GAPIC library configuration
// for the specified language.
//
// Returns nil if no GAPIC rule is found (indicating a proto-only library).
func ParseBuildFile(buildPath string, language string) (*state.API, error) {
	data, err := os.ReadFile(buildPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read BUILD.bazel: %w", err)
	}

	file, err := build.ParseBuild("BUILD.bazel", data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse BUILD.bazel: %w", err)
	}

	// Find the language-specific GAPIC rule
	ruleSuffix := fmt.Sprintf("_%s_gapic", language)
	for _, rule := range file.Rules("") {
		if strings.HasSuffix(rule.Name(), ruleSuffix) {
			return extractAPIConfig(rule), nil
		}
	}

	// No GAPIC rule found - this is a proto-only library
	return nil, nil
}

// extractAPIConfig extracts API configuration from a BUILD rule
func extractAPIConfig(rule *build.Rule) *state.API {
	api := &state.API{}

	// Extract grpc_service_config
	if val := rule.AttrString("grpc_service_config"); val != "" {
		api.GrpcServiceConfig = val
	}

	// Extract service_yaml
	if val := rule.AttrString("service_yaml"); val != "" {
		api.ServiceYaml = val
	}

	// Extract transport
	if val := rule.AttrString("transport"); val != "" {
		api.Transport = val
	}

	// Extract rest_numeric_enums (boolean)
	if attr := rule.Attr("rest_numeric_enums"); attr != nil {
		// The attribute value is a boolean literal in Starlark
		if ident, ok := attr.(*build.Ident); ok && ident.Name == "True" {
			api.RestNumericEnums = true
		}
	}

	// Extract opt_args (list of strings)
	if attr := rule.Attr("opt_args"); attr != nil {
		if list, ok := attr.(*build.ListExpr); ok {
			for _, elem := range list.List {
				if strLit, ok := elem.(*build.StringExpr); ok {
					api.OptArgs = append(api.OptArgs, strLit.Value)
				}
			}
		}
	}

	return api
}
