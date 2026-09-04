package domain

// Cart repraesentiert den Warenkorb eines Benutzers.
// Enthaelt die Benutzer-ID und eine Liste von Warenkorb-Artikeln.
type Cart struct {
	UserID string
	Items  []CartItem
}

// CartItem repraesentiert einen einzelnen Artikel im Warenkorb.
type CartItem struct {
	ProductID string
	Qty       int
}

// NewCart erstellt einen neuen leeren Warenkorb fuer den Benutzer.
func NewCart(userID string) *Cart {
	return &Cart{
		UserID: userID,
		Items:  []CartItem{},
	}
}

// AddItem fuegt einen Artikel zum Warenkorb hinzu.
func (c *Cart) AddItem(productID string, qty int) {
	c.Items = append(c.Items, CartItem{
		ProductID: productID,
		Qty:       qty,
	})
}

// IsEmpty prueft ob der Warenkorb leer ist.
func (c *Cart) IsEmpty() bool {
	return len(c.Items) == 0
}

// TotalItems gibt die Gesamtanzahl der Artikel zurueck.
func (c *Cart) TotalItems() int {
	total := 0
	for _, item := range c.Items {
		total += item.Qty
	}
	return total
}
