package filter

import (
	"mime/multipart"
)

type FilterExecutor interface {
	Parse(map[string]interface{}, map[string][]*multipart.FileHeader) string
}
