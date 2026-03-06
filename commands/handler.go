package commands

import (
	"os"
	"os/exec"
	"strings"
)

var CurrentWorkDir, _ = os.Getwd()

func HandleCommand(input string) string {
	if strings.HasPrefix(input, "!ls") {
		return "ACTION_LIST_FILES"
	}
	if strings.HasPrefix(input, "!cd ") {
		return "ACTION_CHANGE_DIR"
	}
	if strings.HasPrefix(input, "!get") {
		return "ACTION_FETCH_FILE"
	}
	if strings.HasPrefix(input, "!find ") {
		return "ACTION_FIND_FILE"
	}
	if input == "!reset" {
		return "ACTION_HARD_RESET"
	}
	if strings.HasPrefix(input, "!cmd ") {
		return "ACTION_EXECUTE_CMD"
	}
	switch input {
	case "!status":
		return "Windows PC is Online\nOS: Windows 11\nStatus: Ready for commands"
	case "!lock":
		exec.Command("rundll32.exe", "user32.dll,LockWorkStation").Run()
		return "Windows PC is locked successfully"
	case "!screen":
		return "ACTION_SCREENSHOT"
	case "!shutdown":
		return "ACTION_SHUTDOWN"
	case "!restart":
		return "ACTION_RESTART"
	default:
		return "Invalid command"
	}
}
