package backend

import "time"

// Validator represents an item validator interface.
type Validator interface {
	Validate(*Backend) error
}

// Item represents a backend item instance.
type Item struct {
	ID          string     `orm:"type:varchar(36);not_null;primary_key" json:"id"`
	Name        string     `orm:"type:varchar(255);not_null;unique" json:"name"`
	Description *string    `json:"description"`
	Created     time.Time  `orm:"not_null" json:"created"`
	Modified    *time.Time `json:"modified"`
}

// Validate checks whether or not the item instance is valid.
func (i *Item) Validate(backend *Backend) error {
	if i.Name == "" {
		return ErrInvalidName
	}

	return nil
}

// TypedItem represents a typed backend item instance.
type TypedItem struct {
	Item
	Type string `json:"type"`
}
