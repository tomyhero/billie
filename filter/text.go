package filter

import (
	"bytes"
	"mime/multipart"
	"strings"
	"text/template"
)

const defaultTemplates = `{{range $name , $values := .fields}}
{{$name}} : {{join $values ","}}{{end}}
{{range $name , $attachments := .attachment_fields}}
{{$name}}:{{ attachmentJoin $attachments }}{{end}}
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

func (self *Text) Parse(d map[string]interface{}, a map[string][]*multipart.FileHeader) string {

	t, err := template.New("TextTemplate").Funcs(fns).Parse(defaultTemplates)
	if err != nil {
		panic(err)
	}

	var buffer bytes.Buffer
	err = t.ExecuteTemplate(&buffer, "TextTemplate", map[string]interface{}{"fields": d, "attachment_fields": a})
	if err != nil {
		panic(err)
	}

	return buffer.String()
}
