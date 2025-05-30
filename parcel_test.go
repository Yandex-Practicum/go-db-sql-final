package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    randRange.Intn(1000000),
		Status:    ParcelStatusRegistered,
		Address:   "Test Address",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// getTestStore создаёт ParcelStore с временной БД в памяти
func getTestStore(t *testing.T) ParcelStore {
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	require.NoError(t, err)

	// Удалить таблицу, если она уже есть
	_, err = db.Exec(`DROP TABLE IF EXISTS parcel`)
	require.NoError(t, err)

	// Создать таблицу заново
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

	// Add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// Get
	stored, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, parcel.Client, stored.Client)
	require.Equal(t, parcel.Status, stored.Status)
	require.Equal(t, parcel.Address, stored.Address)

	// Delete
	err = store.Delete(id)
	require.NoError(t, err)

	// Get again
	_, err = store.Get(id)
	require.Error(t, err)
}

func TestSetAddress(t *testing.T) {
	store := getTestStore(t)
	parcel := getTestParcel()

	// Add
	id, err := store.Add(parcel)
	require.NoError(t, err)

	// Set Address
	newAddress := "Updated Address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// Get
	stored, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, stored.Address)
}

func TestSetStatus(t *testing.T) {
	store := getTestStore(t)
	parcel := getTestParcel()

	// Add
	id, err := store.Add(parcel)
	require.NoError(t, err)

	// Set Status
	newStatus := ParcelStatusSent
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)

	// Get
	stored, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newStatus, stored.Status)
}

func TestGetByClient(t *testing.T) {
	store := getTestStore(t)
	clientID := randRange.Intn(1000000)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}

	for i := range parcels {
		parcels[i].Client = clientID
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		parcels[i].Number = id
	}

	stored, err := store.GetByClient(clientID)
	require.NoError(t, err)
	require.Len(t, stored, len(parcels))

	// Проверим, что каждая посылка есть в ответе
	for _, original := range parcels {
		var found bool
		for _, s := range stored {
			if s.Number == original.Number {
				require.Equal(t, original.Address, s.Address)
				require.Equal(t, original.Status, s.Status)
				require.Equal(t, original.Client, s.Client)
				found = true
				break
			}
		}
		require.True(t, found, "Посылка %d не найдена", original.Number)
	}
}
