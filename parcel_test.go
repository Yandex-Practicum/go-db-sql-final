package main

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getTestParcel возвращает тестовую посылку с фиксированными значениями
func getTestParcel() Parcel {
	return Parcel{
		Client:    42,
		Status:    ParcelStatusRegistered,
		Address:   "Test Address",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// getTestStore создаёт ParcelStore с временной БД в памяти
func getTestStore(t *testing.T) ParcelStore {
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	require.NoError(t, err)

	_, err = db.Exec(`DROP TABLE IF EXISTS parcel`)
	require.NoError(t, err)

	_, err = db.Exec(`CREATE TABLE parcel (
        number INTEGER PRIMARY KEY AUTOINCREMENT,
        client INTEGER NOT NULL,
        status TEXT NOT NULL,
        address TEXT NOT NULL,
        created_at TEXT NOT NULL
    )`)
	require.NoError(t, err)

	return NewParcelStore(db)
}

func TestAddGetDelete(t *testing.T) {
	store := getTestStore(t)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	stored, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, parcel.Client, stored.Client)
	assert.Equal(t, parcel.Status, stored.Status)
	assert.Equal(t, parcel.Address, stored.Address)

	err = store.Delete(id)
	require.NoError(t, err)

	_, err = store.Get(id)
	require.Error(t, err)
}

func TestSetAddress(t *testing.T) {
	store := getTestStore(t)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)

	newAddress := "Updated Address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	stored, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, newAddress, stored.Address)
}

func TestSetStatus(t *testing.T) {
	store := getTestStore(t)
	parcel := getTestParcel()

	id, err := store.Add(parcel)
	require.NoError(t, err)

	newStatus := ParcelStatusSent
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)

	stored, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, newStatus, stored.Status)
}

func TestGetByClient(t *testing.T) {
	store := getTestStore(t)
	clientID := 42

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}

	parcelMap := map[int]Parcel{}
	for i := range parcels {
		parcels[i].Client = clientID
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	stored, err := store.GetByClient(clientID)
	require.NoError(t, err)
	require.Len(t, stored, len(parcels))

	for _, s := range stored {
		original := parcelMap[s.Number]
		assert.Equal(t, original.Address, s.Address)
		assert.Equal(t, original.Status, s.Status)
		assert.Equal(t, original.Client, s.Client)
	}
}
