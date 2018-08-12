package storage

import (
	"time"

	"github.com/hashicorp/go-uuid"
	"github.com/jinzhu/gorm"
)

// Item represents a storage item instance.
type Item struct {
	Type        string    `gorm:"-" json:"type,omitempty"`
	ID          string    `gorm:"type:varchar(36);not null;primary_key" json:"id"`
	Name        string    `gorm:"type:varchar(128);not null;unique_index" json:"name"`
	Description *string   `gorm:"type:text" json:"description"`
	Created     time.Time `gorm:"not null;default:current_timestamp" json:"created"`
	Modified    time.Time `gorm:"not null;default:current_timestamp" json:"modified"`

	storage *Storage
}

// BeforeSave handles the ORM 'BeforeSave' callback.
func (i *Item) BeforeSave(scope *gorm.Scope) error {
	if i.ID == "" {
		id, err := uuid.GenerateUUID()
		if err != nil {
			return err
		}

		scope.SetColumn("ID", id)
	} else if _, err := uuid.ParseUUID(i.ID); err != nil {
		return ErrInvalidID
	}

	if !nameRegexp.MatchString(i.Name) {
		return ErrInvalidName
	}

	now := time.Now().UTC().Round(time.Second)

	if i.Created.IsZero() {
		scope.SetColumn("Created", now)
	}

	scope.SetColumn("Modified", now)

	// Ensure optional fields are null if empty
	if i.Description != nil && *i.Description == "" {
		scope.SetColumn("Description", nil)
	}

	return nil
}

// SetStorage sets the item internal storage reference.
func (i *Item) SetStorage(s *Storage) {
	i.storage = s
}
