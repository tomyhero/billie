package filter

import (
	"bytes"
	"mime/multipart"
	"strings"
	"text/template"
)

const defaultTemplates = `{{range $x , $values := .fields}}
{{$values.name }} : {{join $values.value ","}}{{end}}
{{range $x , $attachments := .attachment_fields}}
{{$attachments.name }}:{{ attachmentJoin $attachments.value }}{{end}}
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

func (self *Text) Parse(fields []map[string]interface{}, attachmentFields []map[string]interface{}) string {

	t, err := template.New("TextTemplate").Funcs(fns).Parse(defaultTemplates)
	if err != nil {
		panic(err)
	}

	var buffer bytes.Buffer
	err = t.ExecuteTemplate(&buffer, "TextTemplate", map[string]interface{}{"fields": fields, "attachment_fields": attachmentFields})
	if err != nil {
		panic(err)
	}

	return buffer.String()
}
