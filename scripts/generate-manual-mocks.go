package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	baseDir := "."
	if len(os.Args) > 1 {
		baseDir = os.Args[1]
	}

	outputDir := "./test/mocks"
	if len(os.Args) > 2 {
		outputDir = os.Args[2]
	}

	fmt.Println("Generating manual mocks...")

	// Find all interfaces in pkg/domain and pkg/infra
	domainInterfaces := findInterfaces(filepath.Join(baseDir, "pkg/domain"))
	infraInterfaces := findInterfaces(filepath.Join(baseDir, "pkg/infra"))

	allInterfaces := append(domainInterfaces, infraInterfaces...)

	fmt.Printf("Found %d interfaces to mock\n", len(allInterfaces))

	successCount := 0
	failCount := 0

	for _, iface := range allInterfaces {
		if err := generateManualMock(iface, outputDir, baseDir); err != nil {
			fmt.Printf("Failed to generate mock for %s: %v\n", iface.Name, err)
			failCount++
		} else {
			successCount++
			if successCount%10 == 0 {
				fmt.Printf("   Generated %d mocks...\n", successCount)
			}
		}
	}

	fmt.Printf("\nManual mocks generation complete:\n")
	fmt.Printf("   - Successfully generated: %d\n", successCount)
	if failCount > 0 {
		fmt.Printf("   - Failed: %d\n", failCount)
	}
	fmt.Printf("   - Output directory: %s\n", outputDir)
}

type InterfaceInfo struct {
	Name        string
	PackageName string
	FilePath    string
	RelPath     string // Relative path from pkg/domain or pkg/infra
	Methods     []MethodInfo
	Imports     []ImportInfo
	FullPath    string // Full package path
}

type MethodInfo struct {
	Name    string
	Params  []ParamInfo
	Returns []ReturnInfo
}

type ParamInfo struct {
	Name string
	Type string
}

type ReturnInfo struct {
	Name string
	Type string
}

type ImportInfo struct {
	Path  string
	Alias string
}

func findInterfaces(rootDir string) []InterfaceInfo {
	var interfaces []InterfaceInfo

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Parse file
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil // Skip files that can't be parsed
		}

		// Get package name
		packageName := node.Name.Name

		// Extract imports
		var imports []ImportInfo
		for _, imp := range node.Imports {
			impPath := strings.Trim(imp.Path.Value, "\"")
			var alias string
			if imp.Name != nil {
				alias = imp.Name.Name
			} else {
				parts := strings.Split(impPath, "/")
				alias = parts[len(parts)-1]
			}
			imports = append(imports, ImportInfo{
				Path:  impPath,
				Alias: alias,
			})
		}

		// Calculate full package path
		relPath, _ := filepath.Rel(rootDir, filepath.Dir(path))
		fullPath := filepath.Join(rootDir, relPath)
		fullPath = filepath.ToSlash(fullPath)
		if strings.HasPrefix(rootDir, "pkg/domain") {
			fullPath = "github.com/replay-api/replay-api/pkg/domain/" + relPath
		} else if strings.HasPrefix(rootDir, "pkg/infra") {
			fullPath = "github.com/replay-api/replay-api/pkg/infra/" + relPath
		}
		fullPath = strings.ReplaceAll(fullPath, "\\", "/")

		// Find interfaces
		ast.Inspect(node, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.GenDecl:
				if x.Tok == token.TYPE {
					for _, spec := range x.Specs {
						ts, ok := spec.(*ast.TypeSpec)
						if !ok {
							continue
						}

						iface, ok := ts.Type.(*ast.InterfaceType)
						if !ok {
							continue
						}

						// Skip if it's not an interface we want to mock
						if !shouldMock(ts.Name.Name) {
							continue
						}

						// Extract methods
						methods := extractMethods(iface, fset, node)

						// Calculate relative path
						relPath, _ := filepath.Rel(rootDir, filepath.Dir(path))

						interfaces = append(interfaces, InterfaceInfo{
							Name:        ts.Name.Name,
							PackageName: packageName,
							FilePath:    path,
							RelPath:     relPath,
							Methods:     methods,
							Imports:     imports,
							FullPath:    fullPath,
						})
					}
				}
			}
			return true
		})

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
	}

	return interfaces
}

func extractMethods(iface *ast.InterfaceType, fset *token.FileSet, file *ast.File) []MethodInfo {
	var methods []MethodInfo

	for _, field := range iface.Methods.List {
		if len(field.Names) == 0 {
			continue // Embedded interface
		}

		methodName := field.Names[0].Name
		ft, ok := field.Type.(*ast.FuncType)
		if !ok {
			continue
		}

		// Extract parameters with proper type resolution
		packageName := file.Name.Name
		params := extractParams(ft.Params, fset, file, packageName)
		
		// Extract returns with proper type resolution
		returns := extractReturns(ft.Results, fset, file, packageName)

		methods = append(methods, MethodInfo{
			Name:    methodName,
			Params:  params,
			Returns: returns,
		})
	}

	return methods
}

func extractParams(fieldList *ast.FieldList, fset *token.FileSet, file *ast.File, packageName string) []ParamInfo {
	var params []ParamInfo

	if fieldList == nil {
		return params
	}

	for _, field := range fieldList.List {
		typeStr := typeToString(field.Type, file)
		
		if len(field.Names) == 0 {
			params = append(params, ParamInfo{
				Name: "",
				Type: typeStr,
			})
		} else {
			for _, name := range field.Names {
				params = append(params, ParamInfo{
					Name: name.Name,
					Type: typeStr,
				})
			}
		}
	}

	return params
}

func extractReturns(fieldList *ast.FieldList, fset *token.FileSet, file *ast.File, packageName string) []ReturnInfo {
	var returns []ReturnInfo

	if fieldList == nil {
		return returns
	}

	for _, field := range fieldList.List {
		typeStr := typeToString(field.Type, file)

		if len(field.Names) == 0 {
			returns = append(returns, ReturnInfo{
				Name: "",
				Type: typeStr,
			})
		} else {
			for _, name := range field.Names {
				returns = append(returns, ReturnInfo{
					Name: name.Name,
					Type: typeStr,
				})
			}
		}
	}

	return returns
}

func typeToString(expr ast.Expr, file *ast.File) string {
	// Build import map for type resolution
	importMap := make(map[string]string) // package path -> alias
	packageName := file.Name.Name
	
	for _, imp := range file.Imports {
		impPath := strings.Trim(imp.Path.Value, "\"")
		var alias string
		if imp.Name != nil {
			alias = imp.Name.Name
		} else {
			parts := strings.Split(impPath, "/")
			alias = parts[len(parts)-1]
		}
		importMap[impPath] = alias
	}
	
	return typeToStringRecursive(expr, file, importMap, packageName)
}

func typeToStringRecursive(expr ast.Expr, file *ast.File, importMap map[string]string, packageName string) string {
	switch x := expr.(type) {
	case *ast.Ident:
		// Identifiers in the same package don't need prefix
		return x.Name
	case *ast.SelectorExpr:
		// Handle selector expressions like package.Type
		pkgIdent, ok := x.X.(*ast.Ident)
		if !ok {
			return typeToStringRecursive(x.X, file, importMap, packageName) + "." + x.Sel.Name
		}
		
		// Check if this is a qualified import
		pkgName := pkgIdent.Name
		
		// Find which import this refers to
		for _, alias := range importMap {
			if alias == pkgName {
				// This is from an import, use the alias
				return pkgName + "." + x.Sel.Name
			}
		}
		
		// If not found in imports, it might be from the same package
		if pkgName == packageName {
			return x.Sel.Name
		}
		
		// Default: use as is
		return pkgName + "." + x.Sel.Name
	case *ast.StarExpr:
		return "*" + typeToStringRecursive(x.X, file, importMap, packageName)
	case *ast.ArrayType:
		if x.Len == nil {
			return "[]" + typeToStringRecursive(x.Elt, file, importMap, packageName)
		}
		return "[" + typeToStringRecursive(x.Len, file, importMap, packageName) + "]" + typeToStringRecursive(x.Elt, file, importMap, packageName)
	case *ast.MapType:
		return "map[" + typeToStringRecursive(x.Key, file, importMap, packageName) + "]" + typeToStringRecursive(x.Value, file, importMap, packageName)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.Ellipsis:
		return "..." + typeToStringRecursive(x.Elt, file, importMap, packageName)
	case *ast.FuncType:
		// Simplified function type
		return "func(...)"
	case *ast.BasicLit:
		return x.Value
	default:
		return "interface{}"
	}
}

func shouldMock(name string) bool {
	// Skip types that are not interfaces (commands, queries as structs, etc.)
	skipPatterns := []string{
		"Command", // Skip command structs, only mock CommandHandler
		"Query",   // Skip query structs, only mock QueryHandler
		"Request",
		"Response",
		"Result",
		"Filters",
		"Options",
	}

	for _, pattern := range skipPatterns {
		if name == pattern || strings.HasSuffix(name, pattern) {
			// But include if it's actually a Command/Query interface
			if strings.Contains(name, "Handler") || strings.Contains(name, "Reader") || strings.Contains(name, "Writer") || strings.Contains(name, "Command") || strings.Contains(name, "Query") {
				return true
			}
			return false
		}
	}

	// Include interfaces that match these patterns
	includePatterns := []string{
		"Repository",
		"CommandHandler",
		"QueryHandler",
		"Handler",
		"Service",
		"Reader",
		"Writer",
		"Adapter",
		"Provider",
		"Client",
	}

	for _, pattern := range includePatterns {
		if strings.Contains(name, pattern) {
			return true
		}
	}

	return false
}

func generateManualMock(iface InterfaceInfo, outputDir, baseDir string) error {
	// Generate mocks in test/mocks but maintain same directory structure
	// This allows types from same package to be accessible when compiled together
	mockOutputDir := filepath.Join(outputDir, "domain", iface.RelPath)
	if strings.Contains(iface.FilePath, "pkg/infra") {
		// Extract infra path
		relPath := strings.Replace(iface.RelPath, "../", "", 1)
		relPath = strings.TrimPrefix(relPath, "infra/")
		mockOutputDir = filepath.Join(outputDir, "infra", relPath)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(mockOutputDir, 0750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate mock file name
	mockFileName := strings.ToLower(iface.Name) + "_mock.go"
	mockFilePath := filepath.Join(mockOutputDir, mockFileName)

	// Generate code
	code := generateMockCode(iface)

	return os.WriteFile(mockFilePath, []byte(code), 0600)
}

func generateMockCode(iface InterfaceInfo) string {
	var buf strings.Builder

	// Package declaration - use the same package name so types from same package are accessible
	// This assumes mocks will be compiled with the original package files
	buf.WriteString("package " + iface.PackageName + "\n\n")

	// Imports
	buf.WriteString("import (\n")
	buf.WriteString("\t\"context\"\n\n")
	buf.WriteString("\t\"github.com/stretchr/testify/mock\"\n")
	
	// Add imports from the original interface file
	importMap := make(map[string]bool)
	importMap["context"] = true
	importMap["github.com/stretchr/testify/mock"] = true
	
	for _, imp := range iface.Imports {
		if imp.Path != "context" && !strings.Contains(imp.Path, "testify") {
			var impLine string
			if imp.Alias != "" {
				impLine = "\t" + imp.Alias + " \"" + imp.Path + "\"\n"
			} else {
				parts := strings.Split(imp.Path, "/")
				alias := parts[len(parts)-1]
				impLine = "\t" + alias + " \"" + imp.Path + "\"\n"
			}
			if !importMap[impLine] {
				buf.WriteString(impLine)
				importMap[impLine] = true
			}
		}
	}
	
	buf.WriteString(")\n\n")

	// Mock struct
	buf.WriteString("// Mock" + iface.Name + " is a mock implementation of " + iface.Name + "\n")
	buf.WriteString("type Mock" + iface.Name + " struct {\n")
	buf.WriteString("\tmock.Mock\n")
	buf.WriteString("}\n\n")

	// Generate methods
	for _, method := range iface.Methods {
		buf.WriteString("// " + method.Name + " provides a mock function\n")
		buf.WriteString("func (_m *Mock" + iface.Name + ") " + method.Name + "(")
		
		// Parameters
		paramNames := []string{}
		for i, param := range method.Params {
			if i > 0 {
				buf.WriteString(", ")
			}
			if param.Name != "" {
				buf.WriteString(param.Name + " ")
				paramNames = append(paramNames, param.Name)
			} else {
				buf.WriteString("arg" + fmt.Sprintf("%d ", i))
				paramNames = append(paramNames, "arg"+fmt.Sprintf("%d", i))
			}
			buf.WriteString(param.Type)
		}
		buf.WriteString(") ")

		// Return types
		if len(method.Returns) > 1 {
			buf.WriteString("(")
		}
		for i, ret := range method.Returns {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(ret.Type)
		}
		if len(method.Returns) > 1 {
			buf.WriteString(")")
		}
		buf.WriteString(" {\n")

		// Method body
		buf.WriteString("\tret := _m.Called(")
		for i, name := range paramNames {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(name)
		}
		buf.WriteString(")\n\n")

		// Handle returns
		if len(method.Returns) == 0 {
			buf.WriteString("\treturn\n")
		} else if len(method.Returns) == 1 {
			if method.Returns[0].Type == "error" {
				buf.WriteString("\treturn ret.Error(0)\n")
			} else if method.Returns[0].Type == "interface{}" {
				buf.WriteString("\tvar r0 interface{}\n")
				buf.WriteString("\tif ret.Get(0) != nil {\n")
				buf.WriteString("\t\tr0 = ret.Get(0)\n")
				buf.WriteString("\t}\n")
				buf.WriteString("\treturn r0\n")
			} else {
				buf.WriteString("\tvar r0 " + method.Returns[0].Type + "\n")
				buf.WriteString("\tif ret.Get(0) != nil {\n")
				buf.WriteString("\t\tr0 = ret.Get(0).(" + method.Returns[0].Type + ")\n")
				buf.WriteString("\t}\n")
				buf.WriteString("\treturn r0\n")
			}
		} else {
			// Multiple returns - handle each one
			for i, ret := range method.Returns {
				buf.WriteString("\tvar r" + fmt.Sprintf("%d", i) + " " + ret.Type + "\n")
			}
			buf.WriteString("\n")

			// Check for function return type
			allReturns := ""
			for i, ret := range method.Returns {
				if i > 0 {
					allReturns += ", "
				}
				allReturns += ret.Type
			}
			
			buf.WriteString("\tif rf, ok := ret.Get(0).(func(")
			for i, param := range method.Params {
				if i > 0 {
					buf.WriteString(", ")
				}
				buf.WriteString(param.Type)
			}
			buf.WriteString(") (" + allReturns + ")); ok {\n")
			buf.WriteString("\t\treturn rf(")
			for i, name := range paramNames {
				if i > 0 {
					buf.WriteString(", ")
				}
				buf.WriteString(name)
			}
			buf.WriteString(")\n")
			buf.WriteString("\t}\n\n")

			// Handle each return value
			for i, ret := range method.Returns {
				if ret.Type == "error" {
					buf.WriteString("\tr" + fmt.Sprintf("%d", i) + " = ret.Error(" + fmt.Sprintf("%d", i) + ")\n")
				} else if ret.Type == "interface{}" {
					buf.WriteString("\tif ret.Get(" + fmt.Sprintf("%d", i) + ") != nil {\n")
					buf.WriteString("\t\tr" + fmt.Sprintf("%d", i) + " = ret.Get(" + fmt.Sprintf("%d", i) + ")\n")
					buf.WriteString("\t}\n")
				} else {
					buf.WriteString("\tif rf, ok := ret.Get(" + fmt.Sprintf("%d", i) + ").(func(")
					for j, param := range method.Params {
						if j > 0 {
							buf.WriteString(", ")
						}
						buf.WriteString(param.Type)
					}
					buf.WriteString(") " + ret.Type + "); ok {\n")
					buf.WriteString("\t\tr" + fmt.Sprintf("%d", i) + " = rf(")
					for j, name := range paramNames {
						if j > 0 {
							buf.WriteString(", ")
						}
						buf.WriteString(name)
					}
					buf.WriteString(")\n")
					buf.WriteString("\t} else {\n")
					buf.WriteString("\t\tif ret.Get(" + fmt.Sprintf("%d", i) + ") != nil {\n")
					buf.WriteString("\t\t\tr" + fmt.Sprintf("%d", i) + " = ret.Get(" + fmt.Sprintf("%d", i) + ").(" + ret.Type + ")\n")
					buf.WriteString("\t\t}\n")
					buf.WriteString("\t}\n")
				}
			}

			buf.WriteString("\n\treturn ")
			for i := range method.Returns {
				if i > 0 {
					buf.WriteString(", ")
				}
				buf.WriteString("r" + fmt.Sprintf("%d", i))
			}
			buf.WriteString("\n")
		}

		buf.WriteString("}\n\n")
	}

	// NewMock function
	buf.WriteString("// NewMock" + iface.Name + " creates a new instance of Mock" + iface.Name + "\n")
	buf.WriteString("func NewMock" + iface.Name + "(t interface {\n")
	buf.WriteString("\tmock.TestingT\n")
	buf.WriteString("\tCleanup(func())\n")
	buf.WriteString("}) *Mock" + iface.Name + " {\n")
	buf.WriteString("\tmock := &Mock" + iface.Name + "{}\n")
	buf.WriteString("\tmock.Mock.Test(t)\n")
	buf.WriteString("\tt.Cleanup(func() { mock.AssertExpectations(t) })\n")
	buf.WriteString("\treturn mock\n")
	buf.WriteString("}\n")

	// Format the code
	formatted, err := format.Source([]byte(buf.String()))
	if err != nil {
		// Return unformatted if formatting fails
		return buf.String()
	}

	return string(formatted)
}
