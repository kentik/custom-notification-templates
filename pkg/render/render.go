package render

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/kentik/custom-notification-templates/pkg/types"
)

var lineColumnRegex = regexp.MustCompile(`template:\s*[^:]+:(\d+)(?::(\d+))?`)

type RenderRequest struct {
	Name     string          `json:"name"`
	Template string          `json:"template"`
	Data     json.RawMessage `json:"data"`
}

type RenderResponse struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`

	// simple fields
	Line   int  `json:"line,omitempty"`
	Column *int `json:"column,omitempty"`

	// more sophisticated error handling
	StartLine   int  `json:"startLine,omitempty"`
	StartColumn *int `json:"startColumn,omitempty"`
	EndLine     int  `json:"endLine,omitempty"`
	EndColumn   *int `json:"endColumn,omitempty"`
}

func Render(req RenderRequest) RenderResponse {
	name := req.Name
	if name == "" {
		name = "template"
	}
	tmpl, err := template.New(name).
		Funcs(TextTemplateFuncMap).
		Parse(req.Template)

	if err != nil {
		return renderErr(err)
	}

	ctx, _, parseErr := buildContext(req.Data)
	if parseErr != nil {
		return RenderResponse{Error: "Data parse error: " + parseErr.Error()}
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return renderErr(err)
	}

	return RenderResponse{Output: buf.String()}
}

func buildContext(data json.RawMessage) (*types.NotificationViewModel, bool, error) {
	var res types.NotificationViewModel

	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()

	if err := dec.Decode(&res); err != nil {
		return nil, false, err
	}

	if res.Now.IsZero() {
		res.Now = time.Now().UTC()
	}
	return &res, true, nil
}

func renderErr(err error) RenderResponse {
	errMsg := err.Error()
	line, col := extractLineColumn(errMsg)

	var colPtr *int
	if col > 0 {
		colPtr = &col
	}

	resp := RenderResponse{
		Error:  errMsg,
		Line:   line,
		Column: colPtr,
		// range fields - start position is known, end position defaults to same line
		StartLine:   line,
		StartColumn: colPtr,
		EndLine:     line,
		EndColumn:   colPtr,
	}

	return resp
}

func extractLineColumn(errMsg string) (int, int) {
	//  - "template: subject:4: ..."
	//  - "template: body:7:3: ..."
	//  - "template: template:7:3: ..."
	matches := lineColumnRegex.FindStringSubmatch(errMsg)

	if len(matches) > 1 {
		line, _ := strconv.Atoi(matches[1])
		col := 0

		if len(matches) > 2 && matches[2] != "" {
			col, _ = strconv.Atoi(matches[2])
		}

		return line, col
	}

	return 0, 0
}
