package render

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"text/template"
	"time"

	"github.com/kentik/custom-notification-templates/pkg/types"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ToUpper converts a string to uppercase.
// Category: string
func ToUpper(s string) string {
	return strings.ToUpper(s)
}

// ToLower converts a string to lowercase.
// Category: string
func ToLower(s string) string {
	return strings.ToLower(s)
}

// Title converts a string to Title case.
// Category: string
func Title(s string) string {
	return cases.Title(language.English).String(s)
}

// TrimSpace removes leading and trailing whitespace.
// Category: string
func TrimSpace(s string) string {
	return strings.TrimSpace(s)
}

// Split splits a string by the given separator.
// Category: string
func Split(s string, sep string) []string {
	return strings.Split(s, sep)
}

var TextTemplateFuncMap = template.FuncMap{
	"toUpper":   ToUpper,
	"toLower":   ToLower,
	"title":     Title,
	"trimSpace": TrimSpace,
	"split":     Split,

	"toJSON":          ToJSON,
	"j":               ToJSON,
	"uglifyJSON":      CompactJSON,
	"explodeJSONKeys": ExplodeJSONKeys,
	"x":               ExplodeJSONKeys,
	"timeRfc3339":     TimeRfc3339,
	"join":            Join,
	"joinWith":        JoinWith,

	"importanceName":    ImportanceName,
	"importanceLabel":   ImportanceLabel,
	"importanceToColor": ImportanceToColor,
	"importanceToEmoji": ImportanceToEmoji,
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

// TimeRfc3339 converts various time formats to RFC3339.
// Accepts string, int (Unix timestamp), or time.Time.
// Category: time
func TimeRfc3339(input interface{}) string {
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

// ToJSON converts any value to a JSON string.
// Category: conversion
func ToJSON(v interface{}) string {
	bs, err := json.Marshal(v)
	if err != nil {
		return "null"
	}
	return string(bs)
}

// Join returns a comma for index > 0, empty string for index 0.
// Useful for joining list items in templates.
// Category: utility
func Join(index int) string {
	if index == 0 {
		return ""
	}
	return ","
}

// JoinWith returns the separator for index > 0, empty string for index 0.
// Category: utility
func JoinWith(index int, join string) string {
	if index == 0 {
		return ""
	}
	return join
}

// CompactJSON compacts a JSON string by removing whitespace.
// Category: conversion
func CompactJSON(s string) string {
	var v interface{}
	err := json.Unmarshal([]byte(s), &v)
	if err != nil {
		return "uglifyErr"
	}
	return ToJSON(v)
}

// ExplodeJSONKeys extracts object key-values without braces.
// Category: conversion
func ExplodeJSONKeys(s string) string {
	// Input: a stringified JSON object.
	// Output: a not-quite-json string of just the object kvs, fit to be embedded in another object.
	compacted := CompactJSON(s)
	if len(compacted) > 1 && compacted[0] == '{' && compacted[len(compacted)-1] == '}' {
		return compacted[1 : len(compacted)-1]
	}
	return "explodeJSONKeysErr"
}

// ImportanceToColor returns the hex color code for an importance level.
// Category: formatting
func ImportanceToColor(severity types.ViewModelImportance) string {
	if color, ok := types.ImportanceToColors[severity]; ok {
		return color
	}
	return ""
}

// ImportanceToColor returns the hex color code for an importance level.
// Category: formatting
func ImportanceName(severity types.ViewModelImportance) string {
	if label, ok := types.ImportanceNames[severity]; ok {
		return label
	}
	return ""
}

// ImportanceLabel returns the title-case label for an importance level.
// Category: formatting
func ImportanceLabel(severity types.ViewModelImportance) string {
	return cases.Title(language.English).String(ImportanceName(severity))
}

// ImportanceToEmoji returns the emoji(s) for an importance level.
// Category: formatting
func ImportanceToEmoji(severity types.ViewModelImportance) string {
	if emoji, ok := types.ImportanceToEmojis[severity]; ok {
		return emoji
	}
	return ""
}
