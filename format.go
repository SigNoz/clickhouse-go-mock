package mockhouse

import (
	"bytes"
	"context"
	"fmt"
	"io"
)

// ExpectedQueryFormat is used to manage driver.Conn.QueryFormat expectations.
// Returned by *clickhousemock.ExpectQueryFormat.
type ExpectedQueryFormat struct {
	queryBasedExpectation
	format string
	body   []byte
}

// WithArgs will match given expected args to actual QueryFormat arguments.
func (e *ExpectedQueryFormat) WithArgs(args ...any) *ExpectedQueryFormat {
	e.args = args
	return e
}

// WillReturnBody sets the encoded result QueryFormat returns.
func (e *ExpectedQueryFormat) WillReturnBody(body []byte) *ExpectedQueryFormat {
	e.body = body
	return e
}

// WillReturnError allows to set an error for expected QueryFormat action.
func (e *ExpectedQueryFormat) WillReturnError(err error) *ExpectedQueryFormat {
	e.err = err
	return e
}

func (e *ExpectedQueryFormat) String() string {
	msg := "ExpectedQueryFormat => expecting QueryFormat which:"
	msg += "\n  - matches format: '" + e.format + "'"
	msg += "\n  - matches sql: '" + e.expectSQL + "'"
	if e.err != nil {
		msg += fmt.Sprintf("\n  - should return error: %s", e.err)
	}
	return msg
}

// ExpectedInsertFormat is used to manage driver.Conn.InsertFormat expectations.
// Returned by *clickhousemock.ExpectInsertFormat.
type ExpectedInsertFormat struct {
	queryBasedExpectation
	format string
	data   []byte
}

// WithData will match the given payload to the data InsertFormat streams.
func (e *ExpectedInsertFormat) WithData(data []byte) *ExpectedInsertFormat {
	e.data = data
	return e
}

// WillReturnError allows to set an error for expected InsertFormat action.
func (e *ExpectedInsertFormat) WillReturnError(err error) *ExpectedInsertFormat {
	e.err = err
	return e
}

func (e *ExpectedInsertFormat) String() string {
	msg := "ExpectedInsertFormat => expecting InsertFormat which:"
	msg += "\n  - matches format: '" + e.format + "'"
	msg += "\n  - matches sql: '" + e.expectSQL + "'"
	if e.err != nil {
		msg += fmt.Sprintf("\n  - should return error: %s", e.err)
	}
	return msg
}

func (c *clickhousemock) ExpectQueryFormat(format string, expectedSQL string) *ExpectedQueryFormat {
	e := &ExpectedQueryFormat{format: format}
	e.expectSQL = expectedSQL
	c.expected = append(c.expected, e)
	return e
}

func (c *clickhousemock) ExpectInsertFormat(format string, expectedSQL string) *ExpectedInsertFormat {
	e := &ExpectedInsertFormat{format: format}
	e.expectSQL = expectedSQL
	c.expected = append(c.expected, e)
	return e
}

// QueryFormat meets https://pkg.go.dev/github.com/ClickHouse/clickhouse-go/v2/lib/driver#Conn interface
func (c *clickhousemock) QueryFormat(ctx context.Context, format string, query string, args ...any) (io.ReadCloser, error) {
	c.drv.Lock()
	defer c.drv.Unlock()

	expected, err := nextExpectation[*ExpectedQueryFormat](c, "QueryFormat")
	if err != nil {
		return nil, err
	}
	defer expected.Unlock()

	if expected.format != format {
		return nil, fmt.Errorf("call to database QueryFormat with unexpected format: %s", format)
	}
	if err := c.queryMatcherFunc().Match(expected.expectSQL, query); err != nil {
		return nil, fmt.Errorf("call to database QueryFormat with unexpected query: %s", query)
	}
	if err := expected.matchArgs(args); err != nil {
		return nil, fmt.Errorf("call to database QueryFormat with unexpected args: %w", err)
	}

	expected.triggered = true
	if expected.err != nil {
		return nil, expected.err
	}
	return io.NopCloser(bytes.NewReader(expected.body)), nil
}

// InsertFormat meets https://pkg.go.dev/github.com/ClickHouse/clickhouse-go/v2/lib/driver#Conn interface
func (c *clickhousemock) InsertFormat(ctx context.Context, format string, query string, data io.Reader) error {
	c.drv.Lock()
	defer c.drv.Unlock()

	expected, err := nextExpectation[*ExpectedInsertFormat](c, "InsertFormat")
	if err != nil {
		return err
	}
	defer expected.Unlock()

	if expected.format != format {
		return fmt.Errorf("call to database InsertFormat with unexpected format: %s", format)
	}
	if err := c.queryMatcherFunc().Match(expected.expectSQL, query); err != nil {
		return fmt.Errorf("call to database InsertFormat with unexpected query: %s", query)
	}
	if expected.data != nil {
		got, err := io.ReadAll(data)
		if err != nil {
			return err
		}
		if !bytes.Equal(expected.data, got) {
			return fmt.Errorf("call to database InsertFormat with unexpected data: %q", got)
		}
	}

	expected.triggered = true
	return expected.err
}

// nextExpectation returns the first unfulfilled expectation of type T, locked.
func nextExpectation[T expectation](c *clickhousemock, op string) (T, error) {
	var zero T
	var fulfilled int
	for _, next := range c.expected {
		next.Lock()
		if next.fulfilled() {
			next.Unlock()
			fulfilled++
			continue
		}

		if expected, ok := next.(T); ok {
			return expected, nil
		}

		next.Unlock()
		if c.ordered {
			return zero, fmt.Errorf("call to database %s, was not expected, next expectation is: %s", op, next)
		}
	}

	msg := fmt.Sprintf("call to database %s was not expected", op)
	if fulfilled == len(c.expected) {
		msg = "all expectations were already fulfilled, " + msg
	}
	return zero, fmt.Errorf("%s", msg)
}
