package schemas

const (
	keyDescription          = "description"
	keyExamples             = "examples"
	keyFormat               = "format"
	keyDefault              = "default"
	keyEnum                 = "enum"
	keyAdditionalProperties = "additionalProperties"
	keyEmail                = "email"
	keyClientID             = "client_id"
	keyClientName           = "client_name"
	keyInvoiceID            = "invoice_id"
	keyHours                = "hours"
	keyRate                 = "rate"
	keyObject               = "object"
	keyProperties           = "properties"
	keyName                 = "name"
	keyMinLength            = "minLength"
	keyMaxLength            = "maxLength"
	keyRequired             = "required"
	keyMinimum              = "minimum"
	keyMaximum              = "maximum"
	keyType                 = "type"
	typeBoolean             = "boolean"
	typeArray               = "array"
	typeString              = "string"
	typeNumber              = "number"
	typeJSON                = "json"
	formatDate              = "date"
	exampleInvoiceID        = "INV-001"

	// Wire transfer client settings
	keyWireFeeEnabled       = "wire_fee_enabled"
	keyWireFeeAmount        = "wire_fee_amount"
	keyWireType             = "wire_type"
	wireTypeNone            = "none"
	wireTypeDomestic        = "domestic"
	wireTypeInternational   = "international"
	defaultWireFeeAmount    = 20.00
	maxWireFeeAmount        = 10000.00
	wireTypeDescriptionHint = "domestic shows the domestic (US) wire instructions, international shows the international wire instructions, and none shows no wire instructions."
)

// wireTypeEnum returns the accepted wire_type values.
func wireTypeEnum() []interface{} {
	return []interface{}{wireTypeNone, wireTypeDomestic, wireTypeInternational}
}
