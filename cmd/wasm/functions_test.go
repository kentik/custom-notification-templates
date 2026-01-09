package main

import (
	"encoding/json"
	"os"
	"path"
	"strings"
	"testing"
)

func TestProcessRender(t *testing.T) {
	pathRoot := "../../"
	if _, err := os.Stat(path.Join(pathRoot, "templates/json-clean.json.tmpl")); os.IsNotExist(err) {
		t.Fatalf("templates/json-clean.json.tmpl not found, perhaps wrong directory (run from project root)? %v", err)
	}

	templateContent, err := os.ReadFile(path.Join(pathRoot, "templates/json-clean.json.tmpl"))
	if err != nil {
		t.Fatalf("Failed to read template: %v", err)
	}

	dataContent, err := os.ReadFile(path.Join(pathRoot, "pkg/render/fixtures/alarm.json"))
	if err != nil {
		t.Fatalf("Failed to read data: %v", err)
	}

	validTemplate := string(templateContent)
	validData := string(dataContent)

	tests := []struct {
		name           string
		template       string
		data           string
		wantErr        bool
		errContains    string
		validateResult func(*testing.T, string)
	}{
		{
			name:     "Valid template renders successfully",
			template: validTemplate,
			data:     validData,
			wantErr:  false,
			validateResult: func(t *testing.T, output string) {
				if len(output) == 0 {
					t.Error("Output should not be empty")
				}
				var jsonOut map[string]interface{}
				if err := json.Unmarshal([]byte(output), &jsonOut); err != nil {
					t.Errorf("Output should be valid JSON: %v", err)
				}
			},
		},
		{
			name:     "Simple template with variable substitution",
			template: "{{ .CompanyName }}",
			data:     `{"CompanyName": "Test Company", "CompanyID": 1}`,
			wantErr:  false,
			validateResult: func(t *testing.T, output string) {
				if output != "Test Company" {
					t.Errorf("Expected 'Test Company', got '%s'", output)
				}
			},
		},
		{
			name:        "Invalid template syntax returns error",
			template:    "{{ .Invalid",
			data:        validData,
			wantErr:     true,
			errContains: "unclosed action",
			validateResult: func(t *testing.T, fullResponse string) {
				// processRender returns a JSON string that MIGHT contain an error field if it's a wrapper error,
				// OR it returns the JSON of RenderResponse which has Error field.
				// Wait, processRender marshals RenderResponse.
				var resp map[string]interface{}
				json.Unmarshal([]byte(fullResponse), &resp)

				if _, ok := resp["error"]; !ok {
					t.Error("Should have error field")
				}

				// Check range fields
				if _, ok := resp["startLine"]; !ok {
					t.Error("Should have startLine field")
				}
			},
		},
		{
			name:     "Undefined field in strict mode",
			template: "{{ .NonExistentField.SubField }}",
			data:     validData,
			wantErr:  true,
			// The error message depends on text/template implementation
		},
		{
			name:        "Invalid JSON data",
			template:    "{{ .CompanyID }}",
			data:        "{invalid json}",
			wantErr:     true,
			errContains: "Data parse error",
		},
		{
			name:     "Empty template",
			template: "",
			data:     validData,
			wantErr:  false,
			validateResult: func(t *testing.T, output string) {
				if output != "" {
					t.Errorf("Expected empty string, got '%s'", output)
				}
			},
		},
		{
			name:     "Template functions (toUpper)",
			template: "{{ .CompanyName | toUpper }}",
			data:     `{"CompanyName": "Test", "CompanyID": 1}`,
			wantErr:  false,
			validateResult: func(t *testing.T, output string) {
				if output != "TEST" {
					t.Errorf("Expected 'TEST', got '%s'", output)
				}
			},
		},
		{
			name:     "toJSON function",
			template: "{{ toJSON . }}",
			data:     `{"CompanyName": "Test", "CompanyID": 123}`,
			wantErr:  false,
			validateResult: func(t *testing.T, output string) {
				var res map[string]interface{}
				if err := json.Unmarshal([]byte(output), &res); err != nil {
					t.Errorf("Failed to parse output JSON: %v", err)
				}
				if res["CompanyID"].(float64) != 123 {
					t.Errorf("Expected CompanyID 123, got %v", res["CompanyID"])
				}
			},
		},
		{
			name:     "importanceToColor function",
			template: "{{ importanceToColor .Event.Importance }}",
			data:     validData,
			validateResult: func(t *testing.T, output string) {
				if !strings.HasPrefix(output, "#") {
					t.Errorf("Expected hex color, got '%s'", output)
				}
				if len(output) != 7 {
					t.Errorf("Expected 7 chars hex color, got %d", len(output))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resJSON := processRender(tt.template, tt.data)

			// processRender returns a JSON string representing RenderResponse OR an error wrapper.
			// We need to parse it to check for error.
			var resp struct {
				Output    string `json:"output"`
				Error     string `json:"error"`
				Line      int    `json:"line"`
				StartLine int    `json:"startLine"`
			}
			err := json.Unmarshal([]byte(resJSON), &resp)
			if err != nil {
				t.Fatalf("Failed to unmarshal result: %v. Raw: %s", err, resJSON)
			}

			if tt.wantErr {
				if resp.Error == "" {
					t.Errorf("Expected error but got none")
				}
				if tt.errContains != "" && !strings.Contains(resp.Error, tt.errContains) {
					t.Errorf("Error '%s' does not contain '%s'", resp.Error, tt.errContains)
				}
				if tt.validateResult != nil {
					tt.validateResult(t, resJSON)
				}
			} else {
				if resp.Error != "" {
					t.Errorf("Unexpected error: %s", resp.Error)
				}
				if tt.validateResult != nil {
					tt.validateResult(t, resp.Output)
				}
			}
		})
	}
}

func TestProcessGetSchema(t *testing.T) {
	resJSON := processGetSchema()

	var schema struct {
		Fields []struct {
			Name     string        `json:"name"`
			Path     string        `json:"path"`
			Children []interface{} `json:"children"`
			IsMethod bool          `json:"isMethod"`
		} `json:"fields"`
		Functions []struct {
			Name        string `json:"name"`
			Signature   string `json:"signature"`
			Description string `json:"description"`
			Category    string `json:"category"`
		} `json:"functions"`
		Enums map[string]struct {
			Values []string `json:"values"`
		} `json:"enums"`
	}

	if err := json.Unmarshal([]byte(resJSON), &schema); err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	// Check required sections
	if schema.Fields == nil {
		t.Error("Schema should have fields")
	}
	if schema.Functions == nil {
		t.Error("Schema should have functions")
	}
	if schema.Enums == nil {
		t.Error("Schema should have enums")
	}

	// Check fields
	fieldNames := make(map[string]bool)
	for _, f := range schema.Fields {
		fieldNames[f.Name] = true
	}

	requiredFields := []string{"CompanyID", "CompanyName", "Config", "Event"}
	for _, req := range requiredFields {
		if !fieldNames[req] {
			t.Errorf("Schema missing field: %s", req)
		}
	}

	// Check Event field structure
	var eventField *struct {
		Name     string        `json:"name"`
		Path     string        `json:"path"`
		Children []interface{} `json:"children"`
		IsMethod bool          `json:"isMethod"`
	}
	for _, f := range schema.Fields {
		if f.Name == "Event" {
			// we need to take address or copy? range gives copy.
			// create local var
			val := f
			eventField = &val
			break
		}
	}

	if eventField == nil {
		t.Error("Event field not found")
	} else {
		if !eventField.IsMethod && eventField.Children == nil {
			t.Error("Event should be method or have children")
		}
	}

	// Check functions
	funcNames := make(map[string]bool)
	for _, f := range schema.Functions {
		funcNames[f.Name] = true
		if f.Signature == "" {
			t.Errorf("Function %s missing signature", f.Name)
		}
		if f.Description == "" {
			t.Errorf("Function %s missing description", f.Name)
		}
	}

	requiredFuncs := []string{"toJSON", "toUpper", "importanceToColor", "timeRfc3339"}
	for _, req := range requiredFuncs {
		if !funcNames[req] {
			t.Errorf("Schema missing function: %s", req)
		}
	}

	// Check enums
	if _, ok := schema.Enums["ViewModelImportance"]; !ok {
		t.Error("Missing ViewModelImportance enum")
	} else {
		vals := schema.Enums["ViewModelImportance"].Values
		valMap := make(map[string]bool)
		for _, v := range vals {
			valMap[v] = true
		}
		if !valMap["Critical"] || !valMap["Healthy"] {
			t.Error("ViewModelImportance missing expected values")
		}
	}

	if _, ok := schema.Enums["EventType"]; !ok {
		t.Error("Missing EventType enum")
	}
}
