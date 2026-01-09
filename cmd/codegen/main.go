//go:build ignore

package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	var (
		pkg    = flag.String("pkg", "", "Target package name")
		dir    = flag.String("dir", ".", "Directory to scan")
		output = flag.String("output", "", "Output file name")
	)
	flag.Parse()

	if *pkg == "" || *output == "" {
		fmt.Fprintln(os.Stderr, "Usage: codegen -pkg <name> -dir <dir> -output <file>")
		os.Exit(1)
	}

	// Parse Go files
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, *dir, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("Parse error: %v", err)
	}

	// Find target package
	targetPkg, exists := pkgs[*pkg]
	if !exists {
		log.Fatalf("Package '%s' not found in %s", *pkg, *dir)
	}

	// Extract metadata
	methods := extractMethods(targetPkg)
	functions := extractFunctions(targetPkg)
	enums := extractEnums(targetPkg)

	// Generate code
	code := generateCode(*pkg, methods, functions, enums)

	// Write output
	outFile := filepath.Join(*dir, *output)
	if err := os.WriteFile(outFile, []byte(code), 0644); err != nil {
		log.Fatalf("Write error: %v", err)
	}

	fmt.Printf("✓ Generated %s (%d methods, %d functions, %d enums)\n", outFile, len(methods), len(functions), len(enums))
}

// MethodInfo holds method metadata
type MethodInfo struct {
	TypeName   string
	MethodName string
	Doc        string
}

// FunctionInfo holds function metadata
type FunctionInfo struct {
	Name        string
	Signature   string
	Description string
	Category    string
}

// EnumInfo holds enum metadata
type EnumInfo struct {
	Name        string
	Values      []string
	Description string
}

// extractMethods extracts method doc comments from types
func extractMethods(pkg *ast.Package) []MethodInfo {
	var methods []MethodInfo

	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || !ast.IsExported(fn.Name.Name) {
				continue
			}

			// Extract receiver type name
			var typeName string
			if len(fn.Recv.List) > 0 {
				recvType := fn.Recv.List[0].Type
				// Handle pointer receivers
				if star, ok := recvType.(*ast.StarExpr); ok {
					if ident, ok := star.X.(*ast.Ident); ok {
						typeName = ident.Name
					}
				} else if ident, ok := recvType.(*ast.Ident); ok {
					typeName = ident.Name
				}
			}

			if typeName == "" {
				continue
			}

			// Extract doc comment (first sentence or full if short)
			doc := ""
			if fn.Doc != nil {
				doc = extractFirstSentence(fn.Doc.Text())
			}

			methods = append(methods, MethodInfo{
				TypeName:   typeName,
				MethodName: fn.Name.Name,
				Doc:        doc,
			})
		}
	}

	// Sort for deterministic output
	sort.Slice(methods, func(i, j int) bool {
		if methods[i].TypeName == methods[j].TypeName {
			return methods[i].MethodName < methods[j].MethodName
		}
		return methods[i].TypeName < methods[j].TypeName
	})

	return methods
}

// extractFunctions extracts function metadata from template helper functions in functions.go
//
// IMPORTANT: This function ONLY parses functions.go - functions defined in other files
// will NOT be included in the generated metadata, even if they have doc comments and
// category markers. This is intentional to keep template helpers in a single location.
//
// To add a new template function:
// 1. Define it in pkg/render/functions.go (not types.go or other files)
// 2. Add a doc comment with description
// 3. Add "Category: <name>" marker in the doc comment -- this is REQUIRED, warning below
// 4. Run 'go generate ./pkg/render' to regenerate metadata
func extractFunctions(pkg *ast.Package) []FunctionInfo {
	var functions []FunctionInfo
	var skippedFuncs []string

	// Only parse functions.go
	for fileName, file := range pkg.Files {
		if !strings.HasSuffix(fileName, "functions.go") {
			continue
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			// Include both exported and unexported functions (template helpers are lowercase)
			// Skip methods (those with receivers)
			// Skip test functions and other helper functions
			if !ok || fn.Recv != nil {
				continue
			}

			funcName := fn.Name.Name

			// Skip test functions, tryParseTime, importanceName
			if strings.HasPrefix(funcName, "Test") ||
			   strings.HasPrefix(funcName, "Benchmark") ||
			   funcName == "tryParseTime" ||
			   funcName == "importanceName" {
				continue
			}

			// Extract doc comment
			doc := ""
			category := ""
			if fn.Doc != nil {
				fullDoc := fn.Doc.Text()
				doc = extractFirstSentence(fullDoc)
				category = extractCategory(fullDoc)
			}

			// Only include functions that have a category (template helpers)
			if category == "" {
				skippedFuncs = append(skippedFuncs, funcName)
				continue
			}

			// Extract signature
			signature := extractSignature(fn)

			functions = append(functions, FunctionInfo{
				Name:        funcName,
				Signature:   signature,
				Description: doc,
				Category:    category,
			})
		}
	}

	// Warn about functions without category markers
	if len(skippedFuncs) > 0 {
		fmt.Fprintf(os.Stderr, "⚠️  Warning: %d function(s) in functions.go skipped (missing 'Category:' marker):\n", len(skippedFuncs))
		for _, name := range skippedFuncs {
			fmt.Fprintf(os.Stderr, "   - %s\n", name)
		}
		fmt.Fprintf(os.Stderr, "   Add '// Category: <name>' to include them in metadata.\n")
	}

	// Sort for deterministic output
	sort.Slice(functions, func(i, j int) bool {
		return functions[i].Name < functions[j].Name
	})

	return functions
}

// extractEnums extracts enum type definitions with their const values.
//
// Detects two patterns:
// 1. Type definition followed by const block (ViewModelImportance, DetailTag)
// 2. Untyped const blocks with common prefix (EventType)
//
// For string-based enums, extracts actual string values from consts.
// For non-string enums, strips prefix from const names for display.
func extractEnums(pkg *ast.Package) []EnumInfo {
	var enums []EnumInfo

	for _, file := range pkg.Files {
		// First pass: find type definitions with associated const blocks
		for i, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			// Process each type spec in the declaration
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				typeName := typeSpec.Name.Name

				// Look for const block within next 5 declarations
				constBlock := findConstBlockAfter(file.Decls, i, typeName)
				if constBlock == nil {
					continue
				}

				// Determine underlying type for value extraction strategy
				underlyingType := getUnderlyingTypeName(typeSpec.Type)

				// Extract enum values
				values := extractEnumValuesFromConst(constBlock, typeName, underlyingType)
				if len(values) == 0 {
					continue
				}

				// Extract description from type doc comment
				desc := ""
				if genDecl.Doc != nil {
					desc = extractFirstSentence(genDecl.Doc.Text())
				}

				enums = append(enums, EnumInfo{
					Name:        typeName,
					Values:      values,
					Description: desc,
				})
			}
		}

		// Second pass: find untyped const groups (like EventType)
		enums = append(enums, extractUntypedEnums(file)...)
	}

	// Sort for deterministic output
	sort.Slice(enums, func(i, j int) bool {
		return enums[i].Name < enums[j].Name
	})

	return enums
}

// findConstBlockAfter looks for a const block matching the type prefix within the next few declarations
func findConstBlockAfter(decls []ast.Decl, startIdx int, prefix string) *ast.GenDecl {
	for i := startIdx + 1; i < len(decls) && i < startIdx+6; i++ {
		genDecl, ok := decls[i].(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}

		// Check if first const matches pattern TypeName_*
		if len(genDecl.Specs) > 0 {
			if valueSpec, ok := genDecl.Specs[0].(*ast.ValueSpec); ok {
				if len(valueSpec.Names) > 0 {
					name := valueSpec.Names[0].Name
					if strings.HasPrefix(name, prefix+"_") {
						return genDecl
					}
				}
			}
		}
	}
	return nil
}

// getUnderlyingTypeName returns the name of the underlying type
func getUnderlyingTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	default:
		return ""
	}
}

// extractEnumValuesFromConst extracts enum values from a const block
// For string types, extracts actual string values
// For other types, extracts const names (stripped of prefix)
func extractEnumValuesFromConst(constBlock *ast.GenDecl, prefix string, underlyingType string) []string {
	var values []string

	for _, spec := range constBlock.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}

		for i, name := range valueSpec.Names {
			constName := name.Name

			if !strings.HasPrefix(constName, prefix+"_") {
				continue
			}

			// For string types, extract the actual value
			if underlyingType == "string" {
				if len(valueSpec.Values) > i {
					if basicLit, ok := valueSpec.Values[i].(*ast.BasicLit); ok {
						if basicLit.Kind == token.STRING {
							// Strip quotes from string literal
							strVal := strings.Trim(basicLit.Value, `"`)
							values = append(values, strVal)
							continue
						}
					}
				}
			}

			// For non-string types, use the const name (stripped of prefix)
			displayName := strings.TrimPrefix(constName, prefix+"_")
			values = append(values, displayName)
		}
	}

	return values
}

// extractUntypedEnums finds untyped const groups that look like enums (e.g., EventType)
func extractUntypedEnums(file *ast.File) []EnumInfo {
	var enums []EnumInfo

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}

		// Group constants by prefix
		prefixGroups := make(map[string][]*ast.ValueSpec)

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok || len(valueSpec.Names) == 0 {
				continue
			}

			constName := valueSpec.Names[0].Name

			// Check if this is a typed const with explicit string type
			hasStringType := false
			if valueSpec.Type != nil {
				if ident, ok := valueSpec.Type.(*ast.Ident); ok {
					if ident.Name == "string" {
						hasStringType = true
					}
				}
			}

			// Extract prefix (e.g., "EventType" from "EventType_Alarm")
			if idx := strings.Index(constName, "_"); idx > 0 && hasStringType {
				prefix := constName[:idx]
				prefixGroups[prefix] = append(prefixGroups[prefix], valueSpec)
			}
		}

		// Process each prefix group
		for prefix, specs := range prefixGroups {
			if len(specs) < 2 {
				continue // Not an enum if only 1 value
			}

			// Extract string values from constants
			var values []string
			for _, spec := range specs {
				if len(spec.Values) > 0 {
					if basicLit, ok := spec.Values[0].(*ast.BasicLit); ok {
						if basicLit.Kind == token.STRING {
							strVal := strings.Trim(basicLit.Value, `"`)
							values = append(values, strVal)
						}
					}
				}
			}

			if len(values) == 0 {
				continue
			}

			enums = append(enums, EnumInfo{
				Name:        prefix,
				Values:      values,
				Description: "", // No type definition, so no doc comment
			})
		}
	}

	return enums
}

// extractFirstSentence extracts first sentence from doc comment
func extractFirstSentence(doc string) string {
	doc = strings.TrimSpace(doc)
	if doc == "" {
		return ""
	}

	// Find first period followed by space or newline
	for i := 0; i < len(doc); i++ {
		if doc[i] == '.' && (i+1 >= len(doc) || doc[i+1] == ' ' || doc[i+1] == '\n') {
			return strings.TrimSpace(doc[:i+1])
		}
	}

	// No period found, return first line
	lines := strings.Split(doc, "\n")
	return strings.TrimSpace(lines[0])
}

// extractCategory extracts "Category: <name>" from doc comment
func extractCategory(doc string) string {
	lines := strings.Split(doc, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Category:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Category:"))
		}
	}
	return ""
}

// extractSignature builds function signature string from AST
func extractSignature(fn *ast.FuncDecl) string {
	var params []string
	var results []string

	// Extract parameters
	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			typeStr := typeToString(field.Type)
			if len(field.Names) > 0 {
				for _, name := range field.Names {
					params = append(params, name.Name+" "+typeStr)
				}
			} else {
				params = append(params, typeStr)
			}
		}
	}

	// Extract return types
	if fn.Type.Results != nil {
		for _, field := range fn.Type.Results.List {
			results = append(results, typeToString(field.Type))
		}
	}

	// Build signature
	paramStr := strings.Join(params, ", ")
	resultStr := strings.Join(results, ", ")

	if len(results) > 1 {
		resultStr = "(" + resultStr + ")"
	}

	if resultStr == "" {
		return "(" + paramStr + ")"
	}
	return "(" + paramStr + ") " + resultStr
}

// typeToString converts AST type expression to string
func typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeToString(t.X)
	case *ast.ArrayType:
		return "[]" + typeToString(t.Elt)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.SelectorExpr:
		return typeToString(t.X) + "." + t.Sel.Name
	case *ast.MapType:
		return "map[" + typeToString(t.Key) + "]" + typeToString(t.Value)
	case *ast.Ellipsis:
		return "..." + typeToString(t.Elt)
	default:
		return "unknown"
	}
}

// generateCode generates Go source code with metadata
func generateCode(pkgName string, methods []MethodInfo, functions []FunctionInfo, enums []EnumInfo) string {
	var sb strings.Builder

	// Header
	sb.WriteString("// Code generated by go generate; DO NOT EDIT.\n")
	sb.WriteString("// Generated from doc comments in types.go and functions.go\n\n")
	sb.WriteString("package " + pkgName + "\n\n")

	// Method descriptions map
	sb.WriteString("// methodDescriptions maps TypeName.MethodName to their documentation.\n")
	sb.WriteString("// This is auto-generated from doc comments on methods in types.go.\n")
	sb.WriteString("var methodDescriptions = map[string]string{\n")
	for _, m := range methods {
		key := m.TypeName + "." + m.MethodName
		doc := escapeGoString(m.Doc)
		sb.WriteString(fmt.Sprintf("\t%q: %q,\n", key, doc))
	}
	sb.WriteString("}\n\n")

	// Function metadata slice
	sb.WriteString("// functionMetadata contains metadata for all template functions.\n")
	sb.WriteString("// This is auto-generated from doc comments in functions.go.\n")
	sb.WriteString("var functionMetadata = []*SchemaFunction{\n")
	for _, f := range functions {
		sb.WriteString("\t{\n")
		sb.WriteString(fmt.Sprintf("\t\tName:        %q,\n", f.Name))
		sb.WriteString(fmt.Sprintf("\t\tSignature:   %q,\n", f.Signature))
		sb.WriteString(fmt.Sprintf("\t\tDescription: %q,\n", escapeGoString(f.Description)))
		sb.WriteString(fmt.Sprintf("\t\tCategory:    %q,\n", f.Category))
		sb.WriteString("\t},\n")
	}
	sb.WriteString("}\n\n")

	// Enum definitions map
	sb.WriteString("// enumDefinitions contains all enum types.\n")
	sb.WriteString("// This is auto-generated from type definitions and const blocks in types.go.\n")
	sb.WriteString("var enumDefinitions = map[string]*SchemaEnum{\n")
	for _, e := range enums {
		sb.WriteString(fmt.Sprintf("\t%q: {\n", e.Name))

		// Values array
		sb.WriteString("\t\tValues: []string{")
		for i, v := range e.Values {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%q", v))
		}
		sb.WriteString("},\n")

		// Description
		if e.Description != "" {
			sb.WriteString(fmt.Sprintf("\t\tDescription: %q,\n", escapeGoString(e.Description)))
		}

		sb.WriteString("\t},\n")
	}
	sb.WriteString("}\n\n")

	// Helper functions
	sb.WriteString("// getMethodDescription looks up a method's documentation by type and method name.\n")
	sb.WriteString("// Returns empty string if not found.\n")
	sb.WriteString("func getMethodDescription(typeName, methodName string) string {\n")
	sb.WriteString("\tkey := typeName + \".\" + methodName\n")
	sb.WriteString("\tif desc, ok := methodDescriptions[key]; ok {\n")
	sb.WriteString("\t\treturn desc\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn \"\"\n")
	sb.WriteString("}\n\n")

	sb.WriteString("// extractFunctions returns all template function metadata.\n")
	sb.WriteString("// The slice is sorted by function name for consistency.\n")
	sb.WriteString("func extractFunctions() []*SchemaFunction {\n")
	sb.WriteString("\treturn functionMetadata\n")
	sb.WriteString("}\n\n")

	sb.WriteString("// extractEnums returns all enum definitions.\n")
	sb.WriteString("// Enums are auto-generated from type definitions and const blocks.\n")
	sb.WriteString("func extractEnums() map[string]*SchemaEnum {\n")
	sb.WriteString("\treturn enumDefinitions\n")
	sb.WriteString("}\n")

	return sb.String()
}

// escapeGoString escapes a string for Go source code
func escapeGoString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	return s
}
