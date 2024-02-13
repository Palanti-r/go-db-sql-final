package main

import (
	"database/sql"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
	"time"

	_ "github.com/stretchr/testify/require"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
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

func connectDb(t *testing.T) (ParcelStore, Parcel) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err, "Ошибка подключения к БД")

	store := NewParcelStore(db)
	parcel := getTestParcel()

	return store, parcel
}

func addParcel(t *testing.T, store ParcelStore, parcel Parcel) int {
	number, err := store.Add(parcel)
	require.NoError(t, err, "Ошибка при добавлении посылки")
	require.NotZero(t, number, "У добавленной посылки отсутствует идентификатор")
	return number
}

func assertEqualParcels(t *testing.T, expected, actual Parcel) {
	assert.Equal(t, expected.Client, actual.Client,
		fmt.Sprintf("Возвращается неверный client для number=%d", actual.Number))
	assert.Equal(t, expected.Status, actual.Status,
		fmt.Sprintf("Возвращается неверный status для number=%d", actual.Number))
	assert.Equal(t, expected.Address, actual.Address,
		fmt.Sprintf("Возвращается неверный status для number=%d", actual.Number))
	assert.Equal(t, expected.CreatedAt, actual.CreatedAt,
		fmt.Sprintf("Возвращается неверный status для number=%d", actual.Number))
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	store, parcel := connectDb(t)
	number := addParcel(t, store, parcel)

	ans, err := store.Get(number)
	require.NoError(t, err, "Ошибка при получении посылки")
	assert.Equal(t, number, ans.Number, "Возвращается неверный number")
	assertEqualParcels(t, parcel, ans)

	err = store.Delete(number)
	require.NoError(t, err, "Ошибка при удалении посылки")
	ans, err = store.Get(number)
	require.ErrorIs(t, err, sql.ErrNoRows, "Неверная ошибка при получении удаленной посылки")
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	store, parcel := connectDb(t)
	number := addParcel(t, store, parcel)

	newAddress := "new test address"
	err := store.SetAddress(number, newAddress)
	require.NoError(t, err, "Ошибка при обновлении адреса")

	ans, err := store.Get(number)
	require.NoError(t, err, "Ошибка при получении посылки с обновленным адресом")
	require.Equal(t, ans.Address, newAddress, "Полученный адрес не совпадает с ожидаемым")
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	store, parcel := connectDb(t)
	number := addParcel(t, store, parcel)

	err := store.SetStatus(number, ParcelStatusSent)
	require.NoError(t, err, "Ошибка при обновлении статуса")

	ans, err := store.Get(number)
	require.NoError(t, err, "Ошибка при получении посылки с обновленным статусом")
	require.Equal(t, ParcelStatusSent, ans.Status, "Полученный статус не совпадает с ожидаемым")
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	store, _ := connectDb(t)

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
		assertEqualParcels(t, parcel, storedParcel)
	}
}
