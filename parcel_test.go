package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestParcelStore_AddGetUpdateDelete(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)

	// уникальный client id, чтобы не зависеть от содержимого БД
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	client := 10_000 + rnd.Intn(100_000)

	// добавим несколько посылок в статусе registered
	addresses := []string{
		"Город А, Улица 1, д. 1",
		"Город Б, Улица 2, д. 2",
		"Город В, Улица 3, д. 3",
	}
	numbers := make([]int, 0, len(addresses))

	for _, addr := range addresses {
		p := Parcel{Client: client, Status: ParcelStatusRegistered, Address: addr, CreatedAt: time.Now().UTC().Format(time.RFC3339)}
		n, err := store.Add(p)
		require.NoError(t, err)
		numbers = append(numbers, n)
	}

	// проверим получение по client
	got, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Len(t, got, len(addresses))

	// сверим значения
	m := map[int]string{}
	for i, num := range numbers {
		m[num] = addresses[i]
	}

	for _, p := range got {
		require.Equal(t, client, p.Client)
		require.Equal(t, ParcelStatusRegistered, p.Status)
		require.NotEmpty(t, p.CreatedAt)
		addr, ok := m[p.Number]
		require.True(t, ok)
		require.Equal(t, addr, p.Address)
	}

	// SetAddress должен работать в registered
	newAddr := "Новый адрес, д. 7"
	require.NoError(t, store.SetAddress(numbers[0], newAddr))
	p0, err := store.GetByNumber(numbers[0])
	require.NoError(t, err)
	require.Equal(t, newAddr, p0.Address)

	// После смены статуса адрес менять нельзя
	require.NoError(t, store.SetStatus(numbers[0], ParcelStatusSent))
	err = store.SetAddress(numbers[0], "Адрес, который не должен сохраниться")
	require.Error(t, err)

	// Удаление разрешено только в registered
	// 1) Попытка удалить посылку в sent — ошибка
	err = store.Delete(numbers[0])
	require.Error(t, err)

	// 2) Добавим ещё одну registered и удалим её успешно
	p := Parcel{Client: client, Status: ParcelStatusRegistered, Address: "Удаляемая", CreatedAt: time.Now().UTC().Format(time.RFC3339)}
	numDel, err := store.Add(p)
	require.NoError(t, err)
	require.NoError(t, store.Delete(numDel))
	_, err = store.GetByNumber(numDel)
	require.Error(t, err) // должна отсутствовать
}
