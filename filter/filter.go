package filter

import (
//"mime/multipart"
)

type FilterExecutor interface {
	Parse([]map[string]interface{}) string
}
