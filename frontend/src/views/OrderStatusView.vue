<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { checkoutApi, type OrderStatus, type DeliveryStatus } from '@/api/checkout'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const orderId = route.params.orderId as string
const orderStatus = ref<OrderStatus | null>(null)
const deliveryStatus = ref<DeliveryStatus | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)

let pollInterval: number | null = null

const kitchenSteps = [
  { key: 'received', label: 'Bestellt', icon: '📋' },
  { key: 'preparing', label: 'Zubereitung', icon: '👨‍🍳' },
  { key: 'ready', label: 'Fertig', icon: '✅' },
]

const deliverySteps = [
  { key: 'assigned', label: 'Fahrer zugewiesen', icon: '🚗' },
  { key: 'in_transit', label: 'Unterwegs', icon: '🚚' },
  { key: 'delivered', label: 'Geliefert', icon: '🏠' },
]

const isDeliveryOrder = computed(() => deliveryStatus.value?.order_type === 'delivery')
const isPickupOrder = computed(() => deliveryStatus.value?.order_type === 'pickup')

function getKitchenStatusIndex(status: string): number {
  const statusMap: Record<string, number> = {
    'received': 0,
    'pending': 0,
    'preparing': 1,
    'ready': 2,
  }
  return statusMap[status] ?? 0
}

function getDeliveryStatusIndex(status: string): number {
  const statusMap: Record<string, number> = {
    'ready_to_deliver': 0,
    'delivering': 0,
    'in_transit': 1,
    'delivered': 2,
    'completed': 2,
  }
  return statusMap[status] ?? -1
}

async function fetchStatus() {
  try {
    const [kitchen, delivery] = await Promise.all([
      checkoutApi.getOrderStatus(orderId).catch(() => null),
      checkoutApi.getDeliveryStatus(orderId),
    ])

    if (kitchen) orderStatus.value = kitchen
    if (delivery) deliveryStatus.value = delivery

    if (!kitchen && !delivery) {
      error.value = 'Bestellung nicht gefunden'
    } else {
      error.value = null
    }
  } catch (e) {
    error.value = 'Fehler beim Laden des Status'
    console.error('Failed to fetch status:', e)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchStatus()
  pollInterval = window.setInterval(fetchStatus, 5000)
})

onUnmounted(() => {
  if (pollInterval) {
    clearInterval(pollInterval)
  }
})
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
              Zurück zum Menü
            </Button>
          </div>
        </div>
      </div>
    </div>

    <div class="container mx-auto px-4 pt-28 pb-12">
      <div class="mb-8">
        <h2 class="text-5xl font-bold bg-gradient-to-r from-white to-slate-300 bg-clip-text text-transparent mb-2 tracking-tight">
          BESTELLSTATUS
        </h2>
        <p class="text-cyan-400 font-mono text-sm">// {{ orderId }}</p>
      </div>

      <div v-if="loading" class="text-center py-12">
        <p class="text-slate-400">Lade Bestellstatus...</p>
      </div>

      <div v-else-if="error" class="text-center py-12">
        <Card class="bg-slate-900/50 border border-cyan-500/20 max-w-md mx-auto">
          <CardContent class="pt-6">
            <p class="text-cyan-400 font-mono mb-4">{{ error }}</p>
            <p class="text-slate-400 text-sm mb-4">Bestellung wird möglicherweise noch verarbeitet...</p>
            <Button @click="fetchStatus" class="bg-cyan-600 hover:bg-cyan-500">
              Erneut versuchen
            </Button>
          </CardContent>
        </Card>
      </div>

      <div v-else class="max-w-2xl mx-auto space-y-6">
        <!-- Order Type Badge -->
        <div class="text-center">
          <span
            class="inline-flex items-center gap-2 px-4 py-2 rounded-full text-sm font-mono"
            :class="isDeliveryOrder ? 'bg-blue-500/20 text-blue-400 border border-blue-500/30' : 'bg-green-500/20 text-green-400 border border-green-500/30'"
          >
            {{ isDeliveryOrder ? '🚚 Lieferung' : '📍 Abholung' }}
          </span>
        </div>

        <!-- Kitchen Status Card -->
        <Card class="bg-slate-900/50 backdrop-blur-sm border border-cyan-500/30">
          <CardHeader>
            <CardTitle class="text-xl text-white font-mono">🍳 KÜCHE</CardTitle>
          </CardHeader>
          <CardContent>
            <div class="flex justify-between items-center">
              <div
                v-for="(step, index) in kitchenSteps"
                :key="step.key"
                class="flex flex-col items-center flex-1"
              >
                <div
                  class="w-12 h-12 rounded-full flex items-center justify-center text-2xl mb-2 transition-all"
                  :class="index <= getKitchenStatusIndex(orderStatus?.status || 'pending')
                    ? 'bg-gradient-to-r from-cyan-600 to-blue-600 shadow-lg shadow-cyan-500/50'
                    : 'bg-slate-800 border border-slate-700'"
                >
                  {{ step.icon }}
                </div>
                <span
                  class="text-xs font-mono text-center"
                  :class="index <= getKitchenStatusIndex(orderStatus?.status || 'pending')
                    ? 'text-cyan-400'
                    : 'text-slate-500'"
                >
                  {{ step.label }}
                </span>
              </div>
            </div>
          </CardContent>
        </Card>

        <!-- Delivery Status Card -->
        <Card v-if="isDeliveryOrder" class="bg-slate-900/50 backdrop-blur-sm border border-blue-500/30">
          <CardHeader>
            <CardTitle class="text-xl text-white font-mono">🚚 LIEFERUNG</CardTitle>
          </CardHeader>
          <CardContent>
            <div class="flex justify-between items-center">
              <div
                v-for="(step, index) in deliverySteps"
                :key="step.key"
                class="flex flex-col items-center flex-1"
              >
                <div
                  class="w-12 h-12 rounded-full flex items-center justify-center text-2xl mb-2 transition-all"
                  :class="index <= getDeliveryStatusIndex(deliveryStatus?.status || '')
                    ? 'bg-gradient-to-r from-blue-600 to-purple-600 shadow-lg shadow-blue-500/50'
                    : 'bg-slate-800 border border-slate-700'"
                >
                  {{ step.icon }}
                </div>
                <span
                  class="text-xs font-mono text-center"
                  :class="index <= getDeliveryStatusIndex(deliveryStatus?.status || '')
                    ? 'text-blue-400'
                    : 'text-slate-500'"
                >
                  {{ step.label }}
                </span>
              </div>
            </div>

            <!-- Driver Info -->
            <div v-if="deliveryStatus?.driver_id" class="mt-6 pt-4 border-t border-blue-500/20">
              <p class="text-slate-400 text-sm">Fahrer</p>
              <p class="text-blue-400 font-mono">{{ deliveryStatus.driver_id }}</p>
            </div>

            <!-- Delivery Address -->
            <div v-if="deliveryStatus?.delivery_info" class="mt-4 pt-4 border-t border-blue-500/20">
              <p class="text-slate-400 text-sm mb-1">Lieferadresse</p>
              <p class="text-white">
                {{ deliveryStatus.delivery_info.customer_name }}<br>
                {{ deliveryStatus.delivery_info.street }} {{ deliveryStatus.delivery_info.house_number }}<br>
                {{ deliveryStatus.delivery_info.postal_code }} {{ deliveryStatus.delivery_info.city }}
              </p>
            </div>
          </CardContent>
        </Card>

        <!-- Pickup Status Card -->
        <Card v-if="isPickupOrder" class="bg-slate-900/50 backdrop-blur-sm border border-green-500/30">
          <CardHeader>
            <CardTitle class="text-xl text-white font-mono">📍 ABHOLUNG</CardTitle>
          </CardHeader>
          <CardContent>
            <div class="text-center py-4">
              <div
                class="w-20 h-20 mx-auto rounded-full flex items-center justify-center text-4xl mb-4 transition-all"
                :class="deliveryStatus?.status === 'pickup_ready' || orderStatus?.status === 'ready'
                  ? 'bg-gradient-to-r from-green-600 to-emerald-600 shadow-lg shadow-green-500/50'
                  : 'bg-slate-800 border border-slate-700'"
              >
                🏪
              </div>
              <p
                class="text-lg font-mono"
                :class="deliveryStatus?.status === 'pickup_ready' || orderStatus?.status === 'ready'
                  ? 'text-green-400'
                  : 'text-slate-400'"
              >
                {{ deliveryStatus?.status === 'pickup_ready' || orderStatus?.status === 'ready'
                  ? 'Bereit zur Abholung!'
                  : 'Wird zubereitet...' }}
              </p>
            </div>

            <div class="mt-4 pt-4 border-t border-green-500/20 text-center">
              <p class="text-slate-400 text-sm">Abholadresse</p>
              <p class="text-white font-medium">Analytica Restaurant</p>
              <p class="text-slate-300 text-sm">Musterstraße 123, 80333 München</p>
            </div>
          </CardContent>
        </Card>

        <!-- Auto-refresh notice -->
        <div class="text-center text-slate-500 text-xs font-mono">
          // automatische Aktualisierung alle 5s
        </div>

        <!-- New Order Button -->
        <div class="text-center mt-8">
          <Button
            @click="router.push('/products')"
            class="bg-gradient-to-r from-cyan-600 to-blue-600 hover:from-cyan-500 hover:to-blue-500"
          >
            Neue Bestellung
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>
