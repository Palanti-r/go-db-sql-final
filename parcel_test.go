package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/stretchr/testify/require"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
)
var store ParcelStore

func TestMain(m *testing.M) {
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		os.Exit(1)
	}
	defer db.Close()

	store = NewParcelStore(db)
	runTests := m.Run()
	os.Exit(runTests)
}

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func addParcel(t *testing.T, store ParcelStore, parcel Parcel) int {
	number, err := store.Add(parcel)
	require.NoError(t, err, "Ошибка при добавлении посылки")
	require.NotZero(t, number, "У добавленной посылки отсутствует идентификатор")
	return number
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	parcel := getTestParcel()
	number := addParcel(t, store, parcel)

	ans, err := store.Get(number)
	require.NoError(t, err, "Ошибка при получении посылки")
	assert.Equal(t, number, ans.Number, "Возвращается неверный number")

	parcel.Number = number
	assert.Equal(t, parcel, ans)

	err = store.Delete(number)
	require.NoError(t, err, "Ошибка при удалении посылки")
	ans, err = store.Get(number)
	require.ErrorIs(t, err, sql.ErrNoRows, "Неверная ошибка при получении удаленной посылки")
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	parcel := getTestParcel()
	number := addParcel(t, store, parcel)

	newAddress := "new test address"
	err := store.SetAddress(number, newAddress)
	require.NoError(t, err, "Ошибка при обновлении адреса")

	ans, err := store.Get(number)
	require.NoError(t, err, "Ошибка при получении посылки с обновленным адресом")
	require.Equal(t, newAddress, ans.Address, "Полученный адрес не совпадает с ожидаемым")
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	parcel := getTestParcel()
	number := addParcel(t, store, parcel)

	err := store.SetStatus(number, ParcelStatusSent)
	require.NoError(t, err, "Ошибка при обновлении статуса")

	ans, err := store.Get(number)
	require.NoError(t, err, "Ошибка при получении посылки с обновленным статусом")
	require.Equal(t, ParcelStatusSent, ans.Status, "Полученный статус не совпадает с ожидаемым")
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	for _, p := range parcels {
		p.Number = addParcel(t, store, p)
		parcelMap[p.Number] = p
	}

	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err, "Ошибка при получении посылок по клиенту")
	require.Equal(t, len(parcels), len(storedParcels),
		"Количество полученных посылок клиента не совпадает с ожидаемым")

	for _, storedParcel := range storedParcels {
		parcel, ok := parcelMap[storedParcel.Number]
		require.True(t, ok, fmt.Sprintf("Отсутствует посылка %d", storedParcel.Number))
		assert.Equal(t, parcel, storedParcel)
	}
}
