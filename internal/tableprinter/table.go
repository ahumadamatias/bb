// Package tableprinter renders tabular data either as an aligned table
// with headers (when writing to a TTY) or as plain TSV (when piped),
// mirroring the behavior of cli/cli's tableprinter.
package tableprinter

import (
	"fmt"
	"io"
	"strings"
)

// TablePrinter accumulates rows and renders them either as an aligned,
// human-readable table or as TSV depending on isTTY.
type TablePrinter struct {
	out     io.Writer
	isTTY   bool
	headers []string
	rows    [][]string
}

// New creates a TablePrinter that writes to out. When isTTY is false,
// Render emits plain TSV with no header row.
func New(out io.Writer, isTTY bool) *TablePrinter {
	return &TablePrinter{out: out, isTTY: isTTY}
}

// AddHeader sets the column headers. Ignored when isTTY is false.
func (t *TablePrinter) AddHeader(headers ...string) {
	t.headers = headers
}

// AddRow appends a row of fields. The number of fields should match the
// header count when headers are set.
func (t *TablePrinter) AddRow(fields ...string) {
	t.rows = append(t.rows, fields)
}

// Render writes the accumulated table to out.
func (t *TablePrinter) Render() error {
	if !t.isTTY {
		for _, row := range t.rows {
			if _, err := fmt.Fprintln(t.out, strings.Join(row, "\t")); err != nil {
				return err
			}
		}
		return nil
	}

	widths := colWidths(t.headers, t.rows)

	if len(t.headers) > 0 {
		if err := writeRow(t.out, t.headers, widths); err != nil {
			return err
		}
	}
	for _, row := range t.rows {
		if err := writeRow(t.out, row, widths); err != nil {
			return err
		}
	}
	return nil
}

func colWidths(headers []string, rows [][]string) []int {
	numCols := len(headers)
	for _, row := range rows {
		if len(row) > numCols {
			numCols = len(row)
		}
	}

	widths := make([]int, numCols)
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, field := range row {
			if len(field) > widths[i] {
				widths[i] = len(field)
			}
		}
	}
	return widths
}

func writeRow(out io.Writer, fields []string, widths []int) error {
	parts := make([]string, len(fields))
	for i, field := range fields {
		if i == len(fields)-1 {
			// Don't pad the last column; avoids trailing whitespace.
			parts[i] = field
			continue
		}
		width := 0
		if i < len(widths) {
			width = widths[i]
		}
		parts[i] = field + strings.Repeat(" ", width-len(field)+2)
	}
	_, err := fmt.Fprintln(out, strings.Join(parts, ""))
	return err
}
