package auth

import (
	"embed"
	"fmt"
	"github.com/zatrano/framework/contracts"
	"os"
	"path/filepath"

	"github.com/zatrano/packages/bootutil"
)

//go:embed all:stubs
var stubFiles embed.FS

func readStub(app contracts.App, rel string) ([]byte, error) {
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
