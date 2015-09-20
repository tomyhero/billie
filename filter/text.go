package filter

import (
	"fmt"
	"mime/multipart"
	"strings"
)

type Text struct {
}

func (self *Text) Parse(d map[string]interface{}, a map[string][]*multipart.FileHeader) string {
	body := ""
	for name, v := range d {
		values := v.([]string)
		line := fmt.Sprintf("%s : %s\n", name, strings.Join(values, ","))
		body = body + line
	}

	for name, tmp := range a {
		fileNames := []string{}
		for _, attachment := range tmp {
			fileNames = append(fileNames, attachment.Filename)
		}
		line := fmt.Sprintf("%s : %s\n", name, strings.Join(fileNames, ","))
		body = body + line
	}
	return body
}
