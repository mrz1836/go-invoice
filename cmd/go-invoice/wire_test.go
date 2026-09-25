package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/go-invoice/internal/cli"
	"github.com/mrz1836/go-invoice/internal/config"
	"github.com/mrz1836/go-invoice/internal/models"
	jsonStorage "github.com/mrz1836/go-invoice/internal/storage/json"
)

// Obviously fake wire details used across the wire transfer tests
const (
	testWireAccount = "000123456789"
	testWireRouting = "012345678"
	testWireIBAN    = "GB00TEST00000000000000"
)

// wireEnvKeys lists every wire transfer setting read from the environment
func wireEnvKeys() []string {
	return []string{
		"WIRE_DOMESTIC_ENABLED",
		"WIRE_DOMESTIC_BENEFICIARY",
		"WIRE_DOMESTIC_BANK_NAME",
		"WIRE_DOMESTIC_ACCOUNT",
		"WIRE_DOMESTIC_ROUTING",
		"WIRE_DOMESTIC_ACCOUNT_TYPE",
		"WIRE_DOMESTIC_REFERENCE",
		"WIRE_INTL_ENABLED",
		"WIRE_INTL_BENEFICIARY",
		"WIRE_INTL_BENEFICIARY_ADDRESS",
		"WIRE_INTL_BANK_NAME",
		"WIRE_INTL_BANK_ADDRESS",
		"WIRE_INTL_SWIFT",
		"WIRE_INTL_IBAN",
		"WIRE_INTL_ACCOUNT",
		"WIRE_INTL_INTERMEDIARY_BANK",
		"WIRE_INTL_INTERMEDIARY_SWIFT",
		"WIRE_INTL_REFERENCE",
	}
}

func completeDomesticWire() config.DomesticWire {
	return config.DomesticWire{
		Enabled:         true,
		BeneficiaryName: "Example Consulting LLC",
		BankName:        "Example Bank",
		AccountNumber:   testWireAccount,
		RoutingNumber:   testWireRouting,
		AccountType:     "Checking",
		Reference:       "Invoice number",
	}
}

func completeInternationalWire() config.InternationalWire {
	return config.InternationalWire{
		Enabled:            true,
		BeneficiaryName:    "Example Consulting LLC",
		BeneficiaryAddress: "1 Example Way, Springfield, US",
		BankName:           "Example Bank",
		BankAddress:        "2 Example Plaza, Springfield, US",
		SWIFT:              "TESTGB2L",
		IBAN:               testWireIBAN,
		Reference:          "Invoice number",
	}
}

func TestResolveWireTransfer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		wireType    models.WireType
		business    func() config.BusinessConfig
		wantEnabled bool
		wantErr     error
		wantInError []string
	}{
		{
			name:     "NoneDisablesWire",
			wireType: models.WireTypeNone,
			business: func() config.BusinessConfig {
				return config.BusinessConfig{DomesticWire: completeDomesticWire(), InternationalWire: completeInternationalWire()}
			},
		},
		{
			name:     "UnsetTreatedAsNone",
			wireType: "",
			business: func() config.BusinessConfig { return config.BusinessConfig{} },
		},
		{
			name:     "NoneIgnoresMissingConfiguration",
			wireType: models.WireTypeNone,
			business: func() config.BusinessConfig { return config.BusinessConfig{} },
		},
		{
			name:        "UnknownWireTypeRejected",
			wireType:    "carrier-pigeon",
			business:    func() config.BusinessConfig { return config.BusinessConfig{} },
			wantErr:     ErrUnknownWireType,
			wantInError: []string{"carrier-pigeon"},
		},
		{
			name:        "DomesticComplete",
			wireType:    models.WireTypeDomestic,
			business:    func() config.BusinessConfig { return config.BusinessConfig{DomesticWire: completeDomesticWire()} },
			wantEnabled: true,
		},
		{
			name:     "DomesticDisabled",
			wireType: models.WireTypeDomestic,
			business: func() config.BusinessConfig {
				wire := completeDomesticWire()
				wire.Enabled = false
				return config.BusinessConfig{DomesticWire: wire, InternationalWire: completeInternationalWire()}
			},
			wantErr:     ErrWireTransferDisabled,
			wantInError: []string{"domestic"},
		},
		{
			name:     "DomesticMissingRequiredFields",
			wireType: models.WireTypeDomestic,
			business: func() config.BusinessConfig {
				wire := completeDomesticWire()
				wire.AccountNumber = ""
				wire.RoutingNumber = "   "
				wire.Reference = ""
				return config.BusinessConfig{DomesticWire: wire}
			},
			wantErr:     ErrWireTransferIncomplete,
			wantInError: []string{"WIRE_DOMESTIC_ACCOUNT", "WIRE_DOMESTIC_ROUTING", "WIRE_DOMESTIC_REFERENCE"},
		},
		{
			name:     "DomesticNotSatisfiedByInternationalSet",
			wireType: models.WireTypeDomestic,
			business: func() config.BusinessConfig {
				return config.BusinessConfig{InternationalWire: completeInternationalWire()}
			},
			wantErr: ErrWireTransferDisabled,
		},
		{
			name:     "InternationalCompleteWithIBAN",
			wireType: models.WireTypeInternational,
			business: func() config.BusinessConfig {
				return config.BusinessConfig{InternationalWire: completeInternationalWire()}
			},
			wantEnabled: true,
		},
		{
			name:     "InternationalCompleteWithAccountNumberOnly",
			wireType: models.WireTypeInternational,
			business: func() config.BusinessConfig {
				wire := completeInternationalWire()
				wire.IBAN = ""
				wire.AccountNumber = testWireAccount
				return config.BusinessConfig{InternationalWire: wire}
			},
			wantEnabled: true,
		},
		{
			name:     "InternationalCompleteWithIntermediaryPair",
			wireType: models.WireTypeInternational,
			business: func() config.BusinessConfig {
				wire := completeInternationalWire()
				wire.IntermediaryBank = "Example Correspondent Bank"
				wire.IntermediarySWIFT = "TESTUS33"
				return config.BusinessConfig{InternationalWire: wire}
			},
			wantEnabled: true,
		},
		{
			name:     "InternationalDisabled",
			wireType: models.WireTypeInternational,
			business: func() config.BusinessConfig {
				wire := completeInternationalWire()
				wire.Enabled = false
				return config.BusinessConfig{DomesticWire: completeDomesticWire(), InternationalWire: wire}
			},
			wantErr:     ErrWireTransferDisabled,
			wantInError: []string{"international"},
		},
		{
			name:     "InternationalMissingIBANAndAccount",
			wireType: models.WireTypeInternational,
			business: func() config.BusinessConfig {
				wire := completeInternationalWire()
				wire.IBAN = ""
				return config.BusinessConfig{InternationalWire: wire}
			},
			wantErr:     ErrWireTransferIncomplete,
			wantInError: []string{"WIRE_INTL_IBAN or WIRE_INTL_ACCOUNT"},
		},
		{
			name:     "InternationalMissingRequiredFields",
			wireType: models.WireTypeInternational,
			business: func() config.BusinessConfig {
				wire := completeInternationalWire()
				wire.BeneficiaryAddress = ""
				wire.BankAddress = ""
				wire.SWIFT = ""
				return config.BusinessConfig{InternationalWire: wire}
			},
			wantErr:     ErrWireTransferIncomplete,
			wantInError: []string{"WIRE_INTL_BENEFICIARY_ADDRESS", "WIRE_INTL_BANK_ADDRESS", "WIRE_INTL_SWIFT"},
		},
		{
			name:     "InternationalIntermediaryBankWithoutSWIFT",
			wireType: models.WireTypeInternational,
			business: func() config.BusinessConfig {
				wire := completeInternationalWire()
				wire.IntermediaryBank = "Example Correspondent Bank"
				return config.BusinessConfig{InternationalWire: wire}
			},
			wantErr:     ErrWireTransferIncomplete,
			wantInError: []string{"WIRE_INTL_INTERMEDIARY_SWIFT"},
		},
		{
			name:     "InternationalIntermediarySWIFTWithoutBank",
			wireType: models.WireTypeInternational,
			business: func() config.BusinessConfig {
				wire := completeInternationalWire()
				wire.IntermediarySWIFT = "TESTUS33"
				return config.BusinessConfig{InternationalWire: wire}
			},
			wantErr:     ErrWireTransferIncomplete,
			wantInError: []string{"WIRE_INTL_INTERMEDIARY_BANK"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			enabled, err := resolveWireTransfer(tt.wireType, tt.business())

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.False(t, enabled)
				for _, want := range tt.wantInError {
					assert.Contains(t, err.Error(), want)
				}
				// Configured wire values must never be echoed back in errors
				for _, secret := range []string{testWireAccount, testWireRouting, testWireIBAN} {
					assert.NotContains(t, err.Error(), secret)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantEnabled, enabled)
		})
	}
}

// setWireTestEnv isolates configuration loading from the host environment: it points the
// data directory at dataDir, supplies the required business settings, clears every wire
// setting, and then applies the given overrides.
func setWireTestEnv(t *testing.T, dataDir string, overrides map[string]string) string {
	t.Helper()

	t.Setenv("DATA_DIR", dataDir)
	t.Setenv("BUSINESS_NAME", "Example Consulting LLC")
	t.Setenv("BUSINESS_ADDRESS", "1 Example Way, Springfield, US")
	t.Setenv("BUSINESS_EMAIL", "billing@example.com")
	t.Setenv("USDC_ENABLED", "")
	for _, key := range wireEnvKeys() {
		t.Setenv(key, "")
	}
	for key, value := range overrides {
		t.Setenv(key, value)
	}

	// A config path that does not exist so only the environment above is used
	return filepath.Join(dataDir, "absent.env.config")
}

func newWireTestApp() *App {
	logger := cli.NewLogger(false)
	return &App{
		logger:        logger,
		configService: config.NewConfigService(logger, config.NewSimpleValidator(logger)),
	}
}

func newInitializedStorage(t *testing.T, dataDir string) *jsonStorage.JSONStorage {
	t.Helper()
	store := jsonStorage.NewJSONStorage(dataDir, cli.NewLogger(false))
	require.NoError(t, store.Initialize(context.Background()))
	return store
}

func runClientCommand(t *testing.T, cmd *cobra.Command, configPath string, args ...string) error {
	t.Helper()
	cmd.Flags().String("config", configPath, "Path to configuration file")
	cmd.SetArgs(args)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	return cmd.Execute()
}

func TestClientWireFlagsRegistered(t *testing.T) {
	t.Parallel()

	app := newWireTestApp()
	commands := map[string]*cobra.Command{
		"create": app.buildClientCreateCommand(),
		"update": app.buildClientUpdateCommand(),
	}

	for name, cmd := range commands {
		wireFee := cmd.Flags().Lookup("wire-fee")
		require.NotNilf(t, wireFee, "client %s registers --wire-fee", name)
		assert.Equal(t, "bool", wireFee.Value.Type())
		assert.Equal(t, "false", wireFee.DefValue)

		wireFeeAmount := cmd.Flags().Lookup("wire-fee-amount")
		require.NotNilf(t, wireFeeAmount, "client %s registers --wire-fee-amount", name)
		assert.Equal(t, "float64", wireFeeAmount.Value.Type())
		assert.Equal(t, "20", wireFeeAmount.DefValue)

		wireType := cmd.Flags().Lookup("wire-type")
		require.NotNilf(t, wireType, "client %s registers --wire-type", name)
		assert.Equal(t, "string", wireType.Value.Type())
		assert.Equal(t, "none", wireType.DefValue)
	}
}

func TestClientWireFlagsRoundTrip(t *testing.T) {
	dataDir := t.TempDir()
	configPath := setWireTestEnv(t, dataDir, nil)
	store := newInitializedStorage(t, dataDir)

	findClient := func(t *testing.T, email string) *models.Client {
		t.Helper()
		client, err := store.FindClientByEmail(context.Background(), email)
		require.NoError(t, err)
		return client
	}

	t.Run("CreateDefaultsToNoWire", func(t *testing.T) {
		err := runClientCommand(t, newWireTestApp().buildClientCreateCommand(), configPath,
			"--name", "Example Plain Co", "--email", "plain@example.com")
		require.NoError(t, err)

		client := findClient(t, "plain@example.com")
		assert.Equal(t, models.WireTypeNone, client.WireType)
		assert.False(t, client.WireFeeEnabled)
	})

	t.Run("CreateWithWireSettings", func(t *testing.T) {
		err := runClientCommand(t, newWireTestApp().buildClientCreateCommand(), configPath,
			"--name", "Acme Corp", "--email", "ap@acme.example",
			"--wire-type", "international", "--wire-fee")
		require.NoError(t, err)

		client := findClient(t, "ap@acme.example")
		assert.Equal(t, models.WireTypeInternational, client.WireType)
		assert.True(t, client.WireFeeEnabled)
		assert.InDelta(t, models.DefaultWireFeeAmount, client.WireFeeAmount, 0.001)
		assert.False(t, client.CryptoFeeEnabled)
	})

	t.Run("CreateRejectsInvalidWireType", func(t *testing.T) {
		err := runClientCommand(t, newWireTestApp().buildClientCreateCommand(), configPath,
			"--name", "Example Invalid Co", "--email", "invalid@example.com", "--wire-type", "carrier-pigeon")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "wire_type")

		_, findErr := store.FindClientByEmail(context.Background(), "invalid@example.com")
		require.Error(t, findErr, "rejected client must not be stored")
	})

	t.Run("UpdateChangesOnlyGivenWireSettings", func(t *testing.T) {
		clientID := string(findClient(t, "ap@acme.example").ID)

		err := runClientCommand(t, newWireTestApp().buildClientUpdateCommand(), configPath,
			clientID, "--wire-type", "domestic", "--wire-fee-amount", "35")
		require.NoError(t, err)

		client := findClient(t, "ap@acme.example")
		assert.Equal(t, models.WireTypeDomestic, client.WireType)
		assert.True(t, client.WireFeeEnabled, "unchanged flag keeps its stored value")
		assert.InDelta(t, 35.0, client.WireFeeAmount, 0.001)
	})

	t.Run("UpdateDisablesWireFee", func(t *testing.T) {
		clientID := string(findClient(t, "ap@acme.example").ID)

		err := runClientCommand(t, newWireTestApp().buildClientUpdateCommand(), configPath,
			clientID, "--wire-fee=false")
		require.NoError(t, err)

		client := findClient(t, "ap@acme.example")
		assert.False(t, client.WireFeeEnabled)
		assert.Equal(t, models.WireTypeDomestic, client.WireType)
		assert.InDelta(t, 35.0, client.WireFeeAmount, 0.001)
	})

	t.Run("UpdateRejectsInvalidWireType", func(t *testing.T) {
		clientID := string(findClient(t, "ap@acme.example").ID)

		err := runClientCommand(t, newWireTestApp().buildClientUpdateCommand(), configPath,
			clientID, "--wire-type", "carrier-pigeon")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "wire_type")

		assert.Equal(t, models.WireTypeDomestic, findClient(t, "ap@acme.example").WireType)
	})

	t.Run("UpdateRejectsOutOfRangeWireFee", func(t *testing.T) {
		clientID := string(findClient(t, "ap@acme.example").ID)

		err := runClientCommand(t, newWireTestApp().buildClientUpdateCommand(), configPath,
			clientID, "--wire-fee-amount", "-5")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "wire_fee_amount")

		assert.InDelta(t, 35.0, findClient(t, "ap@acme.example").WireFeeAmount, 0.001)
	})
}

// wireGenerateFixture stores one client and one draft invoice worth $1,000 at a 10% tax rate
type wireGenerateFixture struct {
	app        *App
	store      *jsonStorage.JSONStorage
	configPath string
	dataDir    string
	invoice    *models.Invoice
}

func newWireGenerateFixture(t *testing.T, env map[string]string, configureClient func(*models.Client), configureInvoice func(*models.Invoice)) *wireGenerateFixture {
	t.Helper()
	ctx := context.Background()
	dataDir := t.TempDir()
	configPath := setWireTestEnv(t, dataDir, env)
	store := newInitializedStorage(t, dataDir)

	client, err := models.NewClient(ctx, "client-wire", "Acme Corp", "ap@acme.example")
	require.NoError(t, err)
	if configureClient != nil {
		configureClient(client)
	}
	require.NoError(t, store.CreateClient(ctx, client))

	now := time.Now()
	invoice, err := models.NewInvoice(ctx, "wire-invoice", "INV-9001", now, now.AddDate(0, 0, 30), *client, 0.10)
	require.NoError(t, err)
	item, err := models.NewWorkItem(ctx, "work-1", now, 10, 100, "Consulting services")
	require.NoError(t, err)
	require.NoError(t, invoice.AddWorkItem(ctx, *item))
	if configureInvoice != nil {
		configureInvoice(invoice)
	}
	require.NoError(t, store.CreateInvoice(ctx, invoice))

	return &wireGenerateFixture{
		app:        newWireTestApp(),
		store:      store,
		configPath: configPath,
		dataDir:    dataDir,
		invoice:    invoice,
	}
}

func (f *wireGenerateFixture) generate(t *testing.T) error {
	t.Helper()
	return f.app.executeGenerateInvoice(context.Background(), string(f.invoice.ID), f.configPath,
		GenerateInvoiceOptions{TemplateName: "default"})
}

func (f *wireGenerateFixture) stored(t *testing.T) *models.Invoice {
	t.Helper()
	invoice, err := f.store.GetInvoice(context.Background(), f.invoice.ID)
	require.NoError(t, err)
	return invoice
}

func (f *wireGenerateFixture) outputExists() bool {
	_, err := os.Stat(filepath.Join(f.dataDir, "generated", f.invoice.Number+".html"))
	return err == nil
}

func internationalWireEnv() map[string]string {
	return map[string]string{
		"WIRE_INTL_ENABLED":             "true",
		"WIRE_INTL_BENEFICIARY":         "Example Consulting LLC",
		"WIRE_INTL_BENEFICIARY_ADDRESS": "1 Example Way, Springfield, US",
		"WIRE_INTL_BANK_NAME":           "Example Bank",
		"WIRE_INTL_BANK_ADDRESS":        "2 Example Plaza, Springfield, US",
		"WIRE_INTL_SWIFT":               "TESTGB2L",
		"WIRE_INTL_IBAN":                testWireIBAN,
		"WIRE_INTL_REFERENCE":           "Invoice number",
	}
}

func domesticWireEnv() map[string]string {
	return map[string]string{
		"WIRE_DOMESTIC_ENABLED":      "true",
		"WIRE_DOMESTIC_BENEFICIARY":  "Example Consulting LLC",
		"WIRE_DOMESTIC_BANK_NAME":    "Example Bank",
		"WIRE_DOMESTIC_ACCOUNT":      testWireAccount,
		"WIRE_DOMESTIC_ROUTING":      testWireRouting,
		"WIRE_DOMESTIC_ACCOUNT_TYPE": "Checking",
		"WIRE_DOMESTIC_REFERENCE":    "Invoice number",
	}
}

func TestExecuteGenerateInvoiceWireTransfer(t *testing.T) {
	t.Run("InternationalAppliesDefaultWireFee", func(t *testing.T) {
		f := newWireGenerateFixture(t, internationalWireEnv(), func(c *models.Client) {
			c.WireType = models.WireTypeInternational
			c.WireFeeEnabled = true
		}, nil)

		require.NoError(t, f.generate(t))

		stored := f.stored(t)
		assert.InDelta(t, models.DefaultWireFeeAmount, stored.WireFee, 0.001)
		assert.InDelta(t, 1000.0, stored.Subtotal, 0.001)
		assert.InDelta(t, 102.0, stored.TaxAmount, 0.001, "wire fee is taxed")
		assert.InDelta(t, 1122.0, stored.Total, 0.001)
		assert.Equal(t, models.WireTypeInternational, stored.Client.WireType)
		assert.True(t, f.outputExists())
	})

	t.Run("DomesticAppliesConfiguredWireFee", func(t *testing.T) {
		f := newWireGenerateFixture(t, domesticWireEnv(), func(c *models.Client) {
			c.WireType = models.WireTypeDomestic
			c.WireFeeEnabled = true
			c.WireFeeAmount = 35
		}, nil)

		require.NoError(t, f.generate(t))

		stored := f.stored(t)
		assert.InDelta(t, 35.0, stored.WireFee, 0.001)
		assert.InDelta(t, 1138.5, stored.Total, 0.001)
		assert.True(t, f.outputExists())
	})

	t.Run("WireWithoutFeeAddsNothing", func(t *testing.T) {
		f := newWireGenerateFixture(t, domesticWireEnv(), func(c *models.Client) {
			c.WireType = models.WireTypeDomestic
		}, nil)

		require.NoError(t, f.generate(t))

		stored := f.stored(t)
		assert.InDelta(t, 0.0, stored.WireFee, 0.001)
		assert.InDelta(t, 1100.0, stored.Total, 0.001)
	})

	t.Run("NoneClearsStaleWireFee", func(t *testing.T) {
		f := newWireGenerateFixture(t, nil, func(c *models.Client) {
			c.WireType = models.WireTypeNone
			c.WireFeeEnabled = true
		}, func(inv *models.Invoice) {
			require.NoError(t, inv.SetWireFee(context.Background(), true, true, 20))
		})
		require.InDelta(t, 20.0, f.stored(t).WireFee, 0.001, "fixture starts with a stale wire fee")

		require.NoError(t, f.generate(t))

		stored := f.stored(t)
		assert.InDelta(t, 0.0, stored.WireFee, 0.001)
		assert.InDelta(t, 1100.0, stored.Total, 0.001)
		assert.True(t, f.outputExists())
	})

	t.Run("DisabledSetFailsBeforeSaveOrRender", func(t *testing.T) {
		env := internationalWireEnv()
		env["WIRE_INTL_ENABLED"] = "false"
		f := newWireGenerateFixture(t, env, func(c *models.Client) {
			c.WireType = models.WireTypeInternational
			c.WireFeeEnabled = true
		}, nil)

		err := f.generate(t)

		require.ErrorIs(t, err, ErrWireTransferDisabled)
		assert.Contains(t, err.Error(), f.invoice.Number)
		stored := f.stored(t)
		assert.Equal(t, f.invoice.Version, stored.Version, "invoice must not be saved")
		assert.InDelta(t, 0.0, stored.WireFee, 0.001)
		assert.False(t, f.outputExists(), "invoice must not be rendered")
	})

	t.Run("IncompleteSetFailsBeforeSaveOrRender", func(t *testing.T) {
		env := domesticWireEnv()
		delete(env, "WIRE_DOMESTIC_ROUTING")
		f := newWireGenerateFixture(t, env, func(c *models.Client) {
			c.WireType = models.WireTypeDomestic
			c.WireFeeEnabled = true
		}, nil)

		err := f.generate(t)

		require.ErrorIs(t, err, ErrWireTransferIncomplete)
		assert.Contains(t, err.Error(), "WIRE_DOMESTIC_ROUTING")
		assert.NotContains(t, err.Error(), testWireAccount)
		stored := f.stored(t)
		assert.Equal(t, f.invoice.Version, stored.Version, "invoice must not be saved")
		assert.InDelta(t, 0.0, stored.WireFee, 0.001)
		assert.False(t, f.outputExists(), "invoice must not be rendered")
	})
}
