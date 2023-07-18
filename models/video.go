package models

import (
	"encoding/json"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/gofrs/uuid"
	"time"
)

type Video struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	Name       string     `json:"name" db:"name"`
	Processed  bool       `json:"processed" db:"processed"`
	UploadedAt nulls.Time `json:"uploaded" db:"uploaded_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

// String is not required by pop and may be deleted
func (v Video) String() string {
	jv, _ := json.Marshal(v)
	return string(jv)
}

// Videoes is not required by pop and may be deleted
type Videoes []Video

// String is not required by pop and may be deleted
func (v Videoes) String() string {
	jv, _ := json.Marshal(v)
	return string(jv)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (v *Video) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: v.Name, Name: "Name"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (v *Video) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (v *Video) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
