package extid_test

import (
	"context"
	"math"
	"os"
	"testing"

	"github.com/jackc/go-extid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var defaultConnTestRunner pgxtest.ConnTestRunner

func init() {
	defaultConnTestRunner = pgxtest.DefaultConnTestRunner()
	defaultConnTestRunner.CreateConfig = func(ctx context.Context, t testing.TB) *pgx.ConnConfig {
		config, err := pgx.ParseConfig(os.Getenv("EXTID_TEST_DATABASE"))
		require.NoError(t, err)
		return config
	}
}

func TestEncodeKnownValues(t *testing.T) {
	prefix := "user"
	key := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}

	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, t testing.TB, conn *pgx.Conn) {
		tx, err := conn.Begin(ctx)
		require.NoError(t, err)
		defer tx.Rollback(ctx)

		_, err = tx.Exec(ctx, "insert into extid_types (prefix, secret_key) values ($1, $2)", prefix, key)
		require.NoError(t, err)

		for i, tt := range []struct {
			id  int64
			xid string
		}{
			{id: math.MinInt64, xid: "user_4399572cd6ea5341b8d35876a7098af7"},
			{id: -1, xid: "user_25d4e948bd5e1296afc0bf87095a7248"},
			{id: 0, xid: "user_c6a13b37878f5b826f4f8162a1c8d879"},
			{id: 1, xid: "user_13189a6ae4ab07ae70a3aabd30be99de"},
			{id: math.MaxInt64, xid: "user_edc17bee21fb24e211e6419412e1c32e"},
		} {
			var xid string
			err = tx.QueryRow(ctx, "select encode_extid($1, $2)", prefix, tt.id).Scan(&xid)
			assert.NoErrorf(t, err, "%d", i)
			assert.Equalf(t, tt.xid, xid, "%d", i)
		}
	})

}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	prefix := "user"
	key := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}

	et, err := extid.NewType(prefix, key)
	require.NoError(t, err)

	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, t testing.TB, conn *pgx.Conn) {
		tx, err := conn.Begin(ctx)
		require.NoError(t, err)
		defer tx.Rollback(ctx)

		_, err = tx.Exec(ctx, "insert into extid_types (prefix, secret_key) values ($1, $2)", prefix, key)
		require.NoError(t, err)

		for id := int64(-1000); id < 1000; id++ {
			var xid string
			err = tx.QueryRow(ctx, "select encode_extid($1, $2)", prefix, id).Scan(&xid)
			assert.NoErrorf(t, err, "id: %d", id)

			// Ensure same encoding as Go implementation
			assert.Equalf(t, et.Encode(id), xid, "id: %d", id)

			var roundTripID int64
			err = tx.QueryRow(ctx, "select decode_extid($1)", xid).Scan(&roundTripID)
			assert.NoErrorf(t, err, "id: %d", id)

			assert.Equalf(t, id, roundTripID, "id: %d", id)
		}
	})
}

func FuzzEncodeDecode(f *testing.F) {
	prefix := "user"
	key := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}

	testcases := []struct {
		id int64
	}{
		{id: math.MinInt64},
		{id: -1},
		{id: 0},
		{id: 1},
		{id: math.MaxInt64},
	}
	for _, tc := range testcases {
		f.Add(tc.id)
	}

	f.Fuzz(func(t *testing.T, id int64) {
		et, err := extid.NewType(prefix, key)
		require.NoError(t, err)

		defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, t testing.TB, conn *pgx.Conn) {
			tx, err := conn.Begin(ctx)
			require.NoError(t, err)
			defer tx.Rollback(ctx)

			_, err = tx.Exec(ctx, "insert into extid_types (prefix, secret_key) values ($1, $2)", prefix, key)
			require.NoError(t, err)

			var xid string
			err = tx.QueryRow(ctx, "select encode_extid($1, $2)", prefix, id).Scan(&xid)
			assert.NoErrorf(t, err, "id: %d", id)

			// Ensure same encoding as Go implementation
			assert.Equalf(t, et.Encode(id), xid, "id: %d", id)

			var roundTripID int64
			err = tx.QueryRow(ctx, "select decode_extid($1)", xid).Scan(&roundTripID)
			assert.NoErrorf(t, err, "id: %d", id)

			assert.Equalf(t, id, roundTripID, "id: %d", id)
		})
	})
}
