package schema

import (
	"data-comparator/internal/pkg/types"
	"io"
)

// testReader is a simple implementation for testing
type testReader struct {
	records []types.Record
	index   int
}

func newTestReader(records []types.Record) *testReader {
	return &testReader{records: records, index: 0}
}

func (r *testReader) Read() (types.Record, error) {
	if r.index >= len(r.records) {
		return nil, io.EOF
	}
	record := r.records[r.index]
	r.index++
	return record, nil
}

func (r *testReader) Close() error {
	return nil
}