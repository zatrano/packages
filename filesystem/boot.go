package filesystem

import (
	"strings"

	"github.com/zatrano/framework/v2/contracts"
	"github.com/zatrano/framework/v2/kernel/env"
)

func boot(app contracts.App) error {
	localDisk, err := NewLocalDisk(app.BasePath("storage", "app"))
	if err != nil {
		return err
	}
	publicDisk, err := NewLocalDisk(app.BasePath("storage", "app", "public"))
	if err != nil {
		return err
	}
	appURL := app.Config().GetString("app.url", env.Get("APP_URL", "http://localhost:8080"))
	publicDisk.SetBaseURL(strings.TrimRight(appURL, "/") + "/storage")
	signingKey := app.Config().GetString("app.key", env.Get("APP_KEY", "zatrano-dev-key"))
	localDisk.SetSigningKey(signingKey)
	localDisk.SetServePath("/storage/temporary")
	localDisk.SetBaseURL(strings.TrimRight(appURL, "/"))
	publicDisk.SetSigningKey(signingKey)
	publicDisk.SetServePath("/storage/temporary")
	mgr := NewManager(env.Get("FILESYSTEM_DISK", "local"), map[string]Disk{
		"local":  localDisk,
		"public": publicDisk,
		"s3": NewCloudDisk(
			env.Get("AWS_BUCKET", "zatrano"),
			env.Get("AWS_URL", "https://s3.example.com"),
		),
	})
	app.Container().Instance("files", mgr)
	return nil
}
