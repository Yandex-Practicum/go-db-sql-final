package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", ":memory:") // настройте подключение к БД
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE parcel (
	   number INTEGER PRIMARY KEY AUTOINCREMENT,
	   client INTEGER,
	   status TEXT,
	   address TEXT,
	   created_at TEXT
	)`)
	require.NoError(t, err)

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)
	parcel.Number = id

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	storedParcel, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, parcel, storedParcel)

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(id)
	require.NoError(t, err)

	_, err = store.Get(id)
	require.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", ":memory:") // настройте подключение к БД
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE parcel (
	number INTEGER PRIMARY KEY AUTOINCREMENT,
	client INTEGER,
	status TEXT,
	address TEXT,
	created_at TEXT
	)`)

	require.NoError(t, err)

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	updateParcel, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, newAddress, updateParcel.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", ":memory:") // настройте подключение к БД
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE parcel (
	number INTEGER PRIMARY KEY AUTOINCREMENT,
	client INTEGER,
	status TEXT,
	address TEXT,
	created_at TEXT
	)`)
	require.NoError(t, err)

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.NotZero(t, id)

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	err = store.SetStatus(id, ParcelStatusSent)
	require.NoError(t, err)

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	updateParcel, err := store.Get(id)
	require.NoError(t, err)
	assert.Equal(t, ParcelStatusSent, updateParcel.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", ":memory:") // настройте подключение к БД
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE parcel (
	number INTEGER PRIMARY KEY AUTOINCREMENT,
	client INTEGER,
	status TEXT,
	address TEXT,
	created_at TEXT
	)`)
	require.NoError(t, err)

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i]) // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		require.NoError(t, err)
		require.NotZero(t, id)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	assert.Len(t, storedParcels, len(parcels))
	// получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных

	// check
	for _, parcel := range storedParcels {
		expParcel, ok := parcelMap[parcel.Number]
		assert.True(t, ok)
		require.Equal(t, expParcel, parcel)

		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
	}
}

func TestParcelService_Register(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE parcel (
        number INTEGER PRIMARY KEY AUTOINCREMENT,
        client INTEGER,
        status TEXT,
        address TEXT,
        created_at TEXT
    )`)
	require.NoError(t, err)

	store := NewParcelStore(db)
	service := NewParcelService(store)

	parcel, err := service.Register(123, "Test Address")
	require.NoError(t, err)
	assert.Equal(t, ParcelStatusRegistered, parcel.Status)
	assert.Equal(t, 123, parcel.Client)
	assert.Equal(t, "Test Address", parcel.Address)
	assert.NotEmpty(t, parcel.CreatedAt)
}

func TestParcelService_NextStatus_Flow(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE parcel (
        number INTEGER PRIMARY KEY AUTOINCREMENT,
        client INTEGER,
        status TEXT,
        address TEXT,
        created_at TEXT
    )`)
	require.NoError(t, err)

	store := NewParcelStore(db)
	service := NewParcelService(store)

	parcel, err := service.Register(123, "Test Address")
	require.NoError(t, err)

	err = service.NextStatus(parcel.Number)
	require.NoError(t, err)

	updated, err := store.Get(parcel.Number)
	require.NoError(t, err)
	assert.Equal(t, ParcelStatusSent, updated.Status)

	err = service.NextStatus(parcel.Number)
	require.NoError(t, err)

	updated, err = store.Get(parcel.Number)
	require.NoError(t, err)
	assert.Equal(t, ParcelStatusDelivered, updated.Status)

	err = service.NextStatus(parcel.Number)
	require.NoError(t, err)
}

func TestParcelService_ChangeAddress_OnlyForRegistered(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE parcel (
        number INTEGER PRIMARY KEY AUTOINCREMENT,
        client INTEGER,
        status TEXT,
        address TEXT,
        created_at TEXT
    )`)
	require.NoError(t, err)

	store := NewParcelStore(db)
	service := NewParcelService(store)

	parcel, err := service.Register(123, "Original Address")
	require.NoError(t, err)

	err = service.ChangeAddress(parcel.Number, "New Address")
	require.NoError(t, err)

	updated, err := store.Get(parcel.Number)
	require.NoError(t, err)
	assert.Equal(t, "New Address", updated.Address)

	err = service.NextStatus(parcel.Number)
	require.NoError(t, err)

	err = service.ChangeAddress(parcel.Number, "Another Address")
	require.Error(t, err)
}

func TestParcelService_Delete_OnlyForRegistered(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE parcel (
        number INTEGER PRIMARY KEY AUTOINCREMENT,
        client INTEGER,
        status TEXT,
        address TEXT,
        created_at TEXT
    )`)
	require.NoError(t, err)

	store := NewParcelStore(db)
	service := NewParcelService(store)

	parcel, err := service.Register(123, "Test Address")
	require.NoError(t, err)

	err = service.Delete(parcel.Number)
	require.NoError(t, err)

	_, err = store.Get(parcel.Number)
	require.Error(t, err)

	parcel2, err := service.Register(123, "Test Address 2")
	require.NoError(t, err)

	err = service.NextStatus(parcel2.Number)
	require.NoError(t, err)

	err = service.Delete(parcel2.Number)
	require.Error(t, err)
}

func TestParcelService_PrintClientParcels(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE parcel (
        number INTEGER PRIMARY KEY AUTOINCREMENT,
        client INTEGER,
        status TEXT,
        address TEXT,
        created_at TEXT
    )`)
	require.NoError(t, err)

	store := NewParcelStore(db)
	service := NewParcelService(store)

	clientID := 999

	_, err = service.Register(clientID, "Address 1")
	require.NoError(t, err)

	_, err = service.Register(clientID, "Address 2")
	require.NoError(t, err)

	err = service.PrintClientParcels(clientID)
	require.NoError(t, err)
}

func TestParcelService_ChangeAddress_NonExistentParcel(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE parcel (
        number INTEGER PRIMARY KEY AUTOINCREMENT,
        client INTEGER,
        status TEXT,
        address TEXT,
        created_at TEXT
    )`)
	require.NoError(t, err)

	store := NewParcelStore(db)
	service := NewParcelService(store)

	err = service.ChangeAddress(9999, "New Address")
	require.Error(t, err)
}

func TestParcelService_NextStatus_NonExistentParcel(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE parcel (
        number INTEGER PRIMARY KEY AUTOINCREMENT,
        client INTEGER,
        status TEXT,
        address TEXT,
        created_at TEXT
    )`)
	require.NoError(t, err)

	store := NewParcelStore(db)
	service := NewParcelService(store)

	err = service.NextStatus(9999)
	require.Error(t, err)
}
