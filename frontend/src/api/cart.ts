import { API_BASE_URL, getAuthToken } from './config'

export interface CartItem {
  ProductID: string
  Qty: number
}

export interface Cart {
  UserID: string
  Items: CartItem[]
}

export const cartApi = {
  async get(): Promise<Cart> {
    const token = getAuthToken()
    const response = await fetch(`${API_BASE_URL}/shopping/cart`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    if (!response.ok) throw new Error('Failed to fetch cart')
    return response.json()
  },

  async addItem(productId: string, quantity: number): Promise<void> {
    const token = getAuthToken()
    const response = await fetch(`${API_BASE_URL}/shopping/cart`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ product_id: productId, Qty: quantity }),
    })
    if (!response.ok) throw new Error('Failed to add to cart')
  },

  async updateItem(productId: string, quantity: number): Promise<void> {
    const token = getAuthToken()
    const response = await fetch(`${API_BASE_URL}/shopping/cart`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ product_id: productId, Qty: quantity }),
    })
    if (!response.ok) throw new Error('Failed to update cart')
  },

  async removeItem(productId: string): Promise<void> {
    const token = getAuthToken()
    const response = await fetch(`${API_BASE_URL}/shopping/cart/${productId}`, {
      method: 'DELETE',
      headers: { Authorization: `Bearer ${token}` },
    })
    if (!response.ok) throw new Error('Failed to remove from cart')
  },
}
