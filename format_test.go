// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mockhouse

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestQueryFormatExpectations(t *testing.T) {
	t.Parallel()
	mock, err := NewClickHouseNative(nil)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	mock.ExpectQueryFormat("CSV", "SELECT id, title FROM articles").WillReturnData([]byte("1,\"title\"\n"))

	stream, err := mock.QueryFormat(context.Background(), "CSV", "SELECT id, title FROM articles")
	if err != nil {
		t.Fatalf("an error '%s' was not expected when querying a format", err)
	}
	defer stream.Close()

	data, err := io.ReadAll(stream)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when reading the stream", err)
	}
	if string(data) != "1,\"title\"\n" {
		t.Errorf("unexpected stream data: %q", data)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestQueryFormatError(t *testing.T) {
	t.Parallel()
	mock, err := NewClickHouseNative(nil)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	mock.ExpectQueryFormat("CSV", "SELECT id FROM articles").WillReturnError(errors.New("some error"))

	if _, err := mock.QueryFormat(context.Background(), "CSV", "SELECT id FROM articles"); err == nil {
		t.Error("an error was expected when querying a format")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestQueryFormatUnexpectedFormat(t *testing.T) {
	t.Parallel()
	mock, err := NewClickHouseNative(nil)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	mock.ExpectQueryFormat("CSV", "SELECT id FROM articles")

	if _, err := mock.QueryFormat(context.Background(), "JSONEachRow", "SELECT id FROM articles"); err == nil {
		t.Error("an error was expected when querying with an unexpected format")
	}

	if err := mock.ExpectationsWereMet(); err == nil {
		t.Error("an error was expected due to unfulfilled expectations")
	}
}

func TestInsertFormatExpectations(t *testing.T) {
	t.Parallel()
	mock, err := NewClickHouseNative(nil)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	mock.ExpectInsertFormat("CSV", "INSERT INTO articles").WithData([]byte("1,\"title\"\n"))

	if err := mock.InsertFormat(context.Background(), "CSV", "INSERT INTO articles", strings.NewReader("1,\"title\"\n")); err != nil {
		t.Errorf("an error '%s' was not expected when inserting a format", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestInsertFormatDataMismatch(t *testing.T) {
	t.Parallel()
	mock, err := NewClickHouseNative(nil)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	mock.ExpectInsertFormat("CSV", "INSERT INTO articles").WithData([]byte("1,\"title\"\n"))

	if err := mock.InsertFormat(context.Background(), "CSV", "INSERT INTO articles", strings.NewReader("2,\"other\"\n")); err == nil {
		t.Error("an error was expected when inserting unexpected data")
	}

	if err := mock.ExpectationsWereMet(); err == nil {
		t.Error("an error was expected due to unfulfilled expectations")
	}
}

func TestInsertFormatError(t *testing.T) {
	t.Parallel()
	mock, err := NewClickHouseNative(nil)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	mock.ExpectInsertFormat("Parquet", "INSERT INTO articles").WillReturnError(errors.New("some error"))

	if err := mock.InsertFormat(context.Background(), "Parquet", "INSERT INTO articles", strings.NewReader("data")); err == nil {
		t.Error("an error was expected when inserting a format")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
