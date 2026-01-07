package schemas

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

func jsonPrettyStringify(value any) string {
	valueJson, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		panic(err)
	}
	return "<pre>" + strings.ReplaceAll(string(valueJson), "\n", "<br>") + "</pre>"
}

func htmlList(values []any) string {
	if len(values) == 0 {
		return ""
	}
	items := make([]string, 0, len(values))
	for _, value := range values {
		items = append(items, jsonPrettyStringify(value))
	}
	return "<ul><li>" + strings.Join(items, "</li><li>") + "</li></ul>"
}

func IntoMarkdown(details []Detail) string {
	builder := strings.Builder{}
	builder.WriteString("| Name | Tag | When present | Description  | Value schema | Example values |\n")
	builder.WriteString("| --- | --- | --- | --- | --- | --- |\n")

	for _, detail := range details {
		name := detail.Name
		if name == "" {
			name = "_(any)_"
		}
		tag := detail.Tag
		if tag == "" {
			tag = "_(empty)_"
		}

		fmt.Fprintf(&builder, "| %s | %s | %s | %s | %s | %s |\n",
			name,
			tag,
			detail.When,
			detail.Description,
			jsonPrettyStringify(detail.Value),
			htmlList(detail.Examples),
		)
	}
	return builder.String()
}
