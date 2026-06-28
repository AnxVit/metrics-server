package audit

import (
	"encoding/json"
	"os"
	"sync"
)

var _ Audit = &AuditFile{}

const (
	auditFileName = "audit-file"
)

type AuditFile struct {
	filename string
	mu       sync.Mutex
}

func NewAuditFile(filename string) *AuditFile {
	return &AuditFile{
		filename: filename,
	}
}

func (a *AuditFile) GetID() string {
	return auditFileName
}

func (a *AuditFile) Update(event Event) {
	a.saveEventToFile(event)
}

func (a *AuditFile) saveEventToFile(event Event) error {
	file, err := os.OpenFile(a.filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	bytesEvent, err := json.Marshal(event)
	if err != nil {
		return err
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	_, err = file.Write(bytesEvent)
	if err != nil {
		return err
	}
	file.WriteString("\n")
	return nil
}
