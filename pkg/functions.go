package main

import (
	"encoding/json"
	"strings"
	"text/template"
)

var TextTemplateFuncMap = template.FuncMap{
	"toUpper":   strings.ToUpper,
	"title":     strings.Title,
	"trimSpace": strings.TrimSpace,

	"toJSON":          toJSON,
	"j":               toJSON,
	"uglifyJSON":      compactJSON,
	"explodeJSONKeys": explodeJSONKeys,
	"x":               explodeJSONKeys,

	"join":     join,
	"joinWith": joinWith,
}

func toJSON(v interface{}) string {
	bs, err := json.Marshal(v)
	if err != nil {
		return "null"
	}
	return string(bs)
}

func join(index int) string {
	if index == 0 {
		return ""
	}
	return ","
}

func joinWith(index int, join string) string {
	if index == 0 {
		return ""
	}
	return join
}

func compactJSON(s string) string {
	var v interface{}
	err := json.Unmarshal([]byte(s), &v)
	if err != nil {
		return "uglifyErr"
	}
	return toJSON(v)
}

func explodeJSONKeys(s string) string {
	// Input: a stringified JSON object.
	// Output: a not-quite-json string of just the object kvs, fit to be embedded in another object.
	compacted := compactJSON(s)
	if len(compacted) > 1 && compacted[0] == '{' && compacted[len(compacted)-1] == '}' {
		return compacted[1 : len(compacted)-1]
	}
	return "explodeJSONKeysErr"
}
