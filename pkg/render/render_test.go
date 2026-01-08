package render

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"
)

type TemplateEntry struct {
	Name   string
	Path   string
	IsJson bool
}

func templateFiles(location string) ([]TemplateEntry, error) {
	var result []TemplateEntry

	entries, err := os.ReadDir(location)
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
	entries, err := templateFiles("../../templates")
	if err != nil {
		t.Fatalf("Error reading directory: %s", err)
	}
	t.Logf("Found %d template files", len(entries))
	for _, entry := range entries {
		t.Logf("Testing template file %s", entry.Path)
		templateContent, err := os.ReadFile(entry.Path)
		if err != nil {
			t.Fatalf("Error reading template file %s: %s", entry.Name, err)
		}

		for modelName, model := range TestingViewModels {
			req := RenderRequest{
				Template: string(templateContent),
				Data:     model,
			}
			resp := Render(req)
			if resp.Error != "" {
				t.Fatalf("Error rendering %s using %s: %s", modelName, entry.Name, resp.Error)
			}

			result := resp.Output
			outputPath := fmt.Sprintf("../../output/%s-%s", modelName, strings.TrimSuffix(entry.Name, ".tmpl"))
			os.WriteFile(outputPath, []byte(result), 0644)
			t.Logf("Wrote output to %s", outputPath)

			if entry.IsJson {
				var jsonValue interface{}
				err := json.Unmarshal([]byte(result), &jsonValue)
				if err != nil {
					t.Fatalf("JSON error when rendering %s using %s: %s. Payload: %s", modelName, entry.Name, err, string(result))
				}
			}
		}
	}

	t.Logf("Rendered all successfully using %d view models and %d template files", len(TestingViewModels), len(entries))
}
