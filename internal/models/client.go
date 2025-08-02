// Package models defines the core data structures and types for the invoice system.
package models

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// NewClient creates a new client with validation
func NewClient(ctx context.Context, id ClientID, name, email string) (*Client, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	now := time.Now()
	client := &Client{
		ID:        id,
		Name:      name,
		Email:     email,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Validate the new client
	if err := client.Validate(ctx); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrClientValidationFailed, err)
	}

	return client, nil
}

// Validate performs comprehensive validation of the client
func (c *Client) Validate(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return NewValidationBuilder().
		AddRequired("id", string(c.ID)).
		AddRequired("name", c.Name).
		AddMaxLength("name", c.Name, 200).
		AddRequired("email", c.Email).
		AddEmail("email", c.Email).
		AddLengthRange("phone", c.Phone, 10, 20).
		AddMaxLength("address", c.Address, 500).
		AddMaxLength("tax_id", c.TaxID, 50).
		AddTimeRequired("created_at", c.CreatedAt).
		AddTimeRequired("updated_at", c.UpdatedAt).
		AddTimeOrder("updated_at", c.CreatedAt, c.UpdatedAt, "created_at", "updated_at").
		Build(ErrClientValidationFailed)
}

// UpdateName updates the client name with validation
func (c *Client) UpdateName(ctx context.Context, name string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return ErrNameRequired
	}

	if len(name) > 200 {
		return ErrClientNameTooLong
	}

	c.Name = name
	c.UpdatedAt = time.Now()
	return nil
}

// UpdateEmail updates the client email with validation
func (c *Client) UpdateEmail(ctx context.Context, email string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	email = strings.TrimSpace(email)
	if email == "" {
		return ErrEmailRequired
	}

	if !emailPattern.MatchString(email) {
		return ErrClientEmailInvalid
	}

	c.Email = email
	c.UpdatedAt = time.Now()
	return nil
}

// UpdatePhone updates the client phone with validation
func (c *Client) UpdatePhone(ctx context.Context, phone string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	phone = strings.TrimSpace(phone)
	if phone != "" {
		if len(phone) < 10 || len(phone) > 20 {
			return ErrClientPhoneInvalid
		}
	}

	c.Phone = phone
	c.UpdatedAt = time.Now()
	return nil
}

// UpdateAddress updates the client address with validation
func (c *Client) UpdateAddress(ctx context.Context, address string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	address = strings.TrimSpace(address)
	if address != "" && len(address) > 500 {
		return ErrClientAddressTooLong
	}

	c.Address = address
	c.UpdatedAt = time.Now()
	return nil
}

// UpdateTaxID updates the client tax ID with validation
func (c *Client) UpdateTaxID(ctx context.Context, taxID string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	taxID = strings.TrimSpace(taxID)
	if taxID != "" && len(taxID) > 50 {
		return ErrClientTaxIDTooLong
	}

	c.TaxID = taxID
	c.UpdatedAt = time.Now()
	return nil
}

// Deactivate marks the client as inactive
func (c *Client) Deactivate(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	c.Active = false
	c.UpdatedAt = time.Now()
	return nil
}

// Activate marks the client as active
func (c *Client) Activate(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	c.Active = true
	c.UpdatedAt = time.Now()
	return nil
}

// GetDisplayName returns a formatted display name for the client
func (c *Client) GetDisplayName() string {
	if c.Name == "" {
		return string(c.ID)
	}
	return c.Name
}

// GetContactInfo returns formatted contact information
func (c *Client) GetContactInfo() string {
	var parts []string

	if c.Email != "" {
		parts = append(parts, c.Email)
	}

	if c.Phone != "" {
		parts = append(parts, c.Phone)
	}

	return strings.Join(parts, " | ")
}

// HasCompleteInfo returns true if the client has all basic contact information
func (c *Client) HasCompleteInfo() bool {
	return c.Name != "" && c.Email != "" && c.Address != ""
}

// CreateClientRequest represents a request to create a new client
type CreateClientRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone,omitempty"`
	Address string `json:"address,omitempty"`
	TaxID   string `json:"tax_id,omitempty"`
}

// Validate validates the create client request
func (r *CreateClientRequest) Validate(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return NewValidationBuilder().
		AddRequired("name", r.Name).
		AddMaxLength("name", r.Name, 200).
		AddRequired("email", r.Email).
		AddEmail("email", r.Email).
		AddLengthRange("phone", r.Phone, 10, 20).
		AddMaxLength("address", r.Address, 500).
		AddMaxLength("tax_id", r.TaxID, 50).
		Build(ErrCreateClientRequestInvalid)
}
