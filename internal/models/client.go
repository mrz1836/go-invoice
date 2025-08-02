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
		return nil, fmt.Errorf("client validation failed: %w", err)
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

	var errors []ValidationError

	// Validate ID
	if strings.TrimSpace(string(c.ID)) == "" {
		errors = append(errors, ValidationError{
			Field:   "id",
			Message: "is required",
			Value:   c.ID,
		})
	}

	// Validate name
	if strings.TrimSpace(c.Name) == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "is required",
			Value:   c.Name,
		})
	}

	// Validate name length
	if len(strings.TrimSpace(c.Name)) > 200 {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "cannot exceed 200 characters",
			Value:   len(c.Name),
		})
	}

	// Validate email
	if strings.TrimSpace(c.Email) == "" {
		errors = append(errors, ValidationError{
			Field:   "email",
			Message: "is required",
			Value:   c.Email,
		})
	} else if !emailPattern.MatchString(c.Email) {
		errors = append(errors, ValidationError{
			Field:   "email",
			Message: "must be a valid email address",
			Value:   c.Email,
		})
	}

	// Validate phone if provided
	if c.Phone != "" {
		phone := strings.TrimSpace(c.Phone)
		if len(phone) < 10 || len(phone) > 20 {
			errors = append(errors, ValidationError{
				Field:   "phone",
				Message: "must be between 10 and 20 characters",
				Value:   c.Phone,
			})
		}
	}

	// Validate address length if provided
	if c.Address != "" && len(strings.TrimSpace(c.Address)) > 500 {
		errors = append(errors, ValidationError{
			Field:   "address",
			Message: "cannot exceed 500 characters",
			Value:   len(c.Address),
		})
	}

	// Validate tax ID length if provided
	if c.TaxID != "" && len(strings.TrimSpace(c.TaxID)) > 50 {
		errors = append(errors, ValidationError{
			Field:   "tax_id",
			Message: "cannot exceed 50 characters",
			Value:   len(c.TaxID),
		})
	}

	// Validate timestamps
	if c.CreatedAt.IsZero() {
		errors = append(errors, ValidationError{
			Field:   "created_at",
			Message: "is required",
			Value:   c.CreatedAt,
		})
	}

	if c.UpdatedAt.IsZero() {
		errors = append(errors, ValidationError{
			Field:   "updated_at",
			Message: "is required",
			Value:   c.UpdatedAt,
		})
	}

	if !c.CreatedAt.IsZero() && !c.UpdatedAt.IsZero() && c.UpdatedAt.Before(c.CreatedAt) {
		errors = append(errors, ValidationError{
			Field:   "updated_at",
			Message: "must be on or after created_at",
			Value:   fmt.Sprintf("updated: %v, created: %v", c.UpdatedAt, c.CreatedAt),
		})
	}

	if len(errors) > 0 {
		var messages []string
		for _, err := range errors {
			messages = append(messages, err.Error())
		}
		return fmt.Errorf("client validation failed: %s", strings.Join(messages, "; "))
	}

	return nil
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
		return fmt.Errorf("name cannot exceed 200 characters")
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
		return fmt.Errorf("email must be a valid email address")
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
			return fmt.Errorf("phone must be between 10 and 20 characters")
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
		return fmt.Errorf("address cannot exceed 500 characters")
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
		return fmt.Errorf("tax ID cannot exceed 50 characters")
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

	var errors []ValidationError

	// Validate name
	if strings.TrimSpace(r.Name) == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "is required",
			Value:   r.Name,
		})
	} else if len(strings.TrimSpace(r.Name)) > 200 {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "cannot exceed 200 characters",
			Value:   len(r.Name),
		})
	}

	// Validate email
	if strings.TrimSpace(r.Email) == "" {
		errors = append(errors, ValidationError{
			Field:   "email",
			Message: "is required",
			Value:   r.Email,
		})
	} else if !emailPattern.MatchString(r.Email) {
		errors = append(errors, ValidationError{
			Field:   "email",
			Message: "must be a valid email address",
			Value:   r.Email,
		})
	}

	// Validate phone if provided
	if r.Phone != "" {
		phone := strings.TrimSpace(r.Phone)
		if len(phone) < 10 || len(phone) > 20 {
			errors = append(errors, ValidationError{
				Field:   "phone",
				Message: "must be between 10 and 20 characters",
				Value:   r.Phone,
			})
		}
	}

	// Validate address length if provided
	if r.Address != "" && len(strings.TrimSpace(r.Address)) > 500 {
		errors = append(errors, ValidationError{
			Field:   "address",
			Message: "cannot exceed 500 characters",
			Value:   len(r.Address),
		})
	}

	// Validate tax ID length if provided
	if r.TaxID != "" && len(strings.TrimSpace(r.TaxID)) > 50 {
		errors = append(errors, ValidationError{
			Field:   "tax_id",
			Message: "cannot exceed 50 characters",
			Value:   len(r.TaxID),
		})
	}

	if len(errors) > 0 {
		var messages []string
		for _, err := range errors {
			messages = append(messages, err.Error())
		}
		return fmt.Errorf("create client request validation failed: %s", strings.Join(messages, "; "))
	}

	return nil
}
