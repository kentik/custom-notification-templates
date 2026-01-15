package render

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"text/template"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// toUpper converts a string to uppercase.
// Category: string
func toUpper(s string) string {
	return strings.ToUpper(s)
}

// title converts a string to title case.
// Category: string
func title(s string) string {
	return cases.Title(language.English).String(s)
}

// trimSpace removes leading and trailing whitespace.
// Category: string
func trimSpace(s string) string {
	return strings.TrimSpace(s)
}

// split splits a string by the given separator.
// Category: string
func split(s string, sep string) []string {
	return strings.Split(s, sep)
}

var TextTemplateFuncMap = template.FuncMap{
	"toUpper":   toUpper,
	"title":     title,
	"trimSpace": trimSpace,
	"split":     split,

	"toJSON":          toJSON,
	"j":               toJSON,
	"uglifyJSON":      compactJSON,
	"explodeJSONKeys": explodeJSONKeys,
	"x":               explodeJSONKeys,
	"timeRfc3339":     timeRfc3339,
	"join":            join,
	"joinWith":        joinWith,

	"importanceLabel":   importanceLabel,
	"importanceToColor": importanceToColor,
	"importanceToEmoji": importanceToEmoji,
}

func tryParseTime(input string) (time.Time, error) {
	formats := []string{
		"2006-01-02 15:04:05 MST",
		"2006-01-02 15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
		time.RFC822,
		time.RFC822Z,
		time.ANSIC,
		time.UnixDate,
	}

	var parsed time.Time
	var err error
	for _, format := range formats {
		parsed, err = time.Parse(format, input)
		if err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", input)
}

// timeRfc3339 converts various time formats to RFC3339.
// Accepts string, int (Unix timestamp), or time.Time.
// Category: time
func timeRfc3339(input interface{}) string {
	var t time.Time
	var err error

	switch v := input.(type) {
	case string:
		t, err = tryParseTime(v)
		if err != nil {
			return v
		}
	case int, int8, int16, int32, int64:
		t = time.Unix(int64(reflect.ValueOf(v).Int()), 0)
	case uint, uint8, uint16, uint32, uint64:
		t = time.Unix(int64(reflect.ValueOf(v).Uint()), 0)
	case time.Time:
		t = v
	default:
		return fmt.Sprintf("(%t:%+v)", input, input)
	}

	return t.Format(time.RFC3339)
}

// toJSON converts any value to a JSON string.
// Category: conversion
func toJSON(v interface{}) string {
	bs, err := json.Marshal(v)
	if err != nil {
		return "null"
	}
	return string(bs)
}

// join returns a comma for index > 0, empty string for index 0.
// Useful for joining list items in templates.
// Category: utility
func join(index int) string {
	if index == 0 {
		return ""
	}
	return ","
}

// joinWith returns the separator for index > 0, empty string for index 0.
// Category: utility
func joinWith(index int, join string) string {
	if index == 0 {
		return ""
	}
	return join
}

// compactJSON compacts a JSON string by removing whitespace.
// Category: conversion
func compactJSON(s string) string {
	var v interface{}
	err := json.Unmarshal([]byte(s), &v)
	if err != nil {
		return "uglifyErr"
	}
	return toJSON(v)
}

// explodeJSONKeys extracts object key-values without braces.
// Category: conversion
func explodeJSONKeys(s string) string {
	// Input: a stringified JSON object.
	// Output: a not-quite-json string of just the object kvs, fit to be embedded in another object.
	compacted := compactJSON(s)
	if len(compacted) > 1 && compacted[0] == '{' && compacted[len(compacted)-1] == '}' {
		return compacted[1 : len(compacted)-1]
	}
	return "explodeJSONKeysErr"
}

// importanceToColor returns the hex color code for an importance level.
// Category: formatting
func importanceToColor(severity ViewModelImportance) string {
	if color, ok := ImportanceToColors[severity]; ok {
		return color
	}
	return ""
}

func importanceName(severity ViewModelImportance) string {
	if label, ok := ImportanceNames[severity]; ok {
		return label
	}
	return ""
}

// importanceLabel returns the title-case label for an importance level.
// Category: formatting
func importanceLabel(severity ViewModelImportance) string {
	return cases.Title(language.English).String(importanceName(severity))
}

// importanceToEmoji returns the emoji(s) for an importance level.
// Category: formatting
func importanceToEmoji(severity ViewModelImportance) string {
	if emoji, ok := ImportanceToEmojis[severity]; ok {
		return emoji
	}
	return ""
}
