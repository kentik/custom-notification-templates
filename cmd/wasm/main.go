//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"
	"time"

	"github.com/kentik/custom-notification-templates/pkg/render"
)

func renderTemplate(this js.Value, args []js.Value) (result any) {
	resultErrWrapper := func(err error) {
		errMap := map[string]string{"error": err.Error()}
		b, errJ := json.Marshal(errMap)
		if errJ != nil {
			result = `{"error": "Unexpected error: failed to marshal error response"}`
			return
		}
		result = string(b)
	}

	defer func() {
		if r := recover(); r != nil {
			resultErrWrapper(fmt.Errorf("Unexpected error: %v", r))
		}
	}()

	if len(args) < 2 {
		resultErrWrapper(fmt.Errorf("Expected arguments: (template: string, dataJson: string)"))
		return
	}

	if args[0].Type() != js.TypeString || args[1].Type() != js.TypeString {
		resultErrWrapper(fmt.Errorf("Arguments must be strings"))
		return
	}

	templateText := args[0].String()
	dataJSON := args[1].String()

	req := render.RenderRequest{
		Template: templateText,
		Data:     json.RawMessage(dataJSON),
	}

	// Channel to receive the result or error/panic from the render goroutine
	type renderResult struct {
		resp render.RenderResponse
		err  error
	}
	ch := make(chan renderResult, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				ch <- renderResult{err: fmt.Errorf("panic in render: %v", r)}
			}
		}()
		ch <- renderResult{resp: render.Render(req)}
	}()

	select {
	case res := <-ch:
		if res.err != nil {
			resultErrWrapper(res.err)
			return
		}

		b, err := json.Marshal(res.resp)
		if err != nil {
			resultErrWrapper(fmt.Errorf("Unexpected error: failed to marshal response: %v", err))
			return
		}
		return string(b)

	case <-time.After(5 * time.Second):
		resultErrWrapper(fmt.Errorf("render timed out after 5s"))
		return
	}
}

func main() {
	js.Global().Set("goTemplateRender", js.FuncOf(renderTemplate))
	select {}
}
