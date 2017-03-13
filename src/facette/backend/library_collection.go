package backend

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"facette/mapper"
	"facette/template"

	"github.com/facette/sliceutil"
)

const collectionCacheTime = 30

// collectionsCache stores parent relations between collections in order to prevent requesting data from database
// upon each validation (see collectionCacheTime above)
var collectionsCache struct {
	updated time.Time
	data    map[string]string
}

// Collection represents a library collection item instance.
type Collection struct {
	Item
	Entries    CollectionEntryList `json:"entries,omitempty"`
	Link       *Collection         `orm:"type:varchar(36);foreign_key:ID" json:"-"`
	LinkID     string              `orm:"-" json:"link,omitempty"`
	Attributes mapper.Map          `json:"attributes,omitempty"`
	Alias      string              `orm:"type:varchar(128);unique" json:"alias,omitempty"`
	Options    mapper.Map          `json:"options,omitempty"`
	Parent     *Collection         `orm:"type:varchar(36);foreign_key:ID" json:"-"`
	ParentID   string              `orm:"-" json:"parent,omitempty"`
	Template   bool                `orm:"not_null;default:false" json:"template"`
}

// Validate checks whether or not the collection item instance is valid.
func (c Collection) Validate(backend *Backend) error {
	if err := c.Item.Validate(backend); err != nil {
		return err
	}

	if c.Alias != "" && !authorizedAliasChars.MatchString(c.Alias) {
		return ErrInvalidAlias
	}

	// Check for parent identifier
	if c.ParentID == "" {
		return nil
	} else if c.ID == c.ParentID {
		return ErrInvalidParent
	}

	// Get collections associations
	if collectionsCache.updated.Before(time.Now().Add(-1 * collectionCacheTime * time.Second)) {
		tx := backend.db.Begin()
		defer tx.Commit()

		q := tx.Select("id", "parent").Where("parent IS NOT NULL").From(Collection{})

		rows, err := q.Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		collectionsCache.data = map[string]string{}
		for rows.Next() {
			var collection Collection
			if err := q.Scan(rows, &collection).Error(); err != nil {
				return err
			}
			collectionsCache.data[collection.ID] = collection.ParentID
		}

		collectionsCache.updated = time.Now()
	}

	// Loop through collections for conflicting parenting
	parent := c.ParentID
	for {
		var ok bool

		parent, ok = collectionsCache.data[parent]
		if !ok {
			break
		} else if parent == c.ID {
			return ErrInvalidParent
		}
	}

	// Update cache entry for current collection
	collectionsCache.data[c.ID] = c.ParentID

	return nil
}

// CollectionEntryList represents a list of library collection entries.
type CollectionEntryList []CollectionEntry

// Value marshals the collection entries for compatibility with SQL drivers.
func (l CollectionEntryList) Value() (driver.Value, error) {
	data, err := json.Marshal(l)
	return data, err
}

// Scan unmarshals the collection entries retrieved from SQL drivers.
func (l *CollectionEntryList) Scan(v interface{}) error {
	return json.Unmarshal(v.([]byte), l)
}

// CollectionEntry represents a library collection entry instance.
type CollectionEntry struct {
	ID         string     `json:"id"`
	Attributes mapper.Map `json:"attributes,omitempty"`
	Options    mapper.Map `json:"options,omitempty"`
}

// Expand expands the collection template with attributes passed as parameter.
func (c *Collection) Expand(attrs mapper.Map, backend *Backend) error {
	var err error

	if title, ok := c.Options["title"].(string); ok {
		if c.Options["title"], err = template.Expand(title, attrs); err != nil {
			return err
		}
	}

	c.Attributes.Merge(attrs, true)
	c.Template = false

	// Fecth graph entries titles
	titles := map[string]string{}

	ids := []interface{}{}
	for _, entry := range c.Entries {
		if !sliceutil.Has(ids, entry.ID) {
			ids = append(ids, entry.ID)
		}
	}

	if len(ids) > 0 {
		tx := backend.db.Begin()
		defer tx.Commit()

		graphs := []Graph{}
		tx.Select("id", "options").Where("id IN (?)", ids).Find(&graphs)

		for _, graph := range graphs {
			if title, ok := graph.Options["title"].(string); ok {
				titles[graph.ID] = title
			}
		}
	}

	// Update collection entries with resolved data
	for i := range c.Entries {
		if title, ok := titles[c.Entries[i].ID]; ok {
			c.Entries[i].Attributes.Merge(c.Attributes, true)

			// Set title in options
			if c.Entries[i].Options == nil {
				c.Entries[i].Options = mapper.Map{}
			}

			if c.Entries[i].Options["title"], err = template.Expand(title, c.Entries[i].Attributes); err != nil {
				return err
			}
		}
	}

	return nil
}
