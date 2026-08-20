package drug

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type Drug struct {
	ID             string    `json:"id"`
	Code           string    `json:"code"`
	GenericName    string    `json:"generic_name"`
	TradeName      string    `json:"trade_name"`
	DosageForm     string    `json:"dosage_form"`
	Strength       string    `json:"strength"`
	RxCategory     string    `json:"rx_category"`
	Storage        string    `json:"storage"`
	ManufacturerID string    `json:"manufacturer_id"`
	CreatedAt      time.Time `json:"created_at"`
}

var dosageForms = map[string]bool{
	"tablet": true, "capsule": true, "injection": true, "oral-liquid": true,
	"ointment": true, "granule": true, "powder": true, "aerosol": true,
}

var rxCategories = map[string]bool{"rx": true, "otc": true, "otc-a": true, "otc-b": true}

var storageConditions = map[string]bool{
	"room": true, "cool": true, "cold": true, "frozen": true, "light-proof": true,
}

var codePattern = regexp.MustCompile(`^[A-Z][0-9]{8,10}$`)

func Validate(d Drug) error {
	if d.Code == "" || !codePattern.MatchString(d.Code) {
		return fmt.Errorf("drug: invalid approval code %q", d.Code)
	}
	if d.GenericName == "" {
		return fmt.Errorf("drug: generic name required")
	}
	if !dosageForms[d.DosageForm] {
		return fmt.Errorf("drug: unknown dosage form %q", d.DosageForm)
	}
	if !rxCategories[d.RxCategory] {
		return fmt.Errorf("drug: unknown rx category %q", d.RxCategory)
	}
	if !storageConditions[d.Storage] {
		return fmt.Errorf("drug: unknown storage condition %q", d.Storage)
	}
	return nil
}

func (d *Drug) Normalize() {
	d.Code = strings.TrimSpace(d.Code)
	d.GenericName = strings.TrimSpace(d.GenericName)
	d.TradeName = strings.TrimSpace(d.TradeName)
	d.DosageForm = strings.ToLower(strings.TrimSpace(d.DosageForm))
	d.RxCategory = strings.ToLower(strings.TrimSpace(d.RxCategory))
	d.Storage = strings.ToLower(strings.TrimSpace(d.Storage))
	d.ManufacturerID = strings.TrimSpace(d.ManufacturerID)
}
