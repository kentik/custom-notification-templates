//go:generate go run ../../cmd/codegen/main.go -pkg render -dir . -output metadata_gen.go -extra-dirs ../types

package render

import (
	"reflect"

	"github.com/kentik/custom-notification-templates/pkg/types"
)

// SchemaField represents a field in the view model
type SchemaField struct {
	Name        string         `json:"name"`
	Type        string         `json:"type"`
	Path        string         `json:"path"`
	Description string         `json:"description,omitempty"`
	IsArray     bool           `json:"isArray,omitempty"`
	ElementType string         `json:"elementType,omitempty"`
	Children    []*SchemaField `json:"children,omitempty"`
	IsMethod    bool           `json:"isMethod,omitempty"`
	ReturnType  string         `json:"returnType,omitempty"`
	EnumType    string         `json:"enumType,omitempty"`
}

// enumFieldMapping maps field names to their corresponding enum types
var enumFieldMapping = map[string]string{
	"Type":       "EventType",
	"Tag":        "DetailTag",
	"Importance": "ViewModelImportance",
}

// SchemaFunction represents a template function
type SchemaFunction struct {
	Name        string `json:"name"`
	Signature   string `json:"signature"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
}

// SchemaEnum represents an enumeration type
type SchemaEnum struct {
	Values      []string `json:"values"`
	Description string   `json:"description,omitempty"`
}

// Schema represents the complete template schema
type Schema struct {
	Fields    []*SchemaField         `json:"fields"`
	Functions []*SchemaFunction      `json:"functions"`
	Enums     map[string]*SchemaEnum `json:"enums"`
}

// GetSchema returns the complete schema for template editing
func GetSchema() *Schema {
	return &Schema{
		Fields:    extractFields(),
		Functions: extractFunctions(),
		Enums:     extractEnums(),
	}
}

// extractFields uses reflection to get all fields from NotificationViewModel
func extractFields() []*SchemaField {
	var result []*SchemaField

	t := reflect.TypeOf(types.NotificationViewModel{})
	result = append(result, extractTypeFields(t, ".")...)

	// Also extract methods that are commonly used
	pt := reflect.TypeOf(&types.NotificationViewModel{})
	result = append(result, extractTypeMethods(pt, ".")...)

	return result
}

func extractTypeFields(t reflect.Type, path string) []*SchemaField {
	var result []*SchemaField

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		sf := &SchemaField{
			Name:        field.Name,
			Path:        path,
			Description: field.Tag.Get("description"),
		}

		fieldType := field.Type

		// Handle pointer types
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		// Handle slices
		if fieldType.Kind() == reflect.Slice {
			sf.IsArray = true
			elemType := fieldType.Elem()
			if elemType.Kind() == reflect.Ptr {
				elemType = elemType.Elem()
			}
			sf.ElementType = elemType.Name()
			sf.Type = "[]" + elemType.Name()

			// Build collection path (without [])
			collectionPath := path
			if collectionPath == "." {
				collectionPath = "." + field.Name
			} else {
				collectionPath = collectionPath + "." + field.Name
			}

			// Extract methods on the collection type itself (e.g., EventViewModelDetails methods)
			// Use the original field type which may be a named slice type with methods
			originalFieldType := field.Type
			if originalFieldType.Kind() == reflect.Ptr {
				originalFieldType = originalFieldType.Elem()
			}
			sf.Children = extractTypeMethods(reflect.PtrTo(originalFieldType), collectionPath)

			// Extract children for struct element types
			if elemType.Kind() == reflect.Struct && !isBasicType(elemType) {
				elemPath := collectionPath + "[]"
				sf.Children = append(sf.Children, extractTypeFields(elemType, elemPath)...)
				sf.Children = append(sf.Children, extractTypeMethods(reflect.PtrTo(elemType), elemPath)...)
			}
		} else if fieldType.Kind() == reflect.Struct && !isBasicType(fieldType) {
			sf.Type = fieldType.Name()
			childPath := path
			if childPath == "." {
				childPath = "." + field.Name
			} else {
				childPath = childPath + "." + field.Name
			}
			sf.Children = extractTypeFields(fieldType, childPath)
			sf.Children = append(sf.Children, extractTypeMethods(reflect.PtrTo(fieldType), childPath)...)
		} else {
			sf.Type = getTypeName(fieldType)
		}

		// Set enum type if this field maps to a known enum
		if enumType, ok := enumFieldMapping[field.Name]; ok {
			sf.EnumType = enumType
		}

		result = append(result, sf)
	}

	return result
}

func extractTypeMethods(t reflect.Type, path string) []*SchemaField {
	var result []*SchemaField

	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)

		// Skip unexported methods
		if !method.IsExported() {
			continue
		}

		// Only include methods with no arguments (beyond receiver) or simple args
		methodType := method.Type
		numIn := methodType.NumIn()
		if numIn > 2 { // receiver + at most 1 arg
			continue
		}

		numOut := methodType.NumOut()
		if numOut == 0 || numOut > 2 {
			continue
		}

		sf := &SchemaField{
			Name:        method.Name,
			Path:        path,
			IsMethod:    true,
			Description: getMethodDescription(t.Elem().Name(), method.Name),
		}

		// Build return type
		var returnType reflect.Type
		if numOut == 1 {
			returnType = methodType.Out(0)
			sf.ReturnType = getTypeName(returnType)
			sf.Type = sf.ReturnType
		} else if numOut == 2 {
			returnType = methodType.Out(0)
			sf.ReturnType = getTypeName(returnType)
			sf.Type = sf.ReturnType
		}

		// Expand struct return types to include their children
		// This allows traversing method chains like .Event.Details.HasTag
		if returnType != nil {
			// Dereference pointer return types
			if returnType.Kind() == reflect.Ptr {
				returnType = returnType.Elem()
			}

			// For struct return types, include their fields and methods as children
			if returnType.Kind() == reflect.Struct && !isBasicType(returnType) {
				childPath := path
				if childPath == "." {
					childPath = "." + method.Name
				} else {
					childPath = childPath + "." + method.Name
				}
				sf.Children = extractTypeFields(returnType, childPath)
				sf.Children = append(sf.Children, extractTypeMethods(reflect.PtrTo(returnType), childPath)...)
			}
		}

		result = append(result, sf)
	}

	return result
}

func isBasicType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String:
		return true
	}
	// Also treat time.Time as basic
	if t.PkgPath() == "time" && t.Name() == "Time" {
		return true
	}
	return false
}

func getTypeName(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "int"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "int"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.String:
		return "string"
	case reflect.Slice:
		return "[]" + getTypeName(t.Elem())
	case reflect.Map:
		return "map[" + getTypeName(t.Key()) + "]" + getTypeName(t.Elem())
	case reflect.Ptr:
		return "*" + getTypeName(t.Elem())
	case reflect.Interface:
		return "interface{}"
	case reflect.Struct:
		if t.PkgPath() == "time" && t.Name() == "Time" {
			return "time.Time"
		}
		return t.Name()
	default:
		return t.String()
	}
}

// ImportanceValueMap provides int to string mapping for enum values
var ImportanceValueMap = map[types.ViewModelImportance]string{
	types.ViewModelImportance_None:     "None",
	types.ViewModelImportance_Healthy:  "Healthy",
	types.ViewModelImportance_Notice:   "Notice",
	types.ViewModelImportance_Minor:    "Minor",
	types.ViewModelImportance_Warning:  "Warning",
	types.ViewModelImportance_Major:    "Major",
	types.ViewModelImportance_Severe:   "Severe",
	types.ViewModelImportance_Critical: "Critical",
}
