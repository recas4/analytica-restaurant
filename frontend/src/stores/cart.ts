import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { cartApi, type Cart } from '@/api/cart'
import { productsApi } from '@/api/products'

export interface EnrichedCartItem {
  product_id: string
  product_name: string
  quantity: number
  price: number
}

export const useCartStore = defineStore('cart', () => {
  const cart = ref<Cart | null>(null)
  const enrichedItems = ref<EnrichedCartItem[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const itemCount = computed(() => {
    if (!cart.value) return 0
    return cart.value.Items.reduce((sum, item) => sum + item.Qty, 0)
  })

  const total = computed(() => {
    return enrichedItems.value.reduce((sum, item) => sum + (item.price * item.quantity), 0)
  })

  async function fetchCart() {
    loading.value = true
    error.value = null
    try {
      cart.value = await cartApi.get()

      const products = await productsApi.getAll()
      enrichedItems.value = cart.value.Items.map(item => {
        const product = products.find(p => p.id === item.ProductID)
        return {
          product_id: item.ProductID,
          product_name: product?.name ?? 'Unknown Product',
          quantity: item.Qty,
          price: product?.price ?? 0
        }
      })
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch cart'
      cart.value = null
      enrichedItems.value = []
    } finally {
      loading.value = false
    }
  }

  async function addItem(productId: string, quantity: number = 1) {
    error.value = null
    try {
      await cartApi.addItem(productId, quantity)
      await fetchCart()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to add item'
      throw e
    }
  }

  async function updateItem(productId: string, quantity: number) {
    error.value = null
    try {
      await cartApi.updateItem(productId, quantity)
      await fetchCart()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to update item'
      throw e
    }
  }

  async function removeItem(productId: string) {
    error.value = null
    try {
      await cartApi.removeItem(productId)
      await fetchCart()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to remove item'
      throw e
    }
  }

  return {
    cart,
    enrichedItems,
    loading,
    error,
    itemCount,
    total,
    fetchCart,
    addItem,
    updateItem,
    removeItem,
  }
})
