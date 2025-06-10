package main

import (
	"database/sql"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// randRange использует текущий источник случайных чисел для генерации
var randRange = rand.New(rand.NewSource(time.Now().UnixNano()))

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	number, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, number)

	// get
	storedParcel, err := store.Get(number)
	require.NoError(t, err)
	require.Equal(t, parcel.Client, storedParcel.Client)
	require.Equal(t, parcel.Status, storedParcel.Status)
	require.Equal(t, parcel.Address, storedParcel.Address)
	require.Equal(t, parcel.CreatedAt, storedParcel.CreatedAt)
	require.Equal(t, number, storedParcel.Number)

	// delete
	err = store.Delete(number)
	require.NoError(t, err)

	// check if parcel is deleted
	_, err = store.Get(number)
	require.Error(t, err)
	require.True(t, errors.Is(err, sql.ErrNoRows))
}

func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	number, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, number)

	// set address
	newAddress := "new test address"
	err = store.SetAddress(number, newAddress)
	require.NoError(t, err)

	// check
	updatedParcel, err := store.Get(number)
	require.NoError(t, err)
	require.Equal(t, newAddress, updatedParcel.Address)
}

func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	number, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, number)

	// set status
	newStatus := "delivered"
	err = store.SetStatus(number, newStatus)
	require.NoError(t, err)

	// check
	updatedParcel, err := store.Get(number)
	require.NoError(t, err)
	require.Equal(t, newStatus, updatedParcel.Status)
}

func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	for i := range parcels {
		parcels[i].Client = client
	}

	// add
	for i := range parcels {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotZero(t, id)

		// обновляем идентификатор добавленной посылки
		parcels[i].Number = id

		// сохраняем посылку в map
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Len(t, storedParcels, len(parcels))

	// check
	for _, parcel := range storedParcels {
		originalParcel, exists := parcelMap[parcel.Number]
		require.True(t, exists)
		require.Equal(t, originalParcel.Client, parcel.Client)
		require.Equal(t, originalParcel.Status, parcel.Status)
		require.Equal(t, originalParcel.Address, parcel.Address)
		require.Equal(t, originalParcel.CreatedAt, parcel.CreatedAt)
	}
}
