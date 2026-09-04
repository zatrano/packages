package auth

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zatrano/framework/kernel"
	"github.com/zatrano/packages/bootutil"
)

//go:embed all:stubs
var stubFiles embed.FS

func readStub(app *kernel.Application, rel string) ([]byte, error) {
	slash := filepath.ToSlash(rel)
	if b, err := stubFiles.ReadFile("stubs/" + slash); err == nil {
		return b, nil
	}
	if root := bootutil.ConsoleStubsDir(app); root != "" {
		b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(slash)))
		if err == nil {
			return b, nil
		}
	}
	return nil, fmt.Errorf("auth stub not found: %s", rel)
}
