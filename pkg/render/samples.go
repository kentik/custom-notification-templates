package render

import (
	_ "embed"
	"encoding/json"
)

//go:embed fixtures/insight.json
var insight []byte

//go:embed fixtures/alarm.json
var alarm []byte

//go:embed fixtures/synthetics.json
var synthetics []byte

//go:embed fixtures/mitigation.json
var mitigation []byte

//go:embed fixtures/digest.json
var digest []byte

var TestingViewModels = map[string]json.RawMessage{
	"insight":    insight,
	"alarm":      alarm,
	"synthetics": synthetics,
	"mitigation": mitigation,
	"digest":     digest,
}
