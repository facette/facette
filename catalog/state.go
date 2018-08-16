package catalog

import (
	"encoding/gob"
	"io"
	"os"
)

// Dump dumps the catalog records to a file.
func (c *Catalog) Dump(path string) error {
	var records []*Record

	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0640)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, o := range c.Origins {
		for _, s := range o.Sources {
			for _, m := range s.Metrics {
				records = append(records, &Record{
					Origin:     m.Origin().Name,
					Source:     m.Source().Name,
					Metric:     m.Name,
					Attributes: m.Attributes,
				})
			}
		}
	}

	return gob.NewEncoder(f).Encode(records)
}

// Restore restores the catalog records from a file.
func (c *Catalog) Restore(path string) error {
	var records []*Record

	f, err := os.OpenFile(path, os.O_RDONLY, 0640)
	if err != nil {
		return err
	}
	defer f.Close()

	err = gob.NewDecoder(f).Decode(&records)
	if err == io.EOF {
		return nil
	} else if err != nil {
		return err
	}

	for _, r := range records {
		c.Insert(r)
	}

	return nil
}
