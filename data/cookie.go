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

type CookieManager struct {
	*Manager
}

type CookieData struct {
	Host         string
	Path         string
	KeyName      string
	encryptValue []byte
	Value        string
	IsSecure     bool
	IsHTTPOnly   bool
	HasExpire    bool
	IsPersistent bool
	CreateDate   time.Time
	ExpireDate   time.Time
}

const (
	queryChromiumCookie = `SELECT name, encrypted_value, host_key, path, creation_utc, expires_utc, is_secure, is_httponly, has_expires, is_persistent FROM cookies`
)

func (cm *CookieManager) Parse(masterKey []byte, dbfile string) error {
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return err
	}
	defer db.Close()
	rows, err := db.Query(queryChromiumCookie)
	if err != nil {
		return err
	}
	defer rows.Close()
	log.Infof("reading sqlite db")
	cookies := make([]*CookieData, 0, 1024)
	for rows.Next() {
		var (
			key, host, path                               string
			isSecure, isHTTPOnly, hasExpire, isPersistent int
			createDate, expireDate                        int64
			value, encryptValue                           []byte
		)
		if err = rows.Scan(&key, &encryptValue, &host, &path, &createDate, &expireDate, &isSecure, &isHTTPOnly, &hasExpire, &isPersistent); err != nil {
			log.Warn(err)
		}

		cookie := CookieData{
			KeyName:      key,
			Host:         host,
			Path:         path,
			encryptValue: encryptValue,
			IsSecure:     utils.IntToBool(isSecure),
			IsHTTPOnly:   utils.IntToBool(isHTTPOnly),
			HasExpire:    utils.IntToBool(hasExpire),
			IsPersistent: utils.IntToBool(isPersistent),
			CreateDate:   utils.TimeEpoch(createDate),
			ExpireDate:   utils.TimeEpoch(expireDate),
		}
		if len(encryptValue) > 0 {
			if len(masterKey) == 0 {
				value, err = crypto.DPAPI(encryptValue)
			} else {
				value, err = crypto.DecryptPass(masterKey, encryptValue)
			}
			if err != nil {
				log.Error(err)
			}
		}
		cookie.Value = string(value)
		cookies = append(cookies, &cookie)
	}
	// sort with create date
	sort.Slice(cookies, func(i, j int) bool {
		return cookies[i].CreateDate.After(cookies[j].CreateDate)
	})
	cm.InnerData = &cookies
	return nil
}
