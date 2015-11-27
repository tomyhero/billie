package filter

import (
	. "github.com/tomyhero/billie/core"
)

type FilterExecutor interface {
	Parse([]Field) string
}
