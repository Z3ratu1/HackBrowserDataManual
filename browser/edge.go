package browser

import "HackBrowserDataManual/item"

type EdgeUtil struct {
}

func (e *EdgeUtil) getUserDir() string {
	return item.EdgeUserDataPath
}

func (e *EdgeUtil) findBinaryPath() string {
	// TODO impl
	panic("impl")
}
