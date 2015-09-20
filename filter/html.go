package filter

import (
	"bytes"
	"html/template"
	"mime/multipart"

	log "github.com/Sirupsen/logrus"
)

var defaultTemplate = template.Must(template.New("default.html").Funcs(template.FuncMap(fns)).ParseFiles("filter/template/default.html"))

// HTML is format of filter. it converts input to html and output.
type HTML struct {
}

// Parse is converts the input data into html.
func (h *HTML) Parse(f map[string]interface{}, a map[string][]*multipart.FileHeader) string {
	var buffer bytes.Buffer
	err := defaultTemplate.Execute(&buffer, map[string]interface{}{"fields": f, "attachment_fields": a})
	if err != nil {
		log.Panicf("execute template error: %v", err)
	}

	return buffer.String()
}
