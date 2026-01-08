package render

import (
	"reflect"
	"sort"
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

	t := reflect.TypeOf(NotificationViewModel{})
	result = append(result, extractTypeFields(t, ".")...)

	// Also extract methods that are commonly used
	pt := reflect.TypeOf(&NotificationViewModel{})
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

			// Extract children for struct element types
			if elemType.Kind() == reflect.Struct && !isBasicType(elemType) {
				childPath := path
				if childPath == "." {
					childPath = "." + field.Name + "[]"
				} else {
					childPath = childPath + "." + field.Name + "[]"
				}
				sf.Children = extractTypeFields(elemType, childPath)
				sf.Children = append(sf.Children, extractTypeMethods(reflect.PtrTo(elemType), childPath)...)
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
		if numOut == 1 {
			sf.ReturnType = getTypeName(methodType.Out(0))
			sf.Type = sf.ReturnType
		} else if numOut == 2 {
			sf.ReturnType = getTypeName(methodType.Out(0))
			sf.Type = sf.ReturnType
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


func getMethodDescription(typeName, methodName string) string {
	key := typeName + "." + methodName

	descriptions := map[string]string{
		// NotificationViewModel methods
		"NotificationViewModel.Event":                  "Returns the first event (or nil if empty)",
		"NotificationViewModel.Events":                 "Returns all events as a slice",
		"NotificationViewModel.IsSingleEvent":          "True if exactly one event",
		"NotificationViewModel.IsMultipleEvents":       "True if more than one event",
		"NotificationViewModel.IsAtLeastOneEvent":      "True if at least one event exists",
		"NotificationViewModel.ActiveCount":            "Count of currently active events",
		"NotificationViewModel.InactiveCount":          "Count of inactive events",
		"NotificationViewModel.IsInsightsOnly":         "True if all events are insights",
		"NotificationViewModel.IsSyntheticsOnly":       "True if all events are synthetics",
		"NotificationViewModel.IsSynthOnly":            "Alias for IsSyntheticsOnly",
		"NotificationViewModel.IsSingleCustomInsightOnly": "True if single custom insight event",
		"NotificationViewModel.Headline":               "Generated headline text",
		"NotificationViewModel.Summary":                "Generated summary text",
		"NotificationViewModel.BasePortalURL":          "Returns portal URL",
		"NotificationViewModel.NotificationsSettingsURL": "Returns notifications settings URL",
		"NotificationViewModel.SyntheticsDashboardURL": "Returns synthetics dashboard URL",
		"NotificationViewModel.NowDate":                "Current date formatted as 'January 2, 2006'",
		"NotificationViewModel.NowRFC3339":             "Current time in RFC3339 format",
		"NotificationViewModel.NowDatetime":            "Current time as '2006-01-02 15:04:05 UTC'",
		"NotificationViewModel.NowUnix":                "Current time as Unix timestamp",
		"NotificationViewModel.Copyrights":             "Copyright string with current year",

		// EventViewModel methods
		"EventViewModel.IsAlarm":         "True if event type is alarm",
		"EventViewModel.IsInsight":       "True if event type is insight or custom-insight",
		"EventViewModel.IsCustomInsight": "True if event type is custom-insight",
		"EventViewModel.IsMitigation":    "True if event type is mitigation",
		"EventViewModel.IsSynthetic":     "True if event type is synthetic",

		// EventViewModelDetails methods
		"EventViewModelDetails.WithTag":        "Filter details by tag",
		"EventViewModelDetails.General":        "Get details with empty tag",
		"EventViewModelDetails.WithNames":      "Filter details by names",
		"EventViewModelDetails.Names":          "Get all detail names",
		"EventViewModelDetails.Values":         "Get all detail values",
		"EventViewModelDetails.ToMap":          "Convert to name->value map",
		"EventViewModelDetails.Has":            "Check if detail with name exists",
		"EventViewModelDetails.HasTag":         "Check if any detail has tag",
		"EventViewModelDetails.Get":            "Get detail by name",
		"EventViewModelDetails.GetValue":       "Get value by name",
		"EventViewModelDetails.PrettifiedMetrics": "Get metric details with formatted values",

		// EventViewModelDetail methods
		"EventViewModelDetail.LabelOrName": "Returns Label if set, otherwise Name",
	}

	if desc, ok := descriptions[key]; ok {
		return desc
	}
	return ""
}

// extractFunctions returns all template functions
func extractFunctions() []*SchemaFunction {
	functions := []*SchemaFunction{
		// String functions
		{Name: "toUpper", Signature: "(s string) string", Description: "Convert string to uppercase", Category: "string"},
		{Name: "title", Signature: "(s string) string", Description: "Convert string to title case", Category: "string"},
		{Name: "trimSpace", Signature: "(s string) string", Description: "Remove leading and trailing whitespace", Category: "string"},
		{Name: "split", Signature: "(s string, sep string) []string", Description: "Split string by separator", Category: "string"},

		// JSON functions
		{Name: "toJSON", Signature: "(v interface{}) string", Description: "Convert value to JSON string", Category: "conversion"},
		{Name: "j", Signature: "(v interface{}) string", Description: "Alias for toJSON", Category: "conversion"},
		{Name: "uglifyJSON", Signature: "(s string) string", Description: "Compact JSON string (remove whitespace)", Category: "conversion"},
		{Name: "explodeJSONKeys", Signature: "(s string) string", Description: "Extract object key-values without braces", Category: "conversion"},
		{Name: "x", Signature: "(s string) string", Description: "Alias for explodeJSONKeys", Category: "conversion"},

		// Time functions
		{Name: "timeRfc3339", Signature: "(v interface{}) string", Description: "Convert to RFC3339 format (accepts string, int, time.Time)", Category: "time"},

		// Utility functions
		{Name: "join", Signature: "(index int) string", Description: "Returns comma for index > 0, empty for 0 (for list joining)", Category: "utility"},
		{Name: "joinWith", Signature: "(index int, sep string) string", Description: "Returns separator for index > 0, empty for 0", Category: "utility"},

		// Importance/severity functions
		{Name: "importanceLabel", Signature: "(severity ViewModelImportance) string", Description: "Get title-case label for importance level", Category: "formatting"},
		{Name: "importanceToColor", Signature: "(severity ViewModelImportance) string", Description: "Get hex color for importance level", Category: "formatting"},
		{Name: "importanceToEmoji", Signature: "(severity ViewModelImportance) string", Description: "Get emoji(s) for importance level", Category: "formatting"},
	}

	// Sort by name for consistent output
	sort.Slice(functions, func(i, j int) bool {
		return functions[i].Name < functions[j].Name
	})

	return functions
}

// extractEnums returns all enum types
func extractEnums() map[string]*SchemaEnum {
	return map[string]*SchemaEnum{
		"ViewModelImportance": {
			Values:      []string{"None", "Healthy", "Notice", "Minor", "Warning", "Major", "Severe", "Critical"},
			Description: "Alert severity levels (0-7)",
		},
		"EventType": {
			Values:      []string{EventType_Alarm, EventType_Insight, EventType_CustomInsight, EventType_Synthetics, EventType_Mitigation, EventType_Generic},
			Description: "Event type classification",
		},
		"DetailTag": {
			Values:      []string{"", "metric", "dimension", "url", "device", "device_labels", "device_label"},
			Description: "Detail categorization tags",
		},
	}
}

// ImportanceValueMap provides int to string mapping for enum values
var ImportanceValueMap = map[ViewModelImportance]string{
	ViewModelImportance_None:     "None",
	ViewModelImportance_Healthy:  "Healthy",
	ViewModelImportance_Notice:   "Notice",
	ViewModelImportance_Minor:    "Minor",
	ViewModelImportance_Warning:  "Warning",
	ViewModelImportance_Major:    "Major",
	ViewModelImportance_Severe:   "Severe",
	ViewModelImportance_Critical: "Critical",
}
