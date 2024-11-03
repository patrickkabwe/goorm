package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:     "generate",
	Aliases: []string{"-g"},
	Short:   "Generates a typed model",
	Args:    cobra.ExactArgs(0),
	Long:    "Generates a typed model from a database schema",
	Run: func(cmd *cobra.Command, args []string) {
		generate()
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}


type Relation struct {
	Name       string
	Type       string // "hasOne", "hasMany", "belongsTo"
	Model      string
	ForeignKey string
	References string
}

type Model struct {
	Name      string
	Relations []Relation
}

type GenerateModelsConfig struct {
	PackageName string
	Models      []Model
}

func generate() {
	// Parse the source file to find model structs
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "../models.go", nil, parser.ParseComments)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse models.go: %v", err))
	}

	var models []Model

	// Find model structs and their relationships
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			if structType, ok := x.Type.(*ast.StructType); ok {
				model := Model{Name: x.Name.Name}

				// Look for relationships in struct fields
				for _, field := range structType.Fields.List {
					if len(field.Names) == 0 {
						continue
					}

					fieldName := field.Names[0].Name

					// Skip db tag fields
					if field.Tag != nil {
						tagValue := strings.Trim(field.Tag.Value, "`")
						if strings.Contains(tagValue, DB_COL) {
							continue
						}
					}

					// Check field type for relationships
					switch t := field.Type.(type) {
					case *ast.ArrayType:
						// Has Many relationship
						if ident, ok := t.Elt.(*ast.Ident); ok {
							model.Relations = append(model.Relations, Relation{
								Name:       fieldName,
								Type:       "hasMany",
								Model:      ident.Name,
								ForeignKey: strings.TrimSuffix(model.Name, "s") + "ID",
								References: "ID",
							})
						}
					case *ast.StarExpr:
						// Has One or Belongs To relationship
						if ident, ok := t.X.(*ast.Ident); ok {
							if strings.HasSuffix(fieldName, "ID") {
								continue // Skip ID fields
							}
							// Determine relationship type based on field name
							relationType := "hasOne"
							foreignKey := model.Name + "ID"

							// If field name matches another model, it's likely belongsTo
							if strings.EqualFold(fieldName, ident.Name) {
								relationType = "belongsTo"
								foreignKey = fieldName + "ID"
							}

							model.Relations = append(model.Relations, Relation{
								Name:       fieldName,
								Type:       relationType,
								Model:      ident.Name,
								ForeignKey: foreignKey,
								References: "ID",
							})
						}
					}
				}

				models = append(models, model)
			}
		}
		return true
	})

	// Create output directory if it doesn't exist
	outDir := filepath.Dir("../generated_models.go")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create output directory: %v", err))
	}

	// Generate the code
	tmpl := template.Must(template.New("models").Parse(modelTemplate))

	out, err := os.Create("../generated_models.go")
	if err != nil {
		panic(fmt.Sprintf("Failed to create output file: %v", err))
	}
	defer out.Close()

	config := GenerateModelsConfig{
		PackageName: "goorm",
		Models:      models,
	}

	if err := tmpl.Execute(out, config); err != nil {
		panic(fmt.Sprintf("Failed to execute template: %v", err))
	}
}
