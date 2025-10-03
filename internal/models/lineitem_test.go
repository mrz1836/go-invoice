package models

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHourlyLineItem(t *testing.T) {
	ctx := context.Background()

	t.Run("ValidHourlyLineItem", func(t *testing.T) {
		date := time.Now()
		item, err := NewHourlyLineItem(ctx, "item-1", date, 8.0, 125.0, "Development work")

		require.NoError(t, err)
		assert.Equal(t, "item-1", item.ID)
		assert.Equal(t, LineItemTypeHourly, item.Type)
		assert.Equal(t, date, item.Date)
		assert.Equal(t, "Development work", item.Description)
		assert.NotNil(t, item.Hours)
		assert.InDelta(t, 8.0, *item.Hours, 1e-9)
		assert.NotNil(t, item.Rate)
		assert.InDelta(t, 125.0, *item.Rate, 1e-9)
		assert.InDelta(t, 1000.0, item.Total, 1e-9)
		assert.Nil(t, item.Amount)
		assert.Nil(t, item.Quantity)
		assert.Nil(t, item.UnitPrice)
	})

	t.Run("InvalidHoursZero", func(t *testing.T) {
		date := time.Now()
		_, err := NewHourlyLineItem(ctx, "item-1", date, 0.0, 125.0, "Development work")

		require.Error(t, err)
	})

	t.Run("InvalidHoursExceedLimit", func(t *testing.T) {
		date := time.Now()
		_, err := NewHourlyLineItem(ctx, "item-1", date, 25.0, 125.0, "Development work")

		require.Error(t, err)
	})

	t.Run("InvalidRateZero", func(t *testing.T) {
		date := time.Now()
		_, err := NewHourlyLineItem(ctx, "item-1", date, 8.0, 0.0, "Development work")

		require.Error(t, err)
	})
}

func TestNewFixedLineItem(t *testing.T) {
	ctx := context.Background()

	t.Run("ValidFixedLineItem", func(t *testing.T) {
		date := time.Now()
		item, err := NewFixedLineItem(ctx, "item-1", date, 2000.0, "Monthly Retainer")

		require.NoError(t, err)
		assert.Equal(t, "item-1", item.ID)
		assert.Equal(t, LineItemTypeFixed, item.Type)
		assert.Equal(t, date, item.Date)
		assert.Equal(t, "Monthly Retainer", item.Description)
		assert.NotNil(t, item.Amount)
		assert.InDelta(t, 2000.0, *item.Amount, 1e-9)
		assert.InDelta(t, 2000.0, item.Total, 1e-9)
		assert.Nil(t, item.Hours)
		assert.Nil(t, item.Rate)
		assert.Nil(t, item.Quantity)
		assert.Nil(t, item.UnitPrice)
	})

	t.Run("InvalidAmountZero", func(t *testing.T) {
		date := time.Now()
		_, err := NewFixedLineItem(ctx, "item-1", date, 0.0, "Monthly Retainer")

		require.Error(t, err)
	})

	t.Run("InvalidAmountNegative", func(t *testing.T) {
		date := time.Now()
		_, err := NewFixedLineItem(ctx, "item-1", date, -100.0, "Monthly Retainer")

		require.Error(t, err)
	})
}

func TestNewQuantityLineItem(t *testing.T) {
	ctx := context.Background()

	t.Run("ValidQuantityLineItem", func(t *testing.T) {
		date := time.Now()
		item, err := NewQuantityLineItem(ctx, "item-1", date, 3.0, 50.0, "SSL Certificates")

		require.NoError(t, err)
		assert.Equal(t, "item-1", item.ID)
		assert.Equal(t, LineItemTypeQuantity, item.Type)
		assert.Equal(t, date, item.Date)
		assert.Equal(t, "SSL Certificates", item.Description)
		assert.NotNil(t, item.Quantity)
		assert.InDelta(t, 3.0, *item.Quantity, 1e-9)
		assert.NotNil(t, item.UnitPrice)
		assert.InDelta(t, 50.0, *item.UnitPrice, 1e-9)
		assert.InDelta(t, 150.0, item.Total, 1e-9)
		assert.Nil(t, item.Hours)
		assert.Nil(t, item.Rate)
		assert.Nil(t, item.Amount)
	})

	t.Run("InvalidQuantityZero", func(t *testing.T) {
		date := time.Now()
		_, err := NewQuantityLineItem(ctx, "item-1", date, 0.0, 50.0, "SSL Certificates")

		require.Error(t, err)
	})

	t.Run("InvalidUnitPriceZero", func(t *testing.T) {
		date := time.Now()
		_, err := NewQuantityLineItem(ctx, "item-1", date, 3.0, 0.0, "SSL Certificates")

		require.Error(t, err)
	})
}

func TestLineItemValidation(t *testing.T) {
	ctx := context.Background()

	t.Run("HourlyItemWithFixedFields", func(t *testing.T) {
		date := time.Now()
		hours := 8.0
		rate := 125.0
		amount := 1000.0

		item := &LineItem{
			ID:          "item-1",
			Type:        LineItemTypeHourly,
			Date:        date,
			Description: "Test",
			Hours:       &hours,
			Rate:        &rate,
			Amount:      &amount, // Should not be set for hourly
			Total:       1000.0,
			CreatedAt:   time.Now(),
		}

		err := item.Validate(ctx)
		require.Error(t, err)
	})

	t.Run("FixedItemWithHourlyFields", func(t *testing.T) {
		date := time.Now()
		hours := 8.0
		rate := 125.0
		amount := 1000.0

		item := &LineItem{
			ID:          "item-1",
			Type:        LineItemTypeFixed,
			Date:        date,
			Description: "Test",
			Hours:       &hours, // Should not be set for fixed
			Rate:        &rate,  // Should not be set for fixed
			Amount:      &amount,
			Total:       1000.0,
			CreatedAt:   time.Now(),
		}

		err := item.Validate(ctx)
		require.Error(t, err)
	})

	t.Run("InvalidLineItemType", func(t *testing.T) {
		date := time.Now()

		item := &LineItem{
			ID:          "item-1",
			Type:        "invalid",
			Date:        date,
			Description: "Test",
			Total:       100.0,
			CreatedAt:   time.Now(),
		}

		err := item.Validate(ctx)
		require.Error(t, err)
	})
}

func TestLineItemRecalculateTotal(t *testing.T) {
	ctx := context.Background()

	t.Run("RecalculateHourlyTotal", func(t *testing.T) {
		date := time.Now()
		hours := 8.0
		rate := 125.0

		item := &LineItem{
			ID:          "item-1",
			Type:        LineItemTypeHourly,
			Date:        date,
			Description: "Test",
			Hours:       &hours,
			Rate:        &rate,
			Total:       0.0, // Wrong total
			CreatedAt:   time.Now(),
		}

		err := item.RecalculateTotal(ctx)
		require.NoError(t, err)
		assert.InDelta(t, 1000.0, item.Total, 1e-9)
	})

	t.Run("RecalculateFixedTotal", func(t *testing.T) {
		date := time.Now()
		amount := 2000.0

		item := &LineItem{
			ID:          "item-1",
			Type:        LineItemTypeFixed,
			Date:        date,
			Description: "Test",
			Amount:      &amount,
			Total:       0.0, // Wrong total
			CreatedAt:   time.Now(),
		}

		err := item.RecalculateTotal(ctx)
		require.NoError(t, err)
		assert.InDelta(t, 2000.0, item.Total, 1e-9)
	})

	t.Run("RecalculateQuantityTotal", func(t *testing.T) {
		date := time.Now()
		quantity := 3.0
		unitPrice := 50.0

		item := &LineItem{
			ID:          "item-1",
			Type:        LineItemTypeQuantity,
			Date:        date,
			Description: "Test",
			Quantity:    &quantity,
			UnitPrice:   &unitPrice,
			Total:       0.0, // Wrong total
			CreatedAt:   time.Now(),
		}

		err := item.RecalculateTotal(ctx)
		require.NoError(t, err)
		assert.InDelta(t, 150.0, item.Total, 1e-9)
	})
}

func TestLineItemGetDetails(t *testing.T) {
	ctx := context.Background()

	t.Run("HourlyDetails", func(t *testing.T) {
		date := time.Now()
		item, err := NewHourlyLineItem(ctx, "item-1", date, 8.0, 125.0, "Development")
		require.NoError(t, err)

		details := item.GetDetails()
		assert.Equal(t, "8.00 hours @ $125.00/hr", details)
	})

	t.Run("FixedDetails", func(t *testing.T) {
		date := time.Now()
		item, err := NewFixedLineItem(ctx, "item-1", date, 2000.0, "Retainer")
		require.NoError(t, err)

		details := item.GetDetails()
		assert.Equal(t, "Fixed amount", details)
	})

	t.Run("QuantityDetails", func(t *testing.T) {
		date := time.Now()
		item, err := NewQuantityLineItem(ctx, "item-1", date, 3.0, 50.0, "Licenses")
		require.NoError(t, err)

		details := item.GetDetails()
		assert.Equal(t, "3.00 × $50.00", details)
	})
}

func TestConvertWorkItemToLineItem(t *testing.T) {
	ctx := context.Background()

	t.Run("ValidConversion", func(t *testing.T) {
		date := time.Now()
		wi := WorkItem{
			ID:          "work-1",
			Date:        date,
			Hours:       8.0,
			Rate:        125.0,
			Description: "Development work",
			Total:       1000.0,
			CreatedAt:   time.Now(),
		}

		li, err := ConvertWorkItemToLineItem(ctx, wi)
		require.NoError(t, err)
		assert.Equal(t, wi.ID, li.ID)
		assert.Equal(t, LineItemTypeHourly, li.Type)
		assert.Equal(t, wi.Date, li.Date)
		assert.Equal(t, wi.Description, li.Description)
		assert.NotNil(t, li.Hours)
		assert.InDelta(t, wi.Hours, *li.Hours, 1e-9)
		assert.NotNil(t, li.Rate)
		assert.InDelta(t, wi.Rate, *li.Rate, 1e-9)
		assert.InDelta(t, wi.Total, li.Total, 1e-9)
		assert.Equal(t, wi.CreatedAt, li.CreatedAt)
	})
}

func TestLineItemUpdateDescription(t *testing.T) {
	ctx := context.Background()

	t.Run("ValidUpdate", func(t *testing.T) {
		date := time.Now()
		item, err := NewHourlyLineItem(ctx, "item-1", date, 8.0, 125.0, "Old description")
		require.NoError(t, err)

		err = item.UpdateDescription(ctx, "New description")
		require.NoError(t, err)
		assert.Equal(t, "New description", item.Description)
	})

	t.Run("EmptyDescription", func(t *testing.T) {
		date := time.Now()
		item, err := NewHourlyLineItem(ctx, "item-1", date, 8.0, 125.0, "Old description")
		require.NoError(t, err)

		err = item.UpdateDescription(ctx, "")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrDescriptionRequired)
	})

	t.Run("DescriptionTooLong", func(t *testing.T) {
		date := time.Now()
		item, err := NewHourlyLineItem(ctx, "item-1", date, 8.0, 125.0, "Old description")
		require.NoError(t, err)

		longDesc := string(make([]byte, 1001))
		err = item.UpdateDescription(ctx, longDesc)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrDescriptionTooLong)
	})
}
