package queue

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zatrano/framework/v2/bootstrap/addons"
	"github.com/zatrano/framework/v2/contracts"
	"github.com/zatrano/packages/bootutil"
)

func Commands(app contracts.App) []addons.CLICommand {
	return bootutil.CLI(
		&QueueWorkCommand{app: app},
		&MakeJobCommand{app: app},
	)
}

type QueueWorkCommand struct {
	app contracts.App
}

func (c *QueueWorkCommand) Name() string        { return "queue:work" }
func (c *QueueWorkCommand) Description() string { return "Start processing jobs on the queue" }
func (c *QueueWorkCommand) Handle(args []string) error {
	if err := c.app.Bootstrap(); err != nil {
		return err
	}
	once := false
	queueName := ""
	for i := 0; i < len(args); i++ {
		if args[i] == "--once" {
			once = true
		}
		if (args[i] == "--queue" || args[i] == "-q") && i+1 < len(args) {
			queueName = args[i+1]
			i++
		}
	}

	mgr := From(c.app)
	if mgr == nil {
		return fmt.Errorf("queue is not bound; import github.com/zatrano/packages/queue")
	}

	for {
		err := mgr.Work(queueName)
		if err != nil {
			if strings.Contains(err.Error(), "no rows") || err.Error() == "sql: no rows in result set" {
				if once {
					fmt.Println("No jobs available.")
					return nil
				}
				time.Sleep(time.Second)
				continue
			}
			fmt.Printf("Job failed: %v\n", err)
			if once {
				return err
			}
			continue
		}
		fmt.Println("Processed a job.")
		if once {
			return nil
		}
	}
}

type MakeJobCommand struct {
	app contracts.App
}

func (c *MakeJobCommand) Name() string        { return "make:job" }
func (c *MakeJobCommand) Description() string { return "Create a new job class" }
func (c *MakeJobCommand) Handle(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("job name required")
	}
	name := args[0]
	dir := c.app.BasePath("app", "jobs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, bootutil.ToSnake(name)+".go")
	content := fmt.Sprintf(`package jobs

import "fmt"

const %sName = "%s"

// Tries is the suggested max attempts for this job.
const %sTries = 3

func Handle%s(payload map[string]any) error {
	fmt.Printf("Handling %s: %%v\n", payload)
	return nil
}
`, name, bootutil.ToSnake(name), name, name, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	fmt.Printf("Job created: %s\n", path)
	return nil
}
