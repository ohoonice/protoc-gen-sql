package internal

import (
	"bytes"
	"strings"
	"text/template"
)

var sqlTemplate = `
{{- if ne .Database "" }}
CREATE DATABASE IF NOT EXISTS {{.Database}} COLLATE utf8mb4_unicode_ci;

USE {{.Database}};
{{- end}}

CREATE TABLE {{.Table}} (
{{- range $field := .Fields}}
	{{$field}}
{{- end}}
{{- range $key := .UniqueKeys}}
	UNIQUE KEY {{$key}},
{{- end}}
{{- range $key := .Keys}}
	KEY {{$key}},
{{- end}}
    PRIMARY KEY {{.PrimaryKey}}
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

type tableDesc struct {
	Database   string
	Table      string
	Fields     []string
	PrimaryKey string
	UniqueKeys []string
	Keys       []string
}

func (s *tableDesc) execute() string {
	buf := new(bytes.Buffer)

	tmpl, err := template.New("sql").Parse(strings.TrimSpace(sqlTemplate))
	if err != nil {
		panic(err)
	}

	if err := tmpl.Execute(buf, s); err != nil {
		panic(err)
	}

	return strings.Trim(string(buf.Bytes()), "\r\n")
}
