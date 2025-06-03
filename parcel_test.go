package main

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	schema := `
	CREATE TABLE parcel (
		number INTEGER PRIMARY KEY AUTOINCREMENT,
		client INTEGER,
		status TEXT,
		address TEXT,
		created_at TEXT
	);`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}

func TestAddGetDelete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewParcelStore(db)

	parcel := Parcel{
		Client:    1,
		Status:    ParcelStatusRegistered,
		Address:   "Test address",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	got, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, parcel.Client, got.Client)
	require.Equal(t, parcel.Status, got.Status)
	require.Equal(t, parcel.Address, got.Address)

	err = store.Delete(id)
	require.NoError(t, err)

	_, err = store.Get(id)
	require.Error(t, err)
}

func TestSetStatusAndAddress(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewParcelStore(db)

	parcel := Parcel{
		Client:    2,
		Status:    ParcelStatusRegistered,
		Address:   "Initial address",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	id, err := store.Add(parcel)
	require.NoError(t, err)

	newAddress := "Updated address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	p, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, p.Address)

	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)

	p, err = store.Get(id)
	require.NoError(t, err)
	require.Equal(t, ParcelStatusSent, p.Status)
}

func TestGetByClient(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewParcelStore(db)

	parcel := Parcel{
		Client:    42,
		Status:    ParcelStatusRegistered,
		Address:   "Client address",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	_, err := store.Add(parcel)
	require.NoError(t, err)

	parcels, err := store.GetByClient(42)
	require.NoError(t, err)
	require.Len(t, parcels, 1)
	require.Equal(t, 42, parcels[0].Client)
}
