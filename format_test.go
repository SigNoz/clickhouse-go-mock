package mockhouse

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestQueryFormat(t *testing.T) {
	t.Parallel()
	mock, err := NewClickHouseNative(nil)
	if err != nil {
		t.Fatalf("unexpected error opening mock: %s", err)
	}

	mock.ExpectQueryFormat("JSONEachRow", "SELECT id FROM t WHERE id = ?").
		WithArgs(1).
		WillReturnBody([]byte(`{"id":1}`))

	body, err := mock.QueryFormat(context.Background(), "JSONEachRow", "SELECT id FROM t WHERE id = ?", 1)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	got, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("unexpected error reading body: %s", err)
	}
	if string(got) != `{"id":1}` {
		t.Errorf("unexpected body: %s", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %s", err)
	}
}

func TestQueryFormatMismatch(t *testing.T) {
	t.Parallel()
	mock, err := NewClickHouseNative(nil)
	if err != nil {
		t.Fatalf("unexpected error opening mock: %s", err)
	}

	mock.ExpectQueryFormat("JSONEachRow", "SELECT id FROM t")
	if _, err := mock.QueryFormat(context.Background(), "CSV", "SELECT id FROM t"); err == nil {
		t.Error("expected an error for an unexpected format")
	}
	if _, err := mock.QueryFormat(context.Background(), "JSONEachRow", "SELECT name FROM t"); err == nil {
		t.Error("expected an error for an unexpected query")
	}
}

func TestQueryFormatError(t *testing.T) {
	t.Parallel()
	mock, err := NewClickHouseNative(nil)
	if err != nil {
		t.Fatalf("unexpected error opening mock: %s", err)
	}

	want := errors.New("boom")
	mock.ExpectQueryFormat("CSV", "SELECT 1").WillReturnError(want)
	if _, err := mock.QueryFormat(context.Background(), "CSV", "SELECT 1"); !errors.Is(err, want) {
		t.Errorf("expected %v, got %v", want, err)
	}
}

func TestInsertFormat(t *testing.T) {
	t.Parallel()
	mock, err := NewClickHouseNative(nil)
	if err != nil {
		t.Fatalf("unexpected error opening mock: %s", err)
	}

	mock.ExpectInsertFormat("CSV", "INSERT INTO t").WithData([]byte("1,a\n"))
	if err := mock.InsertFormat(context.Background(), "CSV", "INSERT INTO t", strings.NewReader("1,a\n")); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %s", err)
	}
}

func TestInsertFormatMismatch(t *testing.T) {
	t.Parallel()
	mock, err := NewClickHouseNative(nil)
	if err != nil {
		t.Fatalf("unexpected error opening mock: %s", err)
	}

	mock.ExpectInsertFormat("CSV", "INSERT INTO t").WithData([]byte("1,a\n"))
	if err := mock.InsertFormat(context.Background(), "CSV", "INSERT INTO t", bytes.NewReader([]byte("2,b\n"))); err == nil {
		t.Error("expected an error for unexpected data")
	}
	if err := mock.InsertFormat(context.Background(), "TSV", "INSERT INTO t", bytes.NewReader(nil)); err == nil {
		t.Error("expected an error for an unexpected format")
	}
}

func TestInsertFormatError(t *testing.T) {
	t.Parallel()
	mock, err := NewClickHouseNative(nil)
	if err != nil {
		t.Fatalf("unexpected error opening mock: %s", err)
	}

	want := errors.New("boom")
	mock.ExpectInsertFormat("CSV", "INSERT INTO t").WillReturnError(want)
	if err := mock.InsertFormat(context.Background(), "CSV", "INSERT INTO t", strings.NewReader("")); !errors.Is(err, want) {
		t.Errorf("expected %v, got %v", want, err)
	}
}

func TestInsertFormatNotExpected(t *testing.T) {
	t.Parallel()
	mock, err := NewClickHouseNative(nil)
	if err != nil {
		t.Fatalf("unexpected error opening mock: %s", err)
	}

	if err := mock.InsertFormat(context.Background(), "CSV", "INSERT INTO t", strings.NewReader("")); err == nil {
		t.Error("expected an error for an unexpected call")
	}
}
