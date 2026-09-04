<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { productsApi, type Product } from '@/api/products'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { useAuthStore } from '@/stores/auth'
import { useCartStore } from '@/stores/cart'
import { useRouter } from 'vue-router'
import { ShoppingCart } from 'lucide-vue-next'

const authStore = useAuthStore()
const cartStore = useCartStore()
const router = useRouter()
const products = ref<Product[]>([])
const loading = ref(true)
const error = ref<string | null>(null)

async function loadProducts() {
  loading.value = true
  error.value = null
  try {
    products.value = await productsApi.getAll()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load products'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadProducts()
  cartStore.fetchCart()
})

function formatPrice(price: number) {
  return `€${price.toFixed(2)}`
}

async function addToCart(productId: string) {
  try {
    await cartStore.addItem(productId, 1)
  } catch (e) {
    console.error('Failed to add to cart:', e)
  }
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
          <div class="flex gap-2">
            <Button @click="router.push('/cart')" class="relative bg-slate-800 hover:bg-slate-700 border border-cyan-500/30 hover:border-cyan-500/50 text-cyan-400">
              <ShoppingCart class="h-4 w-4" />
              <Badge v-if="cartStore.itemCount > 0" class="absolute -top-2 -right-2 h-5 w-5 flex items-center justify-center p-0 bg-gradient-to-r from-cyan-500 to-blue-500 shadow-lg shadow-cyan-500/50 animate-pulse">
                {{ cartStore.itemCount }}
              </Badge>
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
          DATA MENU
        </h2>
        <p class="text-cyan-400 font-mono text-sm">// algorithmic culinary experiences</p>
      </div>

      <div v-if="loading" class="text-center py-12">
        <p class="text-gray-600">Loading products...</p>
      </div>

      <div v-else-if="error" class="text-center py-12">
        <p class="text-red-600">{{ error }}</p>
        <Button @click="loadProducts" class="mt-4">Retry</Button>
      </div>

      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <Card v-for="product in products" :key="product.id" class="group bg-slate-900/50 backdrop-blur-sm border border-cyan-500/20 hover:border-cyan-500/50 hover:shadow-2xl hover:shadow-cyan-500/20 hover:-translate-y-1 transition-all duration-300 overflow-hidden relative">
          <div class="absolute inset-0 bg-gradient-to-br from-cyan-500/5 to-blue-500/5 opacity-0 group-hover:opacity-100 transition-opacity"></div>
          <div class="h-1 bg-gradient-to-r from-cyan-500 to-blue-500 transform scale-x-0 group-hover:scale-x-100 transition-transform duration-300"></div>
          <CardHeader class="relative">
            <CardTitle class="text-xl text-white group-hover:text-cyan-400 transition-colors font-mono">
              {{ product.name }}
            </CardTitle>
            <CardDescription v-if="product.description" class="text-slate-400">
              {{ product.description }}
            </CardDescription>
          </CardHeader>
          <CardContent class="relative">
            <div class="flex justify-between items-center">
              <span class="text-3xl font-bold bg-gradient-to-r from-cyan-400 to-blue-400 bg-clip-text text-transparent font-mono">
                {{ formatPrice(product.price) }}
              </span>
              <Button @click="addToCart(product.id)" class="bg-gradient-to-r from-cyan-600 to-blue-600 hover:from-cyan-500 hover:to-blue-500 border border-cyan-500/50 shadow-lg shadow-cyan-500/30">
                ADD
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>

      <div v-if="!loading && products.length === 0" class="text-center py-12">
        <p class="text-gray-600">No products available</p>
      </div>
    </div>
  </div>
</template>
