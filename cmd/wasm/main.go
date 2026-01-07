//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/kentik/custom-notification-templates/pkg/render"
)

func renderTemplate(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		resp := render.RenderResponse{Error: "Expected (template: string, dataJson: string)"}
		b, _ := json.Marshal(resp)
		return string(b)
	}

	templateText := args[0].String()
	dataJSON := args[1].String()

	req := render.RenderRequest{
		Template: templateText,
		Data:     json.RawMessage(dataJSON),
	}

	resp := render.Render(req)
	b, _ := json.Marshal(resp)

	return string(b)
}

func main() {
	js.Global().Set("goTemplateRender", js.FuncOf(renderTemplate))
	select {}
}
