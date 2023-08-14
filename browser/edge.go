package browser

import (
	"HackBrowserDataManual/item"
	"os/exec"
	"runtime"
)

type EdgeUtil struct {
}

func (e *EdgeUtil) getUserDir() string {
	return item.EdgeUserDataPath
}

func (e *EdgeUtil) findBinaryPath() string {
	// 真的会有windows外的os用edge吗。。。先乱写
	var locations []string
	switch runtime.GOOS {
	case "darwin":
		locations = []string{
			// Mac
			"edge",
			"msedge",
		}
	case "windows":
		locations = []string{
			// Windows
			`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
			`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
			"msedge.exe",
		}
	default:
		locations = []string{
			// Unix-like
			"edge",
			"msedge",
		}
	}

	for _, path := range locations {
		found, err := exec.LookPath(path)
		if err == nil {
			return found
		}
	}
	// Fall back to something simple and sensible, to give a useful error
	// message.
	return "msedge"
}
