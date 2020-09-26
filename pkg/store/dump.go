// Copyright (c) 2020, The Facette Authors
//
// Licensed under the terms of the BSD 3-Clause License; a copy of the license
// is available at: https://opensource.org/licenses/BSD-3-Clause

package store

import (
	"context"

	"gorm.io/gorm"

	"facette.io/facette/pkg/api"

	"facette.io/facette/pkg/store/internal/types"
)

// Dump dumps back-end storage objects.
func (s *Store) Dump(ch chan<- api.Object) error {
	defer close(ch)

	tx := s.db.Begin()
	defer tx.Commit()

	var err error

	for _, objects := range []types.ObjectList{
		&types.ChartList{},
		&types.DashboardList{},
		&types.ProviderList{},
	} {
		err = s.dump(tx, objects, ch)
		if err != nil {
			return err
		}
	}

	// TODO: check for dashboards parents order

	return nil
}

func (s *Store) dump(db *gorm.DB, objects types.ObjectList, ch chan<- api.Object) error {
	q := db.Model(objects)

	switch objects.(type) {
	case *types.ChartList:
		q = q.Order("template DESC").Order("created_at")

	case *types.DashboardList:
		q = q.Order("template DESC").Order("parent_id").Order("created_at")

	case *types.ProviderList:
		q = q.Order("created_at")
	}

	rows, err := q.Rows()
	if err != nil {
		return err
	}

	for rows.Next() {
		v := objects.New()

		err = db.ScanRows(rows, v)
		if err != nil {
			return err
		}

		obj, err := types.ToAPI(v)
		if err != nil {
			return err
		}

		ch <- obj
	}

	return nil
}

// CancelRestore cancels back-end storage objects restauration.
func (s *Store) CancelRestore() {
	if s.restoreCancel != nil {
		s.restoreCancel()
	}
}

// Restore restores back-end storage data from objects.
func (s *Store) Restore(ctx context.Context, ch <-chan api.Object) error {
	var (
		restoreCtx context.Context
		err        error
	)

	restoreCtx, s.restoreCancel = context.WithCancel(ctx)
	defer s.restoreCancel()

	tx := s.db.Begin().WithContext(restoreCtx)

	for _, table := range []string{
		"charts",
		"dashboards",
		"providers",
	} {
		err = tx.Exec(s.driver.TruncateStmt(table)).Error
		if err != nil {
			goto stop
		}
	}

	for obj := range ch {
		var v types.Object

		v, err = types.FromAPI(obj)
		if err != nil {
			break
		}

		err = tx.Create(v).Error
		if err != nil {
			break
		}
	}

stop:
	if err != nil {
		tx.Rollback()
		return s.driver.Error(err)
	}

	return tx.Commit().Error
}
