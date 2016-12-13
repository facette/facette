package backend

// Scale represents a library scale item instance.
type Scale struct {
	Item
	Value float64 `orm:"not_null;unique" json:"value"`
}

// Validate checks whether or not the scale item instance is valid.
func (s Scale) Validate(backend *Backend) error {
	if err := s.Item.Validate(backend); err != nil {
		return err
	} else if s.Value == 0 {
		return ErrInvalidScale
	}

	return nil
}
