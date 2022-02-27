package kopia

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-logr/logr"
)

type backupSummary struct {
	ID          string    `json:"id"`
	Source      source    `json:"source"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"startTime"`
	EndTime     time.Time `json:"endTime"`
	RootEntry   rootEntry `json:"rootEntry"`
}
type source struct {
	Host     string `json:"host"`
	UserName string `json:"userName"`
	Path     string `json:"path"`
}
type errors struct {
	Path  string `json:"path"`
	Error string `json:"error"`
}
type summ struct {
	Size             int64     `json:"size"`
	Files            int       `json:"files"`
	Symlinks         int       `json:"symlinks"`
	Dirs             int       `json:"dirs"`
	MaxTime          time.Time `json:"maxTime"`
	NumFailed        int       `json:"numFailed"`
	NumIgnoredErrors int       `json:"numIgnoredErrors"`
	Errors           []errors  `json:"errors"`
}
type rootEntry struct {
	Name  string    `json:"name"`
	Type  string    `json:"type"`
	Mode  string    `json:"mode"`
	Mtime time.Time `json:"mtime"`
	UID   int       `json:"uid"`
	Gid   int       `json:"gid"`
	Obj   string    `json:"obj"`
	Summ  summ      `json:"summ"`
}

type kopiaStdoutParser struct {
	log     logr.Logger
	summary *backupSummary
}

func (k *kopiaStdoutParser) parseKopiaStdout(line string) {
	parsedLine := ""
	k.summary = &backupSummary{}

	// Kopia seems to print a carriage return if it's one of these
	// status messages. This kills the output on some terminals.
	if strings.Contains(line, "hashing") {
		parsedLine = trimFirstRune(line)
	} else if json.Unmarshal([]byte(line), k.summary) == nil { // check if the current line is the backup summary
		parsedLine = fmt.Sprintf("backup finished with %d errors", k.summary.RootEntry.Summ.NumFailed)
	} else {
		parsedLine = line
	}

	if strings.Contains(parsedLine, "ERROR") {
		k.log.Error(nil, parsedLine)
	} else {
		k.log.Info(parsedLine)
	}
}

func trimFirstRune(s string) string {
	_, i := utf8.DecodeRuneInString(s)
	return s[i:]
}
