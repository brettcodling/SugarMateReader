package notify

import (
	"github.com/brettcodling/SugarMateReader/pkg/directory"
	"github.com/gen2brain/beeep"
)

// Warning creates a warning notification.
func Warning(title, context string) {
	beeep.Notify(title, context, directory.ConfigDir+"warning.png")
}
