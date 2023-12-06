package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"text/template"
	"time"
)

var TextTemplateFuncMap = template.FuncMap{
	"toUpper":   strings.ToUpper,
	"title":     strings.Title,
	"trimSpace": strings.TrimSpace,
	"split":     strings.Split,

	"toJSON":          toJSON,
	"j":               toJSON,
	"uglifyJSON":      compactJSON,
	"explodeJSONKeys": explodeJSONKeys,
	"x":               explodeJSONKeys,
	"timeRfc3339":     timeRfc3339,
	"join":            join,
	"joinWith":        joinWith,
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
