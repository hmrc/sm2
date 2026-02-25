package servicemanager

import (
	"fmt"
	"strings"
)

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
)

// Status constants
const (
	StatusRunning = "RUNNING"
	StatusOK      = "OK"
	StatusError   = "ERROR"
	StatusWarn    = "WARN"
	StatusInfo    = "INFO"
)

// Component name constants
const (
	CompOS        = "OS"
	CompJava      = "JAVA"
	CompGit       = "GIT"
	CompConfig    = "CONFIG"
	CompWorkspace = "WORKSPACE"
	CompVpnDns    = "VPN DNS"
	CompVpn       = "VPN"
)

// Table builds and prints a simple ASCII table with dynamic column widths.
type Table struct {
	indent string
	colSep string
	widths []int
	rights []bool
}

// NewTable creates a new Table with the given indent and column separator strings.
func NewTable(indent, colSep string) *Table {
	return &Table{indent: indent, colSep: colSep}
}

// SetMinWidth ensures column col has at least minWidth, with optional right-alignment.
func (t *Table) SetMinWidth(col, minWidth int, rightAlign bool) {
	for len(t.widths) <= col {
		t.widths = append(t.widths, 0)
		t.rights = append(t.rights, false)
	}
	if minWidth > t.widths[col] {
		t.widths[col] = minWidth
	}
	t.rights[col] = rightAlign
}

// FitRow expands column widths to accommodate the given cell values.
func (t *Table) FitRow(cells ...string) {
	for i, cell := range cells {
		for len(t.widths) <= i {
			t.widths = append(t.widths, 0)
			t.rights = append(t.rights, false)
		}
		if len(cell) > t.widths[i] {
			t.widths[i] = len(cell)
		}
	}
}

// Width returns the total rendered width of a table row.
func (t *Table) Width() int {
	w := len(t.indent)
	for i, mw := range t.widths {
		w += mw
		if i < len(t.widths)-1 {
			w += len(t.colSep)
		}
	}
	return w
}

// Separator prints a horizontal separator line spanning the full table width.
func (t *Table) Separator() {
	fmt.Println(strings.Repeat("-", t.Width()))
}

// PrintRow prints a single formatted row using the current column widths.
// The last column is never padded to avoid trailing spaces.
func (t *Table) PrintRow(cells ...string) {
	t.PrintColoredRow(cells, nil)
}

// PrintColoredRow prints a formatted row where each cell can optionally be
// wrapped in an ANSI color. colors[i] should be a color constant (e.g.
// ColorGreen) or empty string for no color. Padding is applied to the plain
// text before the color codes are added, keeping column alignment correct.
func (t *Table) PrintColoredRow(cells []string, colors []string) {
	fmt.Print(t.indent)
	last := len(t.widths) - 1
	for i, w := range t.widths {
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		color := ""
		if i < len(colors) {
			color = colors[i]
		}

		// Pad the plain text first, then wrap with color so that ANSI escape
		// codes do not interfere with the width arithmetic.
		var padded string
		if i == last {
			padded = cell
		} else if i < len(t.rights) && t.rights[i] {
			padded = fmt.Sprintf("%*s", w, cell)
		} else {
			padded = fmt.Sprintf("%-*s", w, cell)
		}

		if color != "" {
			fmt.Printf("%s%s%s", color, padded, ColorReset)
		} else {
			fmt.Print(padded)
		}

		if i < last {
			fmt.Print(t.colSep)
		}
	}
	fmt.Println()
}

// startStatus prints an initial RUNNING status for a diagnostic component.
func startStatus(component string, noProgress bool) {
	if !noProgress {
		printStatus(component, StatusRunning, "...", noProgress)
	}
}

// printStatus prints a diagnostic status line with colored status.
func printStatus(component string, status string, details string, noProgress bool) {
	var colorCode string

	switch status {
	case StatusRunning:
		colorCode = ColorYellow
	case StatusOK:
		colorCode = ColorGreen
	case StatusError:
		colorCode = ColorRed
	case StatusWarn:
		colorCode = ColorYellow
	case StatusInfo:
		colorCode = ColorReset
	default:
		colorCode = ColorReset
	}

	// Format component name to be exactly 15 characters
	formattedComponent := component
	if len(component) > 15 {
		formattedComponent = component[:15] + ":"
	} else if len(component) < 15 {
		formattedComponent = component + ":" + strings.Repeat(" ", 14-len(component))
	}

	formattedStatus := fmt.Sprintf("%s%s%s", colorCode, status, ColorReset)

	if noProgress && status == StatusRunning {
		// do not print message for --noprogress flag
	} else {
		fmt.Printf("%s%s (%s)\n", formattedComponent, formattedStatus, details)
	}
}

// updateStatus overwrites the previous status line for a diagnostic component.
func updateStatus(component string, status string, details string, noProgress bool) {
	if !noProgress {
		// Move cursor up one line and clear the line
		fmt.Print("\033[1A\033[K")
	}
	printStatus(component, status, details, noProgress)
}
