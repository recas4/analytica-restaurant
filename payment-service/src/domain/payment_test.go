package domain

import (
	"testing"
)

func TestNewPayment_Success(t *testing.T) {
	orderID := "order-123"
	userID := "user-456"

	payment, err := NewPayment(orderID, userID, "stripe", 99.99, "EUR")

	if err != nil {
		t.Errorf("NewPayment should succeed, got error: %v", err)
	}

	if payment == nil {
		t.Fatal("Payment should not be nil")
	}

	if payment.OrderID != orderID {
		t.Errorf("Expected OrderID %s, got %s", orderID, payment.OrderID)
	}

	if payment.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, payment.UserID)
	}

	if payment.Provider != "stripe" {
		t.Errorf("Expected provider 'stripe', got '%s'", payment.Provider)
	}

	if payment.Amount != 99.99 {
		t.Errorf("Expected amount 99.99, got %f", payment.Amount)
	}

	if payment.Currency != "EUR" {
		t.Errorf("Expected currency 'EUR', got '%s'", payment.Currency)
	}

	if payment.Status != StatusPending {
		t.Errorf("Expected status 'pending', got '%s'", payment.Status)
	}
}

func TestNewPayment_InvalidAmount(t *testing.T) {
	orderID := "order-123"
	userID := "user-456"

	testCases := []float64{0, -1, -100.50}

	for _, amount := range testCases {
		payment, err := NewPayment(orderID, userID, "stripe", amount, "EUR")

		if err == nil {
			t.Errorf("NewPayment with amount %f should fail", amount)
		}

		if payment != nil {
			t.Errorf("Payment should be nil for invalid amount %f", amount)
		}
	}
}

func TestPayment_MarkSuccess(t *testing.T) {
	orderID := "order-123"
	userID := "user-456"
	payment, _ := NewPayment(orderID, userID, "stripe", 50.00, "EUR")

	err := payment.MarkSuccess()

	if err != nil {
		t.Errorf("MarkSuccess should succeed, got error: %v", err)
	}

	if payment.Status != StatusSuccess {
		t.Errorf("Expected status 'success', got '%s'", payment.Status)
	}
}

func TestPayment_MarkSuccess_AlreadySuccess(t *testing.T) {
	orderID := "order-123"
	userID := "user-456"
	payment, _ := NewPayment(orderID, userID, "stripe", 50.00, "EUR")

	_ = payment.MarkSuccess()
	err := payment.MarkSuccess()

	if err == nil {
		t.Error("MarkSuccess on already successful payment should fail")
	}
}

func TestPayment_MarkFailed(t *testing.T) {
	orderID := "order-123"
	userID := "user-456"
	payment, _ := NewPayment(orderID, userID, "stripe", 50.00, "EUR")

	err := payment.MarkFailed()

	if err != nil {
		t.Errorf("MarkFailed should succeed, got error: %v", err)
	}

	if payment.Status != StatusFailed {
		t.Errorf("Expected status 'failed', got '%s'", payment.Status)
	}
}

func TestPayment_MarkFailed_AlreadyFailed(t *testing.T) {
	orderID := "order-123"
	userID := "user-456"
	payment, _ := NewPayment(orderID, userID, "stripe", 50.00, "EUR")

	_ = payment.MarkFailed()
	err := payment.MarkFailed()

	if err == nil {
		t.Error("MarkFailed on already failed payment should fail")
	}
}

func TestPayment_MarkFailed_AfterSuccess(t *testing.T) {
	orderID := "order-123"
	userID := "user-456"
	payment, _ := NewPayment(orderID, userID, "stripe", 50.00, "EUR")

	_ = payment.MarkSuccess()
	err := payment.MarkFailed()

	if err == nil {
		t.Error("MarkFailed on successful payment should fail")
	}
}

func TestStatus_Constants(t *testing.T) {
	if StatusPending != "pending" {
		t.Errorf("StatusPending should be 'pending', got '%s'", StatusPending)
	}

	if StatusSuccess != "success" {
		t.Errorf("StatusSuccess should be 'success', got '%s'", StatusSuccess)
	}

	if StatusFailed != "failed" {
		t.Errorf("StatusFailed should be 'failed', got '%s'", StatusFailed)
	}
}
