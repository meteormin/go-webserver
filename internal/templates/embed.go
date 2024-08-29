package templates

import "embed"

//go:embed *.tmpl
var fs embed.FS

func GetFS() embed.FS {
	return fs
}
