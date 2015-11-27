package filter

import (
	"bytes"
	"mime/multipart"
	"strings"
	"text/template"
)

const defaultTemplates = `{{range $x , $values := .fields}}
{{$values.name }} : {{ if eq $values.type 1 }}{{join $values.value ","}}{{ else }}{{ attachmentJoin $values.value }}{{ end }}
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

func (self *Text) Parse(fields []map[string]interface{}) string {

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
