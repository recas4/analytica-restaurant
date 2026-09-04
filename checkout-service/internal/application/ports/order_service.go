package ports

import (
	"context"
)

// OrderService definiert das Interface fuer Bestellverarbeitungs-Operationen.
// Dies ist der primaere Port fuer die Geschaeftslogik.
type OrderService interface {
	// ProcessCheckout verarbeitet eine Checkout-Anfrage.
	// Empfaengt rohe Nachrichtendaten und erstellt eine Bestellung.
	ProcessCheckout(ctx context.Context, checkoutData []byte) error
}
