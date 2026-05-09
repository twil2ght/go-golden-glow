// Package mount is used to mount on all plugins
package mount

import (
	_ "goldenglow/plugin/builtin/builder"
	_ "goldenglow/plugin/builtin/calculator"
	_ "goldenglow/plugin/builtin/repoaddon"
	_ "goldenglow/plugin/builtin/safeteach"
	_ "goldenglow/plugin/builtin/speaker"
	_ "goldenglow/plugin/builtin/templateaddon"
	_ "goldenglow/plugin/builtin/timer"
	_ "goldenglow/plugin/builtin/tracer"
	_ "goldenglow/plugin/builtin/word"
)
