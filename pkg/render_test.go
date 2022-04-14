package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"testing"
	"text/template"
)

type TemplateEntry struct {
	Name   string
	Path   string
	IsJson bool
}

func templateFiles(location string) ([]TemplateEntry, error) {
	var result []TemplateEntry

	entries, err := ioutil.ReadDir(location)
	if err != nil {
		return nil, fmt.Errorf("Error reading directory: %s", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tmpl") {
			result = append(result, TemplateEntry{
				Name:   entry.Name(),
				Path:   path.Join(location, entry.Name()),
				IsJson: strings.HasSuffix(entry.Name(), ".json.tmpl"),
			})
		}
	}
	return result, nil
}

func Test_AllExamples_Render(t *testing.T) {
	entries, err := templateFiles("../templates")
	if err != nil {
		t.Fatalf("Error reading directory: %s", err)
	}
	for _, entry := range entries {
		templateContent, err := ioutil.ReadFile(entry.Path)
		if err != nil {
			t.Fatalf("Error reading template file %s: %s", entry.Name, err)
		}
		parsed, err := template.New(entry.Name).Funcs(TextTemplateFuncMap).Parse(string(templateContent))
		if err != nil {
			t.Fatalf("Error parsing the template from %s: %s", entry.Name, err)
		}

		for modelName, model := range TestingViewModels {
			var buf bytes.Buffer
			err := parsed.Execute(&buf, model)
			if err != nil {
				t.Fatalf("Error rendering %s using %s: %s", modelName, entry.Name, err)
			}

			result := buf.Bytes()
			outputPath := fmt.Sprintf("../output/%s-%s", modelName, strings.TrimSuffix(entry.Name, ".tmpl"))
			ioutil.WriteFile(outputPath, result, 0644)

			if entry.IsJson {
				var jsonValue interface{}
				err := json.Unmarshal(result, &jsonValue)
				if err != nil {
					t.Fatalf("JSON error when rendering %s using %s: %s. Payload: %s", modelName, entry.Name, err, string(result))
				}
			}
		}
	}

	t.Logf("Rendered all successfully using %d view models and %d template files", len(TestingViewModels), len(entries))
}
