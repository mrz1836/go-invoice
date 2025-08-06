// Package templates provides embedded invoice templates.
package templates

import _ "embed"

// DefaultInvoiceTemplate contains the embedded default invoice template
//
//go:embed default.html
var DefaultInvoiceTemplate string
