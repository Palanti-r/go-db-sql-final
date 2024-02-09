package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
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
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	number, err := store.Add(parcel)
	require.NoError(t, err)
	assert.NotEmpty(t, number)

	stored, err := store.Get(number)
	require.NoError(t, err)
	assert.Equal(t, parcel.Client, stored.Client)
	assert.Equal(t, parcel.Status, stored.Status)
	assert.Equal(t, parcel.Address, stored.Address)
	assert.Equal(t, parcel.CreatedAt, stored.CreatedAt)

	err = store.Delete(number)
	require.NoError(t, err)
	stored, err = store.Get(number)
	require.Equal(t, sql.ErrNoRows, err)

}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	number, err := store.Add(parcel)
	require.NoError(t, err)
	assert.NotEmpty(t, number)

	newAddress := "new test address"
	store.SetAddress(number, newAddress)

	stored, err := store.Get(number)
	require.Equal(t, newAddress, stored.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	number, err := store.Add(parcel)
	require.NoError(t, err)
	assert.NotEmpty(t, number)

	newStatus := ParcelStatusSent
	store.SetStatus(number, newStatus)

	stored, err := store.Get(number)
	require.NoError(t, err)
	assert.Equal(t, newStatus, stored.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
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

	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	for i := 0; i < len(parcels); i++ {
		number, err := store.Add(parcels[i])
		require.NoError(t, err)
		assert.NotEmpty(t, number)
		parcels[i].Number = number
		parcelMap[number] = parcels[i]
	}

	storedParcels, err := store.GetByClient(client) // получите список посылок по идентификатору клиента, сохранённого в переменной client
	require.NoError(t, err)
	assert.Len(t, parcels, len(storedParcels))

	for _, parcel := range storedParcels {
		p, ok := parcelMap[parcel.Number]
		require.True(t, ok)
		assert.Equal(t, p.Client, parcel.Client)
		assert.Equal(t, p.Status, parcel.Status)
		assert.Equal(t, p.Address, parcel.Address)
		assert.Equal(t, p.CreatedAt, parcel.CreatedAt)

	}
}
