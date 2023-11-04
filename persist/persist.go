package persist

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type PersistedResult struct {
	Page     string
	Children []string
}

type Persistence struct {
	f *os.File
}

func Open(name string) Persistence {
	dbName := fmt.Sprintf("mcrawl-result-%d-%s", time.Now().Unix(), name)
	dbName = strings.ReplaceAll(strings.ReplaceAll(dbName, ".", "-"), "/", "-")
	dbName += ".json"

	f, err := os.OpenFile(dbName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		slog.Warn("error openning file", "name", dbName)
	}

	return Persistence{
		f: f,
	}
}

func (p *Persistence) Filename() string {
	fullDbName, err := filepath.Abs(p.f.Name())
	if err == nil {
		return fullDbName
	}
	return p.f.Name()
}
func (p *Persistence) Close() {
	if err := p.f.Sync(); err != nil {
		log.Println("error on f.Sync():", err)
	}
	if err := p.f.Close(); err != nil {
		log.Println("error on f.Close():", err)
	}
	p.f = nil
}

func (p *Persistence) Save(page *url.URL, children []*url.URL) error {
	if p.f == nil {
		return errors.New("attempted to save into uninitialized persistence")
	}
	urls := make([]string, len(children))
	for i, child := range children {
		urls[i] = child.String()
	}

	result := PersistedResult{
		Page:     page.String(),
		Children: urls,
	}

	encoder := json.NewEncoder(p.f)
	if err := encoder.Encode(result); err != nil {
		return err
	}
	return nil
}

func (p *Persistence) load() ([]PersistedResult, error) {
	// open a new file descriptor
	file, err := os.Open(p.f.Name())
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var results []PersistedResult

	decoder := json.NewDecoder(file)
	for {
		var result PersistedResult
		if err := decoder.Decode(&result); err != nil {
			if err.Error() == "EOF" {
				break
			}
			slog.Info("found malformed data in the save file", "file", p.f.Name(), "err", err)
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}
