package check

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

// The published schemastore schemas are vendored so the check stays hermetic: a
// green build never depends on schemastore.org being reachable, and an upstream
// schema edit cannot turn this repository red without a commit here.
//
// Refresh both files with:
//
//	curl -fsSL https://json.schemastore.org/claude-code-marketplace.json \
//	  -o internal/check/schemas/claude-code-marketplace.json
//	curl -fsSL https://json.schemastore.org/claude-code-plugin-manifest.json \
//	  -o internal/check/schemas/claude-code-plugin-manifest.json
//
//go:embed schemas/claude-code-marketplace.json
var marketplaceSchemaJSON []byte

//go:embed schemas/claude-code-plugin-manifest.json
var pluginSchemaJSON []byte

// schemaKeywords is the JSON Schema (draft-07) vocabulary this validator
// implements. The vendored schemas use no other assertion keyword, so a refresh
// that introduces one fails loudly instead of being silently ignored.
var schemaKeywords = map[string]bool{
	"type": true, "const": true, "enum": true,
	"pattern": true, "minLength": true, "maxLength": true,
	"minimum": true, "maximum": true, "exclusiveMinimum": true, "exclusiveMaximum": true,
	"required": true, "properties": true, "additionalProperties": true, "propertyNames": true,
	"items": true, "anyOf": true, "allOf": true, "not": true,
}

// annotationKeywords carry no validation semantics.
var annotationKeywords = map[string]bool{
	"$id": true, "$schema": true, "$comment": true,
	"title": true, "description": true, "default": true, "format": true, "examples": true,
}

// loadSchema decodes a vendored schema document.
func loadSchema(raw []byte, name string) (map[string]any, error) {
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		return nil, fmt.Errorf("vendored %s schema: %w", name, err)
	}
	return schema, nil
}

// validateDocument reports every violation of schema in doc, as JSON-pointer
// keyed messages.
func validateDocument(schema map[string]any, doc any) []string {
	var problems []string
	validateValue(schema, doc, "", &problems)
	sort.Strings(problems)
	return problems
}

func validateValue(schema map[string]any, value any, path string, problems *[]string) {
	here := path
	if here == "" {
		here = "<root>"
	}

	for keyword := range schema {
		if !schemaKeywords[keyword] && !annotationKeywords[keyword] {
			*problems = append(*problems, fmt.Sprintf("%s: schema uses unsupported keyword %q", here, keyword))
		}
	}

	if want, present := schema["type"]; present {
		if types := typeNames(want); len(types) > 0 && !anyTypeMatches(types, value) {
			*problems = append(*problems, fmt.Sprintf("%s: want %s, got %s", here, strings.Join(types, " or "), jsonType(value)))
			return
		}
	}

	if want, present := schema["const"]; present && !reflect.DeepEqual(want, value) {
		*problems = append(*problems, fmt.Sprintf("%s: must equal %v", here, want))
	}
	if raw, present := schema["enum"]; present {
		allowed := values(raw)
		if !anyEqual(allowed, value) {
			*problems = append(*problems, fmt.Sprintf("%s: %v is not one of %v", here, value, allowed))
		}
	}

	switch typed := value.(type) {
	case string:
		validateString(schema, typed, here, problems)
	case float64:
		validateNumber(schema, typed, here, problems)
	case map[string]any:
		validateObject(schema, typed, here, problems)
	case []any:
		validateArray(schema, typed, here, problems)
	}

	if conjunction := schemasOf(schema["allOf"]); conjunction != nil {
		for _, member := range conjunction {
			validateValue(member, value, path, problems)
		}
	}
	if negation, ok := schema["not"].(map[string]any); ok {
		var inner []string
		validateValue(negation, value, path, &inner)
		if len(inner) == 0 {
			*problems = append(*problems, fmt.Sprintf("%s: must not match the excluded shape", here))
		}
	}
	if alternatives := schemasOf(schema["anyOf"]); alternatives != nil {
		matched := false
		for _, alternative := range alternatives {
			var branch []string
			validateValue(alternative, value, path, &branch)
			if len(branch) == 0 {
				matched = true
				break
			}
		}
		if !matched {
			*problems = append(*problems, fmt.Sprintf("%s: matches none of the %d allowed shapes", here, len(alternatives)))
		}
	}
}

func validateString(schema map[string]any, value, here string, problems *[]string) {
	length := float64(len([]rune(value)))
	if min, ok := schema["minLength"].(float64); ok && length < min {
		*problems = append(*problems, fmt.Sprintf("%s: %q is %g characters; minimum is %g", here, value, length, min))
	}
	if max, ok := schema["maxLength"].(float64); ok && length > max {
		*problems = append(*problems, fmt.Sprintf("%s: %q is %g characters; maximum is %g", here, value, length, max))
	}
	pattern, ok := schema["pattern"].(string)
	if !ok {
		return
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		*problems = append(*problems, fmt.Sprintf("%s: schema pattern %q does not compile: %v", here, pattern, err))
		return
	}
	if !re.MatchString(value) {
		*problems = append(*problems, fmt.Sprintf("%s: %q does not match %s", here, value, pattern))
	}
}

func validateNumber(schema map[string]any, value float64, here string, problems *[]string) {
	bounds := []struct {
		keyword  string
		relation string
		violated func(value, limit float64) bool
	}{
		{"minimum", "at least", func(v, l float64) bool { return v < l }},
		{"maximum", "at most", func(v, l float64) bool { return v > l }},
		{"exclusiveMinimum", "greater than", func(v, l float64) bool { return v <= l }},
		{"exclusiveMaximum", "less than", func(v, l float64) bool { return v >= l }},
	}
	for _, bound := range bounds {
		limit, ok := schema[bound.keyword].(float64)
		if ok && bound.violated(value, limit) {
			*problems = append(*problems, fmt.Sprintf("%s: %g must be %s %g", here, value, bound.relation, limit))
		}
	}
}

func validateObject(schema map[string]any, value map[string]any, here string, problems *[]string) {
	properties, _ := schema["properties"].(map[string]any)

	for _, raw := range values(schema["required"]) {
		name, ok := raw.(string)
		if !ok {
			continue
		}
		if _, present := value[name]; !present {
			*problems = append(*problems, fmt.Sprintf("%s: missing required property %q", here, name))
		}
	}

	if names, ok := schema["propertyNames"].(map[string]any); ok {
		for key := range value {
			var inner []string
			validateValue(names, key, here, &inner)
			if len(inner) > 0 {
				*problems = append(*problems, fmt.Sprintf("%s: property name %q is not allowed", here, key))
			}
		}
	}

	switch additional := schema["additionalProperties"].(type) {
	case bool:
		if !additional {
			for key := range value {
				if _, declared := properties[key]; !declared {
					*problems = append(*problems, fmt.Sprintf("%s: property %q is not allowed", here, key))
				}
			}
		}
	case map[string]any:
		for key, member := range value {
			if _, declared := properties[key]; declared {
				continue
			}
			validateValue(additional, member, joinPointer(pathOf(here), key), problems)
		}
	}

	for key, sub := range properties {
		member, present := value[key]
		if !present {
			continue
		}
		subSchema, ok := sub.(map[string]any)
		if !ok {
			continue
		}
		validateValue(subSchema, member, joinPointer(pathOf(here), key), problems)
	}
}

func validateArray(schema map[string]any, value []any, here string, problems *[]string) {
	switch items := schema["items"].(type) {
	case map[string]any:
		for i, member := range value {
			validateValue(items, member, joinPointer(pathOf(here), fmt.Sprint(i)), problems)
		}
	case []any:
		for i, sub := range items {
			if i >= len(value) {
				break
			}
			if subSchema, ok := sub.(map[string]any); ok {
				validateValue(subSchema, value[i], joinPointer(pathOf(here), fmt.Sprint(i)), problems)
			}
		}
	}
}

// pathOf turns the display form of the root back into an empty pointer base.
func pathOf(here string) string {
	if here == "<root>" {
		return ""
	}
	return here
}

// schemasOf coerces a keyword value into a list of schemas, or nil when the
// keyword is absent or empty.
func schemasOf(raw any) []map[string]any {
	list, ok := raw.([]any)
	if !ok {
		return nil
	}
	schemas := make([]map[string]any, 0, len(list))
	for _, item := range list {
		if schema, ok := item.(map[string]any); ok {
			schemas = append(schemas, schema)
		}
	}
	if len(schemas) == 0 {
		return nil
	}
	return schemas
}

func values(raw any) []any {
	list, _ := raw.([]any)
	return list
}

func anyEqual(candidates []any, value any) bool {
	for _, candidate := range candidates {
		if reflect.DeepEqual(candidate, value) {
			return true
		}
	}
	return false
}

func typeNames(raw any) []string {
	switch typed := raw.(type) {
	case string:
		return []string{typed}
	case []any:
		var names []string
		for _, item := range typed {
			if name, ok := item.(string); ok {
				names = append(names, name)
			}
		}
		return names
	}
	return nil
}

func anyTypeMatches(names []string, value any) bool {
	for _, name := range names {
		if matchesType(name, value) {
			return true
		}
	}
	return false
}

func matchesType(name string, value any) bool {
	switch name {
	case "object":
		_, ok := value.(map[string]any)
		return ok
	case "array":
		_, ok := value.([]any)
		return ok
	case "string":
		_, ok := value.(string)
		return ok
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "number":
		_, ok := value.(float64)
		return ok
	case "integer":
		number, ok := value.(float64)
		return ok && number == math.Trunc(number)
	case "null":
		return value == nil
	}
	return false
}

func jsonType(value any) string {
	switch value.(type) {
	case nil:
		return "null"
	case string:
		return "string"
	case bool:
		return "boolean"
	case float64:
		return "number"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	}
	return fmt.Sprintf("%T", value)
}

func joinPointer(base, token string) string {
	escaped := strings.ReplaceAll(strings.ReplaceAll(token, "~", "~0"), "/", "~1")
	return base + "/" + escaped
}
