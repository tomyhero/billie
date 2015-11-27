package notify

import (
//"mime/multipart"
)

type NotifyExecutor interface {
	Notify(string, []map[string]interface{})
}
