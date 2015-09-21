package filter

import (
	"bytes"
	"html/template"
	"mime/multipart"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
)

var defaultTemplate = template.Must(template.New("").Funcs(template.FuncMap(fns)).Parse(defaultTemplateStr))

// HTML is format of filter. it converts input to html and output.
type HTML struct {
	Template string
}

// Parse is converts the input data into html.
func (h *HTML) Parse(f map[string]interface{}, a map[string][]*multipart.FileHeader) string {
	t := defaultTemplate

	if h.Template != "" {
		// template check
		_, err := os.Stat(h.Template)
		if err == nil {
			t = template.Must(template.New(filepath.Base(h.Template)).Funcs(template.FuncMap(fns)).ParseFiles(h.Template))
		} else {
			log.Warnf("template does not exist: %s, stat error: %v", h.Template, err)
		}
	}

	var buffer bytes.Buffer
	err := t.Execute(&buffer, map[string]interface{}{"fields": f, "attachment_fields": a})
	if err != nil {
		log.Panicf("execute template error: %v", err)
	}

	return buffer.String()
}

const defaultTemplateStr = `
<html>
	<head>
		<style type="text/css">
		<!--
		table {
			width: 70%;
			border-top: 1px solid #000000;
			border-left: 1px solid #000000;
			border-spacing: 0px;
		}
		table tr th, table tr td {
			border-bottom: 1px solid #000000;
			border-right: 1px solid #000000;
		}
		table tr th {
			width: 30%;
			background: #e0ffff;
		}
		table tr td {
			width: 70%;
		}
		-->
		</style>
	</head>
	<body>
		<table>
			{{range $name, $values := .fields}}
			<tr><th>{{$name}}</th><td>{{join $values ","}}</td></tr>
			{{end}}
			{{range $name, $attachments := .attachment_fields}}
			<tr><th>{{$name}}</th><td>{{attachmentJoin $attachments}}</td></tr>
			{{end}}
		</table>
	</body>
</html>
`
