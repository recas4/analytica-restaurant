<script setup lang="ts">
import { onMounted } from 'vue'
import { useCartStore } from '@/stores/cart'
import { useAuthStore } from '@/stores/auth'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { useRouter } from 'vue-router'

const cartStore = useCartStore()
const authStore = useAuthStore()
const router = useRouter()

onMounted(() => {
  cartStore.fetchCart()
})

function formatPrice(price: number) {
  return `€${price.toFixed(2)}`
}

async function handleRemove(productId: string) {
  try {
    await cartStore.removeItem(productId)
  } catch (e) {
    console.error('Failed to remove item:', e)
  }
}

async function handleUpdateQuantity(productId: string, newQuantity: number) {
  if (newQuantity < 1) return
  try {
    await cartStore.updateItem(productId, newQuantity)
  } catch (e) {
    console.error('Failed to update quantity:', e)
  }
}

function handleCheckout() {
  router.push('/checkout')
}
</script>

<template>
  <div class="min-h-screen bg-slate-950">
    <div class="fixed top-0 left-0 right-0 bg-slate-900/80 backdrop-blur-md border-b border-cyan-500/20 z-10">
      <div class="container mx-auto px-4 py-4">
        <div class="flex justify-between items-center">
          <div>
            <h1 class="text-2xl font-bold bg-gradient-to-r from-cyan-400 to-blue-500 bg-clip-text text-transparent">
              ANALYTICA RESTAURANT
            </h1>
            <p class="text-sm text-slate-400 mt-0.5 font-mono">
              {{ authStore.user?.email }}
            </p>
          </div>
          <div class="flex gap-4">
            <Button @click="router.push('/products')" class="bg-slate-800 hover:bg-slate-700 border border-cyan-500/30 text-cyan-400">
              Continue Shopping
            </Button>
            <Button @click="authStore.logout" class="bg-slate-800 hover:bg-slate-700 border border-slate-600 text-slate-300">
              Logout
            </Button>
          </div>
        </div>
      </div>
    </div>

    <div class="container mx-auto px-4 pt-28 pb-12">
      <div class="mb-8">
        <h2 class="text-5xl font-bold bg-gradient-to-r from-white to-slate-300 bg-clip-text text-transparent mb-2 tracking-tight">
          SHOPPING CART
        </h2>
        <p class="text-cyan-400 font-mono text-sm">// review order data</p>
      </div>

      <div v-if="cartStore.loading" class="text-center py-12">
        <p class="text-gray-600">Loading cart...</p>
      </div>

      <div v-else-if="cartStore.error" class="text-center py-12">
        <p class="text-red-600">{{ cartStore.error }}</p>
        <Button @click="cartStore.fetchCart" class="mt-4">Retry</Button>
      </div>

      <div v-else-if="!cartStore.cart || cartStore.enrichedItems.length === 0" class="text-center py-12">
        <p class="text-gray-600 mb-4">Your cart is empty</p>
        <Button @click="router.push('/products')">Browse Products</Button>
      </div>

      <div v-else class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div class="lg:col-span-2 space-y-4">
          <Card v-for="item in cartStore.enrichedItems" :key="item.product_id" class="bg-slate-900/50 backdrop-blur-sm border border-cyan-500/20 hover:border-cyan-500/40 hover:shadow-lg hover:shadow-cyan-500/10 transition-all">
            <CardContent class="pt-6">
              <div class="flex justify-between items-center">
                <div class="flex-1">
                  <h3 class="font-semibold text-lg text-white font-mono">{{ item.product_name }}</h3>
                  <p class="text-cyan-400 mt-1 text-sm">{{ formatPrice(item.price) }} each</p>
                </div>

                <div class="flex items-center gap-4">
                  <div class="flex items-center gap-2">
                    <Button
                      size="sm"
                      variant="outline"
                      @click="handleUpdateQuantity(item.product_id, item.quantity - 1)"
                      :disabled="item.quantity <= 1"
                    >
                      -
                    </Button>
                    <span class="w-12 text-center text-white font-bold text-lg bg-slate-800 border border-cyan-500/30 rounded px-3 py-1">{{ item.quantity }}</span>
                    <Button
                      size="sm"
                      variant="outline"
                      @click="handleUpdateQuantity(item.product_id, item.quantity + 1)"
                    >
                      +
                    </Button>
                  </div>

                  <div class="w-24 text-right font-semibold text-cyan-400 font-mono">
                    {{ formatPrice(item.price * item.quantity) }}
                  </div>

                  <Button
                    variant="destructive"
                    size="sm"
                    @click="handleRemove(item.product_id)"
                  >
                    Remove
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        <div class="lg:col-span-1">
          <Card class="bg-slate-900/50 backdrop-blur-sm border border-cyan-500/30 sticky top-24">
            <CardHeader>
              <CardTitle class="text-2xl text-white font-mono">ORDER SUMMARY</CardTitle>
            </CardHeader>
            <CardContent class="space-y-4">
              <div class="flex justify-between text-sm">
                <span class="text-slate-400">Items ({{ cartStore.itemCount }})</span>
                <span class="text-cyan-400 font-mono">{{ formatPrice(cartStore.total) }}</span>
              </div>

              <div class="border-t border-cyan-500/20 pt-4">
                <div class="flex justify-between font-semibold text-lg">
                  <span class="text-white">Total</span>
                  <span class="text-cyan-400 font-mono text-2xl">{{ formatPrice(cartStore.total) }}</span>
                </div>
              </div>

              <Button
                @click="handleCheckout"
                class="w-full bg-gradient-to-r from-cyan-600 to-blue-600 hover:from-cyan-500 hover:to-blue-500 border border-cyan-500/50 shadow-lg shadow-cyan-500/30"
                size="lg"
              >
                CHECKOUT
              </Button>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  </div>
</template>
