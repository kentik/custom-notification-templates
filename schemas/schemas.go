package schemas

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

//go:embed details.yaml
var detailsYaml []byte

type Detail struct {
	Name        string `yaml:"Name"`
	Tag         string `yaml:"Tag"`
	Description string `yaml:"Description"`
	When        string `yaml:"When"`
	Examples    []any  `yaml:"Examples"`
	Value       any    `yaml:"Value"`
}

func Details() []Detail {
	var result []Detail
	err := yaml.Unmarshal(detailsYaml, &result)
	if err != nil {
		panic(fmt.Sprintf("Error unmarshaling details schema yaml: %s", err))
	}
	return result
}

func (ds *Detail) ValueSchema() *gojsonschema.Schema {
	valueJson, err := json.Marshal(ds.Value)
	if err != nil {
		panic(fmt.Sprintf("Error marshaling detail %s schema: %s", ds.Name, err))
	}

	schema, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(string(valueJson)))
	if err != nil {
		panic(fmt.Sprintf("Error compiling JSON Schema for detail %s schema: %s", ds.Name, err))
	}
	return schema
}
