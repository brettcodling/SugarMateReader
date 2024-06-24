package notify

import "os/exec"

func Warning(title, context string) {
	exec.Command("notify-send", title, context, "--icon=warning", "--app-name=SugarMateReader").Run()
}
