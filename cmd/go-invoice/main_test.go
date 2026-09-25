package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mrz1836/go-invoice/internal/cli"
)

func TestGenerateConfigFileContent(t *testing.T) {
	app := &App{
		logger: cli.NewLogger(false),
	}

	generate := func(bankName, bankAccount, bankRouting, paymentInstructions string) string {
		return app.generateConfigFileContent(
			"Acme Corp", "billing@example.com", "1 Example Way", "", "",
			"Net 30", bankName, bankAccount, bankRouting, paymentInstructions,
			"INV", 1000, "USD", 0, 30, "/tmp/go-invoice", false,
		)
	}

	t.Run("BankDetailsWrittenAsDomesticWireKeys", func(t *testing.T) {
		content := generate("Example Bank", "0000123456", "123456789", "Include the invoice number")

		assert.Contains(t, content, "WIRE_DOMESTIC_BANK_NAME=Example Bank\n")
		assert.Contains(t, content, "WIRE_DOMESTIC_ACCOUNT=0000123456\n")
		assert.Contains(t, content, "WIRE_DOMESTIC_ROUTING=123456789\n")
		assert.Contains(t, content, "PAYMENT_INSTRUCTIONS=Include the invoice number\n")
		assert.Contains(t, content, "set WIRE_DOMESTIC_ENABLED=true once all required fields are configured")
		assert.NotContains(t, content, "\nBANK_NAME=")
		assert.NotContains(t, content, "WIRE_DOMESTIC_ENABLED=true\n")
	})

	t.Run("InstructionsOnlyOmitsWireKeys", func(t *testing.T) {
		content := generate("", "", "", "Include the invoice number")

		assert.Contains(t, content, "# Banking Information\n")
		assert.Contains(t, content, "PAYMENT_INSTRUCTIONS=Include the invoice number\n")
		assert.NotContains(t, content, "WIRE_DOMESTIC")
	})

	t.Run("NoBankingSectionWhenEmpty", func(t *testing.T) {
		content := generate("", "", "", "")

		assert.NotContains(t, content, "# Banking Information")
		assert.NotContains(t, content, "WIRE_")
	})
}
