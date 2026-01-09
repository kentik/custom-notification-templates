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

	// Generate code
	code := generateCode(*pkg, methods, functions)

	// Write output
	outFile := filepath.Join(*dir, *output)
	if err := os.WriteFile(outFile, []byte(code), 0644); err != nil {
		log.Fatalf("Write error: %v", err)
	}

	fmt.Printf("✓ Generated %s (%d methods, %d functions)\n", outFile, len(methods), len(functions))
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
// 3. Add "Category: <name>" marker in the doc comment
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
func generateCode(pkgName string, methods []MethodInfo, functions []FunctionInfo) string {
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
