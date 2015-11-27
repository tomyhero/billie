package core

import (
	"mime/multipart"
)

const (
	FIELD_TYPE_TEXT       = 1
	FIELD_TYPE_ATTACHMENT = 2
)

type Field struct {
	Name        string
	FieldType   int
	Values      []string
	Attachments []*multipart.FileHeader
}

func (self Field) IsTextType() bool {
	if self.FieldType == FIELD_TYPE_TEXT {
		return true
	} else {
		return false
	}
}
func (self Field) IsAttachmentType() bool {
	if self.FieldType == FIELD_TYPE_ATTACHMENT {
		return true
	} else {
		return false
	}
}
