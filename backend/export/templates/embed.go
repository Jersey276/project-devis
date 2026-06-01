// Package templates exposes the embedded HTML template used by the export
// renderer. Kept in its own package so the //go:embed directive doesn't have
// to live next to the rendering logic.
package templates

import _ "embed"

//go:embed quote.html
var QuoteHTML []byte

//go:embed schedule.html
var ScheduleHTML []byte
