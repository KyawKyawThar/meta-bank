package db

import (
	"context"
	"github.com/HL/meta-bank/util"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func CreateRandomEntry(t *testing.T, acc Account) Entry {

	arg := CreateEntryParams{
		AccountID: acc.ID,
		Amount:    util.RandomAmount(),
	}

	entry, err := testStore.CreateEntry(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, entry)

	require.Equal(t, arg.AccountID, entry.AccountID)
	require.Equal(t, arg.Amount, entry.Amount)

	require.NotZero(t, entry.Amount)
	require.NotZero(t, entry.ID)
	require.NotZero(t, entry.CreatedAt)

	return entry
}

func TestCreateEntry(t *testing.T) {
	acc := createRandomAccount(t)
	CreateRandomEntry(t, acc)
}

func TestGetEntry(t *testing.T) {
	acc := createRandomAccount(t)

	entry := CreateRandomEntry(t, acc)

	entry1, err := testStore.GetEntry(context.Background(), entry.ID)
	require.NoError(t, err)
	require.NotEmpty(t, entry1)

	require.WithinDuration(t, entry.CreatedAt, entry1.CreatedAt, time.Second)

	require.Equal(t, entry.ID, entry1.ID)
	require.Equal(t, entry.AccountID, entry1.AccountID)
	require.Equal(t, entry.Amount, entry1.Amount)

}

func TestListEntries(t *testing.T) {
	acc := createRandomAccount(t)

	for i := 0; i < 10; i++ {
		CreateRandomEntry(t, acc)

	}

	arg := ListEntriesParams{
		AccountID: acc.ID,
		Limit:     5,
		Offset:    5,
	}

	entries, err := testStore.ListEntries(context.Background(), arg)

	require.NoError(t, err)
	require.NotEmpty(t, entries)
	require.Len(t, entries, 5)

	for _, entry := range entries {
		require.NotEmpty(t, entry)
		require.Equal(t, acc.ID, entry.AccountID)
	}
}
