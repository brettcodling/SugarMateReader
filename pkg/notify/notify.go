package notify

import "os/exec"

// Warning creates a warning notification.
func Warning(title, context string) {
	exec.Command("notify-send", title, context, "--icon=warning", "--app-name=SugarMateReader").Run()
}
