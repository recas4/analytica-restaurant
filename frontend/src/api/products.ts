import { API_BASE_URL, getAuthToken } from './config'

export interface Product {
  id: string
  name: string
  description?: string
  price: number
  category?: string
  image_url?: string
  user_id: string
}

export const productsApi = {
  async getAll(): Promise<Product[]> {
    const response = await fetch(`${API_BASE_URL}/shopping/products`)
    if (!response.ok) throw new Error('Failed to fetch products')
    return response.json()
  },

  async create(product: Omit<Product, 'id' | 'user_id'>): Promise<Product> {
    const token = getAuthToken()
    const response = await fetch(`${API_BASE_URL}/shopping/products`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(product),
    })
    if (!response.ok) throw new Error('Failed to create product')
    return response.json()
  },
}
