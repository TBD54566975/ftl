package outer

import "ftl/outer/inner"

//ftl:data export
type Event struct {
	Value inner.EchoResponse
}
