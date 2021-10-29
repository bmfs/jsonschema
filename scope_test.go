package jsonschema_test

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/qri-io/jsonschema"
	"github.com/qri-io/jsonschema/testutil"
)

func TestUnscopedSchemas(t *testing.T) {
	// ensure schemas use unscoped registries
	jsonschema.UseScopedRegistries = false
	ctx := context.Background()

	ts := testutil.CreateHTTPServer()
	defer ts.Close()

	schema1 := `{
		"$defs":{
			"foo": {
				"type": "integer"
			}
		}
	}`
	ts.SetRoute("/valid_schema.json", schema1)

	jsonschema.LoadDraft2019_09()

	source := fmt.Sprintf(`{
		"allOf": [{
				"$ref": "%s/valid_schema.json#/$defs/foo"
		}]
	}`, ts.URL)

	_, err := validateSchema(ctx, source, "1")
	if err != nil {
		t.Errorf("failure to parse schema: %s", err.Error())
		return
	}

	schema2 := `{
		"$defs":{
			"foo": {
				"type": "string"
			}
		}
	}`
	ts.SetRoute("/valid_schema.json", schema2)

	schema2Errors, err := validateSchema(ctx, source, `"1"`)
	if err != nil {
		t.Errorf("failure to parse schema: %s", err.Error())
		return
	}

	routeCount := ts.GetCount("/valid_schema.json")
	if ts.GetCount("/valid_schema.json") != 1 {
		t.Errorf("unexpected # external schema loads. expected: 1, actual: %d", routeCount)
	}

	if len(schema2Errors) != 1 {
		t.Errorf("unexpected # of errors. expected: 1, actual: %d", len(schema2Errors))
	}
}

func TestScopedSchemas(t *testing.T) {
	// Make jsonschema use scoped registries
	jsonschema.UseScopedRegistries = true
	defer (func() { jsonschema.UseScopedRegistries = false })()

	ctx := context.Background()

	ts := testutil.CreateHTTPServer()
	defer ts.Close()

	schema1 := `{
		"$defs":{
			"foo": {
				"type": "integer"
			}
		}
	}`
	ts.SetRoute("/valid_schema.json", schema1)

	jsonschema.LoadDraft2019_09()

	source := fmt.Sprintf(`{
		"allOf": [{
				"$ref": "%s/valid_schema.json#/$defs/foo"
		}]
	}`, ts.URL)

	_, err := validateSchema(ctx, source, "1")
	if err != nil {
		t.Errorf("failure to parse schema: %s", err.Error())
		return
	}

	schema2 := `{
		"$defs":{
			"foo": {
				"type": "string"
			}
		}
	}`
	ts.SetRoute("/valid_schema.json", schema2)

	schema2Errors, err := validateSchema(ctx, source, `"1"`)
	if err != nil {
		t.Errorf("failure to parse schema: %s", err.Error())
		return
	}

	routeCount := ts.GetCount("/valid_schema.json")
	if ts.GetCount("/valid_schema.json") != 2 {
		t.Errorf("unexpected # external schema loads. expected: 1, actual: %d", routeCount)
	}

	if len(schema2Errors) != 0 {
		t.Errorf("unexpected # of errors. expected: 1, actual: %d", len(schema2Errors))
	}
}

func validateSchema(ctx context.Context, schemaSource string, data string) ([]jsonschema.KeyError, error) {
	schema := jsonschema.Schema{}
	err := schema.UnmarshalJSON([]byte(schemaSource))
	if err != nil {
		return nil, err
	}

	return schema.ValidateBytes(ctx, []byte(data))
}

// TestScopedLoaders tests the support to override
func TestScopedLoaders(t *testing.T) {
	// Make jsonschema use scoped registries
	jsonschema.UseScopedRegistries = true
	defer (func() { jsonschema.UseScopedRegistries = false })()

	// test data

	validSchemaSource := `{
		"$defs":{
			"foo": {
				"type": "string"
			}
		}
	}`

	mainSchemaSource := `{
		"allOf": [{
				"$ref": "protocol:///valid_schema.json#/$defs/foo"
		}]
	}`

	//

	ctx := context.Background()

	jsonschema.LoadDraft2019_09()

	timesGlobalLRWasCalled := 0
	glr := jsonschema.GetSchemaLoaderRegistry()
	glr.Register("protocol", func(ctx context.Context, uri *url.URL, schema *jsonschema.Schema) error {
		timesGlobalLRWasCalled += 1
		return schema.UnmarshalJSON([]byte(validSchemaSource))
	})

	schema := jsonschema.Schema{}
	// this only affects current schema instance
	lr := schema.GetSchemaRegistry().GetLoaderRegistry()
	lr.Register("protocol", func(ctx context.Context, uri *url.URL, schema *jsonschema.Schema) error {
		return schema.UnmarshalJSON([]byte(validSchemaSource))
	})

	err := schema.UnmarshalJSON([]byte(mainSchemaSource))
	if err != nil {
		t.Errorf("failure to unmarshal schema: %s", err.Error())
		return
	}

	errs, err := schema.ValidateBytes(ctx, []byte(`"1"`))
	if err != nil {
		t.Errorf("failure to parse schema: %s", err.Error())
		return
	}

	if len(errs) != 0 {
		t.Errorf("unexpected validation errors. expected 0; actual: %d", len(errs))
	}

	if timesGlobalLRWasCalled != 0 {
		t.Error("global LR was unexpected called")
	}

	// Confirm previous override did not affect global registry
	timesGlobalLRWasCalled = 0
	validateSchema(ctx, mainSchemaSource, `"1"`)

	if timesGlobalLRWasCalled == 0 {
		t.Error("expected global loader func to have been callend. It wasn't")
	}
}
