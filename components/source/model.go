package source

import "goldenglow/components"

type Source interface {
	C() <-chan components.Message
}
