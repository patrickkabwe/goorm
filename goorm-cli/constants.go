package main

const (
	DB_OPT        = "goorm:"
	DB_COL        = "db_col:"
	DB_OPT_LOOKUP = "goorm"
	DB_COL_LOOKUP = "db_col"
)

const testDB = "postgres://postgres:postgres@localhost:5432/goorm-test?sslmode=disable"

const modelTemplate = `
// Code generated by go generate; DO NOT EDIT.
package {{.PackageName}}

type GeneratedDB struct {
    *DB
    {{- range .Models}}
    {{.Name}} *{{.Name}}Model
    {{- end}}
}

func NewGoorm(cfg *GoormConfig) *GeneratedDB {
 	engine, err := NewEngine(cfg.Driver, cfg.DSN, cfg.Logger)
	if err != nil {
		panic(err)
	}
	db := NewDB(engine, cfg.Logger)	
    gdb := &GeneratedDB{
        DB: db,
        {{- range .Models}}
        {{.Name}}: &{{.Name}}Model{BaseModel: NewBaseModel[{{.Name}}](db)},
        {{- end}}
    }

    // Initialize relationships
    {{- range .Models}}
    {{- if .Relations}}
    gdb.{{.Name}}.initRelations()
    {{- end}}
    {{- end}}

    return gdb
}

{{range .Models}}
type {{.Name}}Model struct {
    *BaseModel[{{.Name}}]
}

{{if .Relations}}
func (m *{{.Name}}Model) initRelations() {
    {{- range .Relations}}
    m.RegisterRelation(Relation{
        Name:       "{{.Name}}",
        Type:       "{{.Type}}",
        Model:      {{.Model}}{},
        ForeignKey: "{{.ForeignKey}}",
        References: "{{.References}}",
    })
    {{- end}}
}
{{end}}
{{end}}
`

const migrationTemplate = `-- Migration: {{.MigrationName}}
-- Created at: {{.Timestamp}}

{{ range .Tables }}
CREATE TABLE IF NOT EXISTS {{.Name}} (
    {{- range .Columns}}
    {{.Name}} {{.Type}}{{if .Options}} {{.Options}}{{end}},
    {{- end}}
    PRIMARY KEY (id){{if .ForeignKeys}},
    {{- range .ForeignKeys}}
    CONSTRAINT {{.Name}} FOREIGN KEY ({{.Column}}) REFERENCES {{.RefTable}} ({{.RefColumn}}) {{.Options}}{{if not .Last}},{{end}}
    {{- end}}{{end}}
);
{{end}}

-- Create indexes
{{- range .IndexOrder}}
{{- if .Unique}}
CREATE UNIQUE INDEX IF NOT EXISTS {{.Name}} ON {{.Table}} ({{.Columns}});
{{- else}}
CREATE INDEX IF NOT EXISTS {{.Name}} ON {{.Table}} ({{.Columns}});
{{- end}}
{{- end}}



-- Rollback SQL
/*
{{- range $i := .DropOrder }}
DROP TABLE IF EXISTS {{$i}};
{{- end }}

{{range .IndexOrder}}
DROP INDEX IF EXISTS {{.Name}};
{{- end }}
*/
`

var DB_OPTIONS_TO_OMIT = []string{"type", "column", "options","index","goorm","db_col"}