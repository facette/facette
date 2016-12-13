package backend

// Unit represents a library unit item instance.
type Unit struct {
	Item
	Label string `orm:"type:varchar(32);not_null;unique" json:"label"`
}

// Validate checks whether or not the unit item instance is valid.
func (u Unit) Validate(backend *Backend) error {
	if err := u.Item.Validate(backend); err != nil {
		return err
	} else if u.Label == "" {
		return ErrInvalidUnit
	}

	return nil
}
