package item

import "os"

var homeDir, _ = os.UserHomeDir()

const (
	DefaultProfile = "/Default/"
)

const (
	Chrome = "chrome"
	Edge   = "edge"
)

const (
	ChromiumKey      = "Local State"
	ChromiumPassword = "Login Data"
	ChromiumCookie   = "Network/Cookies"
	ChromiumHistory  = "History"
)

const (
	Password = "password"
	Cookie   = "cookie"
	History  = "history"
)

const (
	Json = "json"
	CSV  = "csv"
)
