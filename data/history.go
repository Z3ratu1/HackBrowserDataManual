package data

import (
	"HackBrowserDataManual/utils"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"sort"
	"time"
)

type HistoryManager struct {
	*Manager
}
type historyData struct {
	Title         string
	URL           string
	VisitCount    int
	LastVisitTime time.Time
}

const (
	queryChromiumHistory = `SELECT url, title, visit_count, last_visit_time FROM urls`
)

func (hm *HistoryManager) Parse(masterKey []byte, dbfile string) error {
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return err
	}
	defer db.Close()
	rows, err := db.Query(queryChromiumHistory)
	if err != nil {
		return err
	}
	historys := make([]*historyData, 0, 256)
	defer rows.Close()
	for rows.Next() {
		var (
			url, title    string
			visitCount    int
			lastVisitTime int64
		)
		if err := rows.Scan(&url, &title, &visitCount, &lastVisitTime); err != nil {
			log.Warn(err)
		}
		data := &historyData{
			URL:           url,
			Title:         title,
			VisitCount:    visitCount,
			LastVisitTime: utils.TimeEpoch(lastVisitTime),
		}
		historys = append(historys, data)
	}
	sort.Slice(historys, func(i, j int) bool {
		return historys[i].VisitCount > historys[j].VisitCount
	})
	hm.InnerData = historys

	return nil
}
