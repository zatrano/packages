package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zatrano/framework/bootstrap/addons"
	"github.com/zatrano/framework/contracts"
)

func backupCLI(app contracts.App) []addons.CLICommand {
	return []addons.CLICommand{
		{Name: "db:backup", Description: "Backup the default (or --connection) database using native tools", Handle: func(args []string) error {
			return runBackupCreate(app, args)
		}},
		{Name: "db:backup:list", Description: "List database backups for the default (or --connection) target directory", Handle: func(args []string) error {
			return runBackupList(app, args)
		}},
		{Name: "db:restore", Description: "Restore the default (or --connection) database from a backup file", Handle: func(args []string) error {
			return runBackupRestore(app, args)
		}},
	}
}

func backupManager(app contracts.App, args []string) (*Manager, []string, error) {
	if err := app.Bootstrap(); err != nil {
		return nil, nil, err
	}
	connection, rest := parseBackupFlags(args)
	mgr, err := ManagerFromApp(app, connection)
	if err != nil {
		return nil, nil, err
	}
	return mgr, rest, nil
}

func parseBackupFlags(args []string) (connection string, rest []string) {
	rest = make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--connection" || arg == "-c":
			if i+1 < len(args) {
				connection = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "--connection="):
			connection = strings.TrimPrefix(arg, "--connection=")
		case strings.HasPrefix(arg, "-c="):
			connection = strings.TrimPrefix(arg, "-c=")
		default:
			rest = append(rest, arg)
		}
	}
	return connection, rest
}

func runBackupCreate(app contracts.App, args []string) error {
	mgr, rest, err := backupManager(app, args)
	if err != nil {
		return err
	}
	label := ""
	for i := 0; i < len(rest); i++ {
		if (rest[i] == "--label" || rest[i] == "-l") && i+1 < len(rest) {
			label = rest[i+1]
			i++
		}
	}
	path, err := mgr.Create(label)
	if err != nil {
		return err
	}
	fmt.Printf("Database backup created (%s): %s\n", mgr.Driver(), path)
	return nil
}

func runBackupList(app contracts.App, args []string) error {
	mgr, _, err := backupManager(app, args)
	if err != nil {
		return err
	}
	files, err := mgr.List()
	if err != nil {
		return err
	}
	if len(files) == 0 {
		fmt.Println("No backups found.")
		return nil
	}
	for _, file := range files {
		info, _ := os.Stat(file)
		size := int64(0)
		mod := ""
		if info != nil {
			size = info.Size()
			mod = info.ModTime().Format(time.RFC3339)
		}
		fmt.Printf("%s\t%d bytes\t%s\n", filepath.Base(file), size, mod)
	}
	return nil
}

func runBackupRestore(app contracts.App, args []string) error {
	mgr, rest, err := backupManager(app, args)
	if err != nil {
		return err
	}
	if len(rest) == 0 {
		return fmt.Errorf("backup filename required")
	}
	if err := mgr.Restore(rest[0]); err != nil {
		return err
	}
	fmt.Printf("Database restored (%s) from %s\n", mgr.Driver(), rest[0])
	return nil
}
