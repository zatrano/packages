package schedule

import (
	"fmt"
	"time"

	"github.com/zatrano/framework/v2/bootstrap/addons"
	"github.com/zatrano/framework/v2/contracts"
	"github.com/zatrano/packages/bootutil"
)

func Commands(app contracts.App) []addons.CLICommand {
	return bootutil.CLI(
		&ScheduleRunCommand{app: app},
		&ScheduleListCommand{app: app},
	)
}

type ScheduleRunCommand struct {
	app contracts.App
}

func (c *ScheduleRunCommand) Name() string        { return "schedule:run" }
func (c *ScheduleRunCommand) Description() string { return "Run the scheduled commands due now" }
func (c *ScheduleRunCommand) Handle(args []string) error {
	if err := c.app.Bootstrap(); err != nil {
		return err
	}
	sched := From(c.app)
	if sched == nil {
		return fmt.Errorf("scheduler is not bound; import github.com/zatrano/packages/schedule")
	}
	errs := sched.RunDue(time.Now())
	if len(errs) == 0 {
		fmt.Println("No scheduled commands are ready to run, or all completed successfully.")
		return nil
	}
	for _, err := range errs {
		fmt.Printf("Scheduled event failed: %v\n", err)
	}
	return errs[0]
}

type ScheduleListCommand struct {
	app contracts.App
}

func (c *ScheduleListCommand) Name() string        { return "schedule:list" }
func (c *ScheduleListCommand) Description() string { return "List all scheduled events" }
func (c *ScheduleListCommand) Handle(args []string) error {
	if err := c.app.Bootstrap(); err != nil {
		return err
	}
	sched := From(c.app)
	if sched == nil {
		return fmt.Errorf("scheduler is not bound; import github.com/zatrano/packages/schedule")
	}
	events := sched.Events()
	if len(events) == 0 {
		fmt.Println("No scheduled events defined.")
		return nil
	}
	for _, event := range events {
		fmt.Printf("%s\t%s\n", event.DisplayName(), event.Expression())
	}
	return nil
}
