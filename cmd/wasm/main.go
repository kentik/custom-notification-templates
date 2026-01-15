//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"
)

func renderTemplate(this js.Value, args []js.Value) (result any) {
	// JS-specific validation
	if len(args) < 2 {
		return resultErrWrapper(fmt.Errorf("Expected arguments: (template: string, dataJson: string)"))
	}

	if args[0].Type() != js.TypeString || args[1].Type() != js.TypeString {
		return resultErrWrapper(fmt.Errorf("Arguments must be strings"))
	}

	return processRender(args[0].String(), args[1].String())
}

// returns the complete schema for template editing, with available fields
func getSchema(this js.Value, args []js.Value) (result any) {
	return processGetSchema()
}

func main() {
	js.Global().Set("goTemplateRender", js.FuncOf(renderTemplate))
	js.Global().Set("goTemplateGetSchema", js.FuncOf(getSchema))
	select {}
}
