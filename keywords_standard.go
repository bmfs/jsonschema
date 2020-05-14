package jsonschema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	jptr "github.com/qri-io/jsonpointer"
)

// Const defines the const JSON Schema keyword
type Const json.RawMessage

// NewConst allocates a new Const keyword
func NewConst() Keyword {
	return &Const{}
}

// Register implements the Keyword interface for Const
func (c *Const) Register(uri string, registry *SchemaRegistry) {}

// Resolve implements the Keyword interface for Const
func (c *Const) Resolve(pointer jptr.Pointer, uri string) *Schema {
	return nil
}

// ValidateFromContext implements the Keyword interface for Const
func (c Const) ValidateFromContext(schCtx *SchemaContext, errs *[]KeyError) {
	SchemaDebug("[Const] Validating")
	var con interface{}
	if err := json.Unmarshal(c, &con); err != nil {
		AddErrorCtx(errs, schCtx, err.Error())
		return
	}

	if !reflect.DeepEqual(con, schCtx.Instance) {
		AddErrorCtx(errs, schCtx, fmt.Sprintf(`must equal %s`, InvalidValueString(con)))
	}
}

// JSONProp implements the JSONPather for Const
func (c Const) JSONProp(name string) interface{} {
	return nil
}

// String implements the Stringer for Const
func (c Const) String() string {
	return string(c)
}

// UnmarshalJSON implements the json.Unmarshaler interface for Const
func (c *Const) UnmarshalJSON(data []byte) error {
	*c = data
	return nil
}

// MarshalJSON implements the json.Marshaler interface for Const
func (c Const) MarshalJSON() ([]byte, error) {
	return json.Marshal(json.RawMessage(c))
}

// Enum defines the enum JSON Schema keyword
type Enum []Const

// NewEnum allocates a new Enum keyword
func NewEnum() Keyword {
	return &Enum{}
}

// Register implements the Keyword interface for Enum
func (e *Enum) Register(uri string, registry *SchemaRegistry) {}

// Resolve implements the Keyword interface for Enum
func (e *Enum) Resolve(pointer jptr.Pointer, uri string) *Schema {
	return nil
}

// ValidateFromContext implements the Keyword interface for Enum
func (e Enum) ValidateFromContext(schCtx *SchemaContext, errs *[]KeyError) {
	SchemaDebug("[Enum] Validating")
	for _, v := range e {
		test := &[]KeyError{}
		v.ValidateFromContext(schCtx, test)
		if len(*test) == 0 {
			return
		}
	}

	AddErrorCtx(errs, schCtx, fmt.Sprintf("should be one of %s", e.String()))
}

// JSONProp implements the JSONPather for Enum
func (e Enum) JSONProp(name string) interface{} {
	idx, err := strconv.Atoi(name)
	if err != nil {
		return nil
	}
	if idx > len(e) || idx < 0 {
		return nil
	}
	return e[idx]
}

// JSONChildren implements the JSONContainer interface for Enum
func (e Enum) JSONChildren() (res map[string]JSONPather) {
	res = map[string]JSONPather{}
	for i, bs := range e {
		res[strconv.Itoa(i)] = bs
	}
	return
}

// String implements the Stringer for Enum
func (e Enum) String() string {
	str := "["
	for _, c := range e {
		str += c.String() + ", "
	}
	return str[:len(str)-2] + "]"
}

// List of primitive types supported and used by JSON Schema
var primitiveTypes = map[string]bool{
	"null":    true,
	"boolean": true,
	"object":  true,
	"array":   true,
	"number":  true,
	"string":  true,
	"integer": true,
}

// DataType attempts to parse the underlying data type
// from the raw data interface
func DataType(data interface{}) string {
	if data == nil {
		return "null"
	}

	switch reflect.TypeOf(data).Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Float64:
		number := reflect.ValueOf(data).Float()
		if float64(int(number)) == number {
			return "integer"
		}
		return "number"
	case reflect.String:
		return "string"
	case reflect.Array, reflect.Slice:
		return "array"
	case reflect.Map, reflect.Struct:
		return "object"
	default:
		return "unknown"
	}
}

// DataTypeWithHint attempts to parse the underlying data type
// by leveraging the schema expectations for better results
func DataTypeWithHint(data interface{}, hint string) string {
	dt := DataType(data)
	if dt == "string" {
		if hint == "boolean" {
			_, err := strconv.ParseBool(data.(string))
			if err == nil {
				return "boolean"
			}
		}
	}
	// deals with traling 0 floats
	if dt == "integer" && hint == "number" {
		return "number"
	}
	return dt
}

// Type defines the type JSON Schema keyword
type Type struct {
	strVal bool
	vals   []string
}

// NewType allocates a new Type keyword
func NewType() Keyword {
	return &Type{}
}

// Register implements the Keyword interface for Type
func (t *Type) Register(uri string, registry *SchemaRegistry) {}

// Resolve implements the Keyword interface for Type
func (t *Type) Resolve(pointer jptr.Pointer, uri string) *Schema {
	return nil
}

// ValidateFromContext implements the Keyword interface for Type
func (t Type) ValidateFromContext(schCtx *SchemaContext, errs *[]KeyError) {
	SchemaDebug("[Type] Validating")
	jt := DataType(schCtx.Instance)
	for _, typestr := range t.vals {
		if jt == typestr || jt == "integer" && typestr == "number" {
			return
		}
		if jt == "string" && (typestr == "boolean" || typestr == "number" || typestr == "integer") {
			if DataTypeWithHint(schCtx.Instance, typestr) == typestr {
				return
			}
		}
	}
	if len(t.vals) == 1 {
		AddErrorCtx(errs, schCtx, fmt.Sprintf(`type should be %s`, t.vals[0]))
		return
	}

	str := ""
	for _, ts := range t.vals {
		str += ts + ","
	}

	AddErrorCtx(errs, schCtx, fmt.Sprintf(`type should be one of: %s`, str[:len(str)-1]))
}

// String implements the Stringer for Type
func (t Type) String() string {
	if len(t.vals) == 0 {
		return "unknown"
	}
	return strings.Join(t.vals, ",")
}

// JSONProp implements the JSONPather for Type
func (t Type) JSONProp(name string) interface{} {
	idx, err := strconv.Atoi(name)
	if err != nil {
		return nil
	}
	if idx > len(t.vals) || idx < 0 {
		return nil
	}
	return t.vals[idx]
}

// UnmarshalJSON implements the json.Unmarshaler interface for Type
func (t *Type) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*t = Type{strVal: true, vals: []string{single}}
	} else {
		var set []string
		if err := json.Unmarshal(data, &set); err == nil {
			*t = Type{vals: set}
		} else {
			return err
		}
	}

	for _, pr := range t.vals {
		if !primitiveTypes[pr] {
			return fmt.Errorf(`"%s" is not a valid type`, pr)
		}
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface for Type
func (t Type) MarshalJSON() ([]byte, error) {
	if t.strVal {
		return json.Marshal(t.vals[0])
	}
	return json.Marshal(t.vals)
}