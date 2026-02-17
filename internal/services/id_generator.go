package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/mrz1836/go-invoice/internal/models"
)

// UUIDGenerator generates unique IDs using random UUIDs
type UUIDGenerator struct{}

// NewUUIDGenerator creates a new UUID generator
func NewUUIDGenerator() *UUIDGenerator {
	return &UUIDGenerator{}
}

// GenerateInvoiceID generates a unique invoice ID
func (g *UUIDGenerator) GenerateInvoiceID(ctx context.Context) (models.InvoiceID, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	id, err := g.generateUUID()
	if err != nil {
		return "", fmt.Errorf("failed to generate invoice ID: %w", err)
	}

	return models.InvoiceID(id), nil
}

// GenerateClientID generates a unique client ID
func (g *UUIDGenerator) GenerateClientID(ctx context.Context) (models.ClientID, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	id, err := g.generateUUID()
	if err != nil {
		return "", fmt.Errorf("failed to generate client ID: %w", err)
	}

	return models.ClientID(id), nil
}

// GenerateWorkItemID generates a unique work item ID
func (g *UUIDGenerator) GenerateWorkItemID(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	id, err := g.generateUUID()
	if err != nil {
		return "", fmt.Errorf("failed to generate work item ID: %w", err)
	}

	return id, nil
}

// generateUUID generates a random UUID v4
func (g *UUIDGenerator) generateUUID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Set version (4) and variant bits
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	// Format as UUID string
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hex.EncodeToString(b[0:4]),
		hex.EncodeToString(b[4:6]),
		hex.EncodeToString(b[6:8]),
		hex.EncodeToString(b[8:10]),
		hex.EncodeToString(b[10:16]),
	), nil
}
