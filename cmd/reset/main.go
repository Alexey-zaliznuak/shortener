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

// StructInfo содержит информацию о структуре, для которой нужно сгенерировать Reset()
type StructInfo struct {
	Package string
	Name    string
	Fields  []FieldInfo
	File    string
}

// FieldInfo содержит информацию о поле структуры
type FieldInfo struct {
	Name              string
	Type              string
	IsPtr             bool
	IsSlice           bool
	IsMap             bool
	IsStruct          bool
	IsAnonymousStruct bool
	AnonymousFields   []FieldInfo // Поля анонимной структуры
}

func main() {
	// Начинаем с корневой директории проекта
	rootDir := "."
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}

	// Собираем все структуры с комментарием // generate:reset
	structs := make(map[string][]StructInfo) // ключ - путь к пакету

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Пропускаем служебные директории
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}
			if info.Name() == "vendor" || info.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		// Обрабатываем только .go файлы, исключая сгенерированные
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.HasSuffix(path, ".gen.go") {
			return nil
		}

		// Парсим файл
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			// Игнорируем ошибки парсинга (могут быть файлы с синтаксическими ошибками)
			return nil
		}

		// Находим структуры с комментарием // generate:reset
		ast.Inspect(node, func(n ast.Node) bool {
			genDecl, ok := n.(*ast.GenDecl)
			if !ok {
				return true
			}

			// Проверяем комментарии
			hasResetComment := false
			if genDecl.Doc != nil {
				for _, comment := range genDecl.Doc.List {
					if strings.Contains(comment.Text, "generate:reset") {
						hasResetComment = true
						break
					}
				}
			}

			if !hasResetComment {
				return true
			}

			// Обрабатываем только type declarations
			if genDecl.Tok != token.TYPE {
				return true
			}

			// Обрабатываем каждое объявление типа
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				// Собираем информацию о полях
				fields := extractFields(structType, fset)

				// Определяем путь к пакету
				packagePath := filepath.Dir(path)
				packageName := node.Name.Name

				structInfo := StructInfo{
					Package: packageName,
					Name:    typeSpec.Name.Name,
					Fields:  fields,
					File:    path,
				}

				structs[packagePath] = append(structs[packagePath], structInfo)
			}

			return true
		})

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка при обходе директорий: %v\n", err)
		os.Exit(1)
	}

	// Генерируем файлы reset.gen.go для каждого пакета
	for packagePath, structsList := range structs {
		if len(structsList) == 0 {
			continue
		}

		// Генерируем код
		generatedCode := generateResetMethods(structsList)

		// Записываем в файл reset.gen.go
		outputPath := filepath.Join(packagePath, "reset.gen.go")
		err := os.WriteFile(outputPath, []byte(generatedCode), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка при записи файла %s: %v\n", outputPath, err)
			continue
		}

		// Форматируем код
		formatted, err := format.Source([]byte(generatedCode))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка при форматировании %s: %v\n", outputPath, err)
			continue
		}

		err = os.WriteFile(outputPath, formatted, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка при записи отформатированного файла %s: %v\n", outputPath, err)
			continue
		}

		fmt.Printf("Сгенерирован файл: %s\n", outputPath)
	}
}

// extractFields извлекает информацию о полях структуры
func extractFields(structType *ast.StructType, fset *token.FileSet) []FieldInfo {
	var fields []FieldInfo

	for _, field := range structType.Fields.List {
		fieldType := getTypeString(field.Type, fset)

		// Определяем базовый тип (без указателя)
		baseType := getBaseType(field.Type)

		// Проверяем, является ли тип анонимной структурой
		isAnonymousStruct := false
		var anonymousFields []FieldInfo
		if structTypeField, ok := baseType.(*ast.StructType); ok {
			isAnonymousStruct = true
			anonymousFields = extractFields(structTypeField, fset)
		}

		fieldInfo := FieldInfo{
			Type:              fieldType,
			IsPtr:             isPointerType(field.Type),
			IsSlice:           isSliceType(baseType),
			IsMap:             isMapType(baseType),
			IsStruct:          isStructType(baseType),
			IsAnonymousStruct: isAnonymousStruct,
			AnonymousFields:   anonymousFields,
		}

		// Обрабатываем имена полей
		if len(field.Names) == 0 {
			// Анонимное поле
			fields = append(fields, fieldInfo)
		} else {
			for _, name := range field.Names {
				fieldInfo.Name = name.Name
				fields = append(fields, fieldInfo)
			}
		}
	}

	return fields
}

// getBaseType возвращает базовый тип, убирая указатель
func getBaseType(expr ast.Expr) ast.Expr {
	if starExpr, ok := expr.(*ast.StarExpr); ok {
		return getBaseType(starExpr.X)
	}
	return expr
}

// getTypeString возвращает строковое представление типа
func getTypeString(expr ast.Expr, fset *token.FileSet) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + getTypeString(t.X, fset)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + getTypeString(t.Elt, fset)
		}
		return "[" + getTypeString(t.Len, fset) + "]" + getTypeString(t.Elt, fset)
	case *ast.MapType:
		return "map[" + getTypeString(t.Key, fset) + "]" + getTypeString(t.Value, fset)
	case *ast.SelectorExpr:
		return getTypeString(t.X, fset) + "." + t.Sel.Name
	default:
		return "unknown"
	}
}

// isPointerType проверяет, является ли тип указателем
func isPointerType(expr ast.Expr) bool {
	_, ok := expr.(*ast.StarExpr)
	return ok
}

// isSliceType проверяет, является ли тип слайсом
func isSliceType(expr ast.Expr) bool {
	arrayType, ok := expr.(*ast.ArrayType)
	if !ok {
		return false
	}
	return arrayType.Len == nil
}

// isMapType проверяет, является ли тип мапой
func isMapType(expr ast.Expr) bool {
	_, ok := expr.(*ast.MapType)
	return ok
}

// isStructType проверяет, является ли тип структурой
func isStructType(expr ast.Expr) bool {
	// Убираем указатель, если есть
	baseType := getBaseType(expr)
	_, ok := baseType.(*ast.StructType)
	if ok {
		return true
	}
	// Также проверяем, может быть это идентификатор типа структуры
	// (для этого нужно проверить, что это не примитивный тип)
	if ident, ok := baseType.(*ast.Ident); ok {
		// Проверяем, что это не встроенный тип
		builtinTypes := map[string]bool{
			"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
			"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true, "uintptr": true,
			"float32": true, "float64": true,
			"complex64": true, "complex128": true,
			"string": true, "bool": true, "byte": true, "rune": true,
		}
		return !builtinTypes[ident.Name]
	}
	return false
}

// generateResetMethods генерирует код методов Reset() для всех структур
func generateResetMethods(structs []StructInfo) string {
	if len(structs) == 0 {
		return ""
	}

	var builder strings.Builder
	packageName := structs[0].Package

	// Заголовок файла
	builder.WriteString("// Code generated by reset tool. DO NOT EDIT.\n")
	builder.WriteString("//go:generate go run github.com/Alexey-zaliznuak/shortener/cmd/reset\n\n")
	builder.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	// Генерируем метод Reset() для каждой структуры
	for _, s := range structs {
		builder.WriteString(generateResetMethod(s))
		builder.WriteString("\n")
	}

	return builder.String()
}

// generateResetMethod генерирует код метода Reset() для одной структуры
func generateResetMethod(s StructInfo) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("func (rs *%s) Reset() {\n", s.Name))
	builder.WriteString("\tif rs == nil {\n")
	builder.WriteString("\t\treturn\n")
	builder.WriteString("\t}\n\n")

	for _, field := range s.Fields {
		if field.Name == "" {
			// Анонимное поле - пропускаем или обрабатываем по-другому
			continue
		}

		fieldName := "rs." + field.Name

		if field.IsPtr {
			// Указатель
			if field.IsSlice {
				// Указатель на слайс - проверяем на nil и обрезаем
				builder.WriteString(fmt.Sprintf("\tif %s != nil {\n", fieldName))
				builder.WriteString(fmt.Sprintf("\t\t*%s = (*%s)[:0]\n", fieldName, fieldName))
				builder.WriteString("\t}\n")
			} else if field.IsMap {
				// Указатель на мапу - проверяем на nil и очищаем
				builder.WriteString(fmt.Sprintf("\tif %s != nil {\n", fieldName))
				builder.WriteString(fmt.Sprintf("\t\tclear(*%s)\n", fieldName))
				builder.WriteString("\t}\n")
			} else if field.IsStruct {
				// Указатель на структуру
				builder.WriteString(fmt.Sprintf("\tif %s != nil {\n", fieldName))
				if field.IsAnonymousStruct {
					// Анонимная структура - сбрасываем поля напрямую
					generateResetForFields(&builder, field.AnonymousFields, "*"+fieldName, "\t\t")
				} else {
					// Именованная структура - проверяем наличие метода Reset() (как в примере)
					builder.WriteString(fmt.Sprintf("\t\tif resetter, ok := interface{}(%s).(interface{ Reset() }); ok {\n", fieldName))
					builder.WriteString("\t\t\tresetter.Reset()\n")
					builder.WriteString("\t\t} else {\n")
					builder.WriteString(fmt.Sprintf("\t\t\t*%s = %s{}\n", fieldName, strings.TrimPrefix(field.Type, "*")))
					builder.WriteString("\t\t}\n")
				}
				builder.WriteString("\t}\n")
			} else {
				// Примитивный указатель - проверяем на nil и сбрасываем
				builder.WriteString(fmt.Sprintf("\tif %s != nil {\n", fieldName))
				baseType := strings.TrimPrefix(field.Type, "*")
				zeroValue := getZeroValue(baseType)
				builder.WriteString(fmt.Sprintf("\t\t*%s = %s\n", fieldName, zeroValue))
				builder.WriteString("\t}\n")
			}
		} else if field.IsSlice {
			// Слайс (не указатель)
			builder.WriteString(fmt.Sprintf("\t%s = %s[:0]\n", fieldName, fieldName))
		} else if field.IsMap {
			// Мапа (не указатель)
			builder.WriteString(fmt.Sprintf("\tclear(%s)\n", fieldName))
		} else if field.IsStruct {
			if field.IsAnonymousStruct {
				// Анонимная структура - сбрасываем поля напрямую
				generateResetForFields(&builder, field.AnonymousFields, fieldName, "\t")
			} else {
				// Именованная структура - проверяем наличие метода Reset()
				// Если метода нет, сбрасываем к нулевому значению
				builder.WriteString(fmt.Sprintf("\tif resetter, ok := interface{}(&%s).(interface{ Reset() }); ok {\n", fieldName))
				builder.WriteString("\t\tresetter.Reset()\n")
				builder.WriteString("\t} else {\n")
				builder.WriteString(fmt.Sprintf("\t\t%s = %s{}\n", fieldName, field.Type))
				builder.WriteString("\t}\n")
			}
		} else {
			// Примитивный тип
			zeroValue := getZeroValue(field.Type)
			builder.WriteString(fmt.Sprintf("\t%s = %s\n", fieldName, zeroValue))
		}
	}

	builder.WriteString("}\n")
	return builder.String()
}

// generateResetForFields генерирует код для сброса полей структуры
func generateResetForFields(builder *strings.Builder, fields []FieldInfo, prefix string, indent string) {
	for _, field := range fields {
		if field.Name == "" {
			// Анонимное поле - пропускаем
			continue
		}

		fieldName := prefix + "." + field.Name

		if field.IsPtr {
			// Указатель
			if field.IsSlice {
				// Указатель на слайс
				builder.WriteString(fmt.Sprintf("%sif %s != nil {\n", indent, fieldName))
				builder.WriteString(fmt.Sprintf("%s\t*%s = (*%s)[:0]\n", indent, fieldName, fieldName))
				builder.WriteString(fmt.Sprintf("%s}\n", indent))
			} else if field.IsMap {
				// Указатель на мапу
				builder.WriteString(fmt.Sprintf("%sif %s != nil {\n", indent, fieldName))
				builder.WriteString(fmt.Sprintf("%s\tclear(*%s)\n", indent, fieldName))
				builder.WriteString(fmt.Sprintf("%s}\n", indent))
			} else if field.IsStruct {
				// Указатель на структуру
				builder.WriteString(fmt.Sprintf("%sif %s != nil {\n", indent, fieldName))
				if field.IsAnonymousStruct {
					// Анонимная структура
					generateResetForFields(builder, field.AnonymousFields, "*"+fieldName, indent+"\t")
				} else {
					// Именованная структура
					builder.WriteString(fmt.Sprintf("%s\tif resetter, ok := interface{}(%s).(interface{ Reset() }); ok {\n", indent, fieldName))
					builder.WriteString(fmt.Sprintf("%s\t\tresetter.Reset()\n", indent))
					builder.WriteString(fmt.Sprintf("%s\t} else {\n", indent))
					builder.WriteString(fmt.Sprintf("%s\t\t*%s = %s{}\n", indent, fieldName, strings.TrimPrefix(field.Type, "*")))
					builder.WriteString(fmt.Sprintf("%s\t}\n", indent))
				}
				builder.WriteString(fmt.Sprintf("%s}\n", indent))
			} else {
				// Примитивный указатель
				builder.WriteString(fmt.Sprintf("%sif %s != nil {\n", indent, fieldName))
				baseType := strings.TrimPrefix(field.Type, "*")
				zeroValue := getZeroValue(baseType)
				builder.WriteString(fmt.Sprintf("%s\t*%s = %s\n", indent, fieldName, zeroValue))
				builder.WriteString(fmt.Sprintf("%s}\n", indent))
			}
		} else if field.IsSlice {
			// Слайс (не указатель)
			builder.WriteString(fmt.Sprintf("%s%s = %s[:0]\n", indent, fieldName, fieldName))
		} else if field.IsMap {
			// Мапа (не указатель)
			builder.WriteString(fmt.Sprintf("%sclear(%s)\n", indent, fieldName))
		} else if field.IsStruct {
			if field.IsAnonymousStruct {
				// Анонимная структура
				generateResetForFields(builder, field.AnonymousFields, fieldName, indent)
			} else {
				// Именованная структура
				builder.WriteString(fmt.Sprintf("%sif resetter, ok := interface{}(&%s).(interface{ Reset() }); ok {\n", indent, fieldName))
				builder.WriteString(fmt.Sprintf("%s\tresetter.Reset()\n", indent))
				builder.WriteString(fmt.Sprintf("%s} else {\n", indent))
				builder.WriteString(fmt.Sprintf("%s\t%s = %s{}\n", indent, fieldName, field.Type))
				builder.WriteString(fmt.Sprintf("%s}\n", indent))
			}
		} else {
			// Примитивный тип
			zeroValue := getZeroValue(field.Type)
			builder.WriteString(fmt.Sprintf("%s%s = %s\n", indent, fieldName, zeroValue))
		}
	}
}

// getZeroValue возвращает нулевое значение для типа
func getZeroValue(typeStr string) string {
	// Убираем указатель, если есть
	typeStr = strings.TrimPrefix(typeStr, "*")

	switch typeStr {
	case "int", "int8", "int16", "int32", "int64":
		return "0"
	case "uint", "uint8", "uint16", "uint32", "uint64", "uintptr":
		return "0"
	case "float32", "float64":
		return "0"
	case "complex64", "complex128":
		return "0"
	case "string":
		return `""`
	case "bool":
		return "false"
	case "byte":
		return "0"
	case "rune":
		return "0"
	default:
		// Для пользовательских типов возвращаем нулевое значение через {}
		return typeStr + "{}"
	}
}
