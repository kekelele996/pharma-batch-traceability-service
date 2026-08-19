package serialization

import (
	"fmt"
	"time"
)

const (
	StatusActive = "active"
	StatusSold   = "sold"
	StatusVoid   = "void"
)

type Serial struct {
	ID          string    `json:"id"`
	GTIN        string    `json:"gtin"`
	SerialNo    string    `json:"serial_no"`
	BatchID     string    `json:"batch_id"`
	Status      string    `json:"status"`
	ActivatedAt time.Time `json:"activated_at"`
}

func validStatus(s string) bool {
	switch s {
	case StatusActive, StatusSold, StatusVoid:
		return true
	}
	return false
}

func (s Serial) Code() string {
	return s.GTIN + s.SerialNo
}

func Validate(s Serial) error {
	if s.GTIN == "" || s.SerialNo == "" {
		return fmt.Errorf("serialization: gtin and serial number required")
	}
	if s.BatchID == "" {
		return fmt.Errorf("serialization: batch id required")
	}
	if !validStatus(s.Status) {
		return fmt.Errorf("serialization: unknown status %q", s.Status)
	}
	return nil
}
