package backend

import (
	"time"

	"github.com/hashicorp/go-uuid"
	"github.com/jinzhu/gorm"
)

// Item represents a back-end item instance.
type Item struct {
	Type        string    `gorm:"-" json:"type,omitempty"`
	ID          string    `gorm:"type:varchar(36);not null;primary_key" json:"id"`
	Name        string    `gorm:"type:varchar(128);not null;unique_index" json:"name"`
	Description *string   `gorm:"type:text" json:"description"`
	Created     time.Time `gorm:"not null;default:current_timestamp" json:"created"`
	Modified    time.Time `gorm:"not null;default:current_timestamp" json:"modified"`

	backend *Backend `gorm:"-" json:"-"`
}

func (i *Item) BeforeSave(scope *gorm.Scope) error {
	var err error

	if !nameRegexp.MatchString(i.Name) {
		return ErrInvalidName
	}

	// Set default fields
	if i.ID == "" {
		i.ID, err = uuid.GenerateUUID()
		if err != nil {
			return err
		}
	} else if _, err := uuid.ParseUUID(uuid); err != nil {
		return ErrInvalidID
	}

	now := time.Now().UTC()

	if i.Created.IsZero() {
		i.Created = now
	}

	i.Modified = now

	// Ensure optional fields are null if empty
	if i.Description != nil && *i.Description == "" {
		i.Description = nil
	}

	return nil
}
