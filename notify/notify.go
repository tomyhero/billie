package notify

import (
	"mime/multipart"
)

type NotifyExecutor interface {
	Notify(string, [][]*multipart.FileHeader)
}
