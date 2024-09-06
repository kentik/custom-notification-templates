package templates

import (
	"embed"
)

//go:embed *.tmpl
var fs embed.FS

func AssetNames() []string {
	var out []string
	dirents, err := fs.ReadDir(".")
	if err != nil {
		panic("BUG: empty go:embed FS")
	}

	for _, dirent := range dirents {
		if !dirent.IsDir() {
			out = append(out, dirent.Name())
		}
	}

	return out
}

func Asset(an string) ([]byte, error) {
	return fs.ReadFile(an)
}
