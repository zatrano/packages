package auth

import (
	"embed"
	"fmt"
	"path/filepath"

	"github.com/zatrano/framework/v2/contracts"
)

//go:embed all:stubs
var stubFiles embed.FS

func readStub(app contracts.App, rel string) ([]byte, error) {
	_ = app
	slash := filepath.ToSlash(rel)
	b, err := stubFiles.ReadFile("stubs/" + slash)
	if err != nil {
		return nil, fmt.Errorf("auth stub not found: %s", rel)
	}
	return b, nil
}
