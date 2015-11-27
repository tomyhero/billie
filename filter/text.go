package filter

import (
	"bytes"
	. "github.com/tomyhero/billie/core"
	"mime/multipart"
	"strings"
	"text/template"
)

const defaultTemplates = `{{range $x , $field := .fields}}
{{$field.Name }} : {{ if $field.IsTextType }}{{join $field.Values ","}}{{ else }}{{ attachmentJoin $field.Attachments }}{{ end }}
{{end}}
`

var fns = template.FuncMap{
	"join": strings.Join,
	"attachmentJoin": func(attachments []*multipart.FileHeader) string {

		fileNames := []string{}
		for _, attachment := range attachments {
			fileNames = append(fileNames, attachment.Filename)
		}
		return strings.Join(fileNames, ",")
	},
}

type Text struct {
}

func (self *Text) Parse(fields []Field) string {

	t, err := template.New("TextTemplate").Funcs(fns).Parse(defaultTemplates)
	if err != nil {
		panic(err)
	}

	var buffer bytes.Buffer
	err = t.ExecuteTemplate(&buffer, "TextTemplate", map[string]interface{}{"fields": fields})
	if err != nil {
		panic(err)
	}

	return buffer.String()
}
