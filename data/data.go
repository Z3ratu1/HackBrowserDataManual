package data

import (
	"HackBrowserDataManual/browser"
	"HackBrowserDataManual/item"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/gocarina/gocsv"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io"
	"os"
	"path/filepath"
	"time"
)

type IManager interface {
	Parse(masterKey []byte, inputFile string) error
	WriteData(browser *browser.Browser) error
}

type Manager struct {
	InnerData      any
	OutputFileName string
	OutputFormat   string
}

func (m *Manager) WriteData(b *browser.Browser) error {
	if m.OutputFileName == "" {
		m.OutputFileName = fmt.Sprintf("%s_%s_%d.%s", b.GetName(), b.Action, time.Now().Unix(), m.ext())
	}
	outputFile, err := m.createFile(m.OutputFileName)
	if err != nil {
		return err
	}

	log.Infof("Writing results to %s", m.OutputFileName)
	return m.write(m.InnerData, outputFile)
}

func (m *Manager) write(data any, writer io.Writer) error {
	switch m.OutputFormat {
	case item.Json:
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("  ", "  ")
		encoder.SetEscapeHTML(false)
		return encoder.Encode(data)
	default:
		gocsv.SetCSVWriter(func(w io.Writer) *gocsv.SafeCSVWriter {
			writer := csv.NewWriter(transform.NewWriter(w, unicode.UTF8BOM.NewEncoder()))
			writer.Comma = ','
			return gocsv.NewSafeCSVWriter(writer)
		})
		return gocsv.Marshal(data, writer)
	}
}

func (m *Manager) createFile(filename string) (*os.File, error) {
	var file *os.File
	var err error
	file, err = os.OpenFile(filepath.Clean(filename), os.O_TRUNC|os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (m *Manager) ext() string {
	if m.OutputFormat == item.Json {
		return "json"
	}
	return "csv"
}
