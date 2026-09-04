<script setup lang="ts">
import { ref, computed } from 'vue'
import { useCartStore } from '@/stores/cart'
import { useAuthStore } from '@/stores/auth'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useRouter } from 'vue-router'
import { checkoutApi, type CheckoutRequest, type DeliveryInfo } from '@/api/checkout'
import { faker } from '@faker-js/faker/locale/de'

const cartStore = useCartStore()
const authStore = useAuthStore()
const router = useRouter()

const loading = ref(false)
const error = ref<string | null>(null)
const orderType = ref<'delivery' | 'pickup'>('delivery')

const deliveryInfo = ref<DeliveryInfo>({
  customer_name: '',
  customer_phone: '',
  street: '',
  house_number: '',
  postal_code: '',
  city: '',
  floor: '',
  instructions: ''
})

const isFormValid = computed(() => {
  if (orderType.value === 'pickup') return true
  return deliveryInfo.value.customer_name.trim() !== '' &&
         deliveryInfo.value.customer_phone.trim() !== '' &&
         deliveryInfo.value.street.trim() !== '' &&
         deliveryInfo.value.house_number.trim() !== '' &&
         deliveryInfo.value.postal_code.trim() !== '' &&
         deliveryInfo.value.city.trim() !== ''
})

function formatPrice(price: number) {
  return `€${price.toFixed(2)}`
}

function generateFakeData() {
  deliveryInfo.value = {
    customer_name: faker.person.fullName(),
    customer_phone: faker.phone.number({ style: 'international' }),
    street: faker.location.street(),
    house_number: faker.number.int({ min: 1, max: 150 }).toString(),
    postal_code: faker.location.zipCode('#####'),
    city: faker.helpers.arrayElement(['München', 'Berlin', 'Hamburg', 'Köln', 'Frankfurt']),
    floor: faker.helpers.maybe(() => `${faker.number.int({ min: 1, max: 5 })}. OG`, { probability: 0.5 }) || '',
    instructions: faker.helpers.maybe(() => faker.helpers.arrayElement([
      'Klingel 2x drücken',
      'Hintereingang nutzen',
      'Bei Nachbar abgeben falls nicht da',
      'Vor der Tür abstellen'
    ]), { probability: 0.4 }) || ''
  }
}

async function handleSubmitOrder() {
  if (!isFormValid.value) return

  loading.value = true
  error.value = null

  try {
    const items = cartStore.enrichedItems.map(item => ({
      product_id: item.product_id,
      product_name: item.product_name,
      quantity: item.quantity,
      unit_price: item.price,
      total_price: item.price * item.quantity
    }))

    const request: CheckoutRequest = {
      user_id: authStore.user?.id || 'unknown',
      items,
      total_amount: cartStore.total + (orderType.value === 'delivery' ? 2.50 : 0),
      currency: 'EUR',
      order_type: orderType.value,
      delivery_info: orderType.value === 'delivery' ? deliveryInfo.value : undefined
    }

    const result = await checkoutApi.checkout(request)

    // Redirect to payment page with order details
    router.push({
      path: '/payment',
      query: {
        order_id: result.order_id,
        amount: (cartStore.total + (orderType.value === 'delivery' ? 2.50 : 0)).toString(),
        currency: 'EUR'
      }
    })
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Bestellung fehlgeschlagen'
    console.error('Checkout failed:', e)
  } finally {
    loading.value = false
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
          <div class="flex gap-4">
            <Button @click="router.push('/cart')" class="bg-slate-800 hover:bg-slate-700 border border-cyan-500/30 text-cyan-400">
              Zurück zum Warenkorb
            </Button>
          </div>
        </div>
      </div>
    </div>

    <div class="container mx-auto px-4 pt-28 pb-12">
      <div class="mb-8">
        <h2 class="text-5xl font-bold bg-gradient-to-r from-white to-slate-300 bg-clip-text text-transparent mb-2 tracking-tight">
          BESTELLUNG
        </h2>
        <p class="text-cyan-400 font-mono text-sm">// schritt 1 von 2 - lieferdetails</p>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Form Section -->
        <div class="lg:col-span-2 space-y-6">
          <!-- Order Type -->
          <Card class="bg-slate-900/50 backdrop-blur-sm border border-cyan-500/20">
            <CardHeader>
              <CardTitle class="text-xl text-white font-mono">BESTELLART</CardTitle>
            </CardHeader>
            <CardContent class="space-y-3">
              <div
                @click="orderType = 'delivery'"
                :class="[
                  'p-4 rounded-lg border-2 cursor-pointer transition-all flex items-center gap-3',
                  orderType === 'delivery'
                    ? 'border-cyan-500 bg-cyan-500/10'
                    : 'border-slate-600 bg-slate-800/50 hover:border-slate-500'
                ]"
              >
                <span class="text-2xl">🚚</span>
                <div class="flex-1">
                  <span class="text-white font-medium">Lieferung</span>
                  <p class="text-sm text-slate-400">Direkt zu dir nach Hause</p>
                </div>
                <span class="text-cyan-400 font-mono">€2.50</span>
              </div>
              <div
                @click="orderType = 'pickup'"
                :class="[
                  'p-4 rounded-lg border-2 cursor-pointer transition-all flex items-center gap-3',
                  orderType === 'pickup'
                    ? 'border-cyan-500 bg-cyan-500/10'
                    : 'border-slate-600 bg-slate-800/50 hover:border-slate-500'
                ]"
              >
                <span class="text-2xl">📍</span>
                <div class="flex-1">
                  <span class="text-white font-medium">Abholung</span>
                  <p class="text-sm text-slate-400">Im Restaurant abholen</p>
                </div>
                <span class="text-green-400 font-mono">Kostenlos</span>
              </div>
            </CardContent>
          </Card>

          <!-- Delivery Address -->
          <Card v-if="orderType === 'delivery'" class="bg-slate-900/50 backdrop-blur-sm border border-cyan-500/20">
            <CardHeader class="flex flex-row items-center justify-between">
              <CardTitle class="text-xl text-white font-mono">LIEFERADRESSE</CardTitle>
              <Button
                @click="generateFakeData"
                variant="outline"
                size="sm"
                class="border-cyan-500/30 text-cyan-400 hover:bg-cyan-500/10"
                title="Testdaten generieren"
              >
                🪄
              </Button>
            </CardHeader>
            <CardContent class="space-y-4">
              <div class="grid grid-cols-2 gap-4">
                <div class="space-y-2">
                  <Label for="name" class="text-slate-300">Name *</Label>
                  <Input
                    id="name"
                    v-model="deliveryInfo.customer_name"
                    placeholder="Max Mustermann"
                    class="bg-slate-800 border-slate-600 text-white placeholder:text-slate-500 focus:border-cyan-500"
                  />
                </div>
                <div class="space-y-2">
                  <Label for="phone" class="text-slate-300">Telefon *</Label>
                  <Input
                    id="phone"
                    v-model="deliveryInfo.customer_phone"
                    placeholder="+49 123 456789"
                    class="bg-slate-800 border-slate-600 text-white placeholder:text-slate-500 focus:border-cyan-500"
                  />
                </div>
              </div>
              <div class="grid grid-cols-3 gap-4">
                <div class="col-span-2 space-y-2">
                  <Label for="street" class="text-slate-300">Straße *</Label>
                  <Input
                    id="street"
                    v-model="deliveryInfo.street"
                    placeholder="Musterstraße"
                    class="bg-slate-800 border-slate-600 text-white placeholder:text-slate-500 focus:border-cyan-500"
                  />
                </div>
                <div class="space-y-2">
                  <Label for="houseNumber" class="text-slate-300">Hausnr. *</Label>
                  <Input
                    id="houseNumber"
                    v-model="deliveryInfo.house_number"
                    placeholder="123"
                    class="bg-slate-800 border-slate-600 text-white placeholder:text-slate-500 focus:border-cyan-500"
                  />
                </div>
              </div>
              <div class="grid grid-cols-2 gap-4">
                <div class="space-y-2">
                  <Label for="zipCode" class="text-slate-300">PLZ *</Label>
                  <Input
                    id="zipCode"
                    v-model="deliveryInfo.postal_code"
                    placeholder="80331"
                    class="bg-slate-800 border-slate-600 text-white placeholder:text-slate-500 focus:border-cyan-500"
                  />
                </div>
                <div class="space-y-2">
                  <Label for="city" class="text-slate-300">Stadt *</Label>
                  <Input
                    id="city"
                    v-model="deliveryInfo.city"
                    placeholder="München"
                    class="bg-slate-800 border-slate-600 text-white placeholder:text-slate-500 focus:border-cyan-500"
                  />
                </div>
              </div>
              <div class="grid grid-cols-2 gap-4">
                <div class="space-y-2">
                  <Label for="floor" class="text-slate-300">Stockwerk</Label>
                  <Input
                    id="floor"
                    v-model="deliveryInfo.floor"
                    placeholder="2. OG"
                    class="bg-slate-800 border-slate-600 text-white placeholder:text-slate-500 focus:border-cyan-500"
                  />
                </div>
                <div class="space-y-2">
                  <Label for="instructions" class="text-slate-300">Hinweise</Label>
                  <Input
                    id="instructions"
                    v-model="deliveryInfo.instructions"
                    placeholder="Klingel 2x drücken"
                    class="bg-slate-800 border-slate-600 text-white placeholder:text-slate-500 focus:border-cyan-500"
                  />
                </div>
              </div>
            </CardContent>
          </Card>

          <!-- Pickup Info -->
          <Card v-if="orderType === 'pickup'" class="bg-slate-900/50 backdrop-blur-sm border border-cyan-500/20">
            <CardHeader>
              <CardTitle class="text-xl text-white font-mono">ABHOLUNG</CardTitle>
            </CardHeader>
            <CardContent class="space-y-2">
              <p class="text-slate-300">
                <span class="text-cyan-400 font-semibold">Analytica Restaurant</span><br>
                Musterstraße 123<br>
                80333 München
              </p>
              <p class="text-sm text-slate-400 mt-4">
                ⏱️ Bereit in ca. 30 Minuten nach Bestellung
              </p>
            </CardContent>
          </Card>
        </div>

        <!-- Order Summary -->
        <div class="lg:col-span-1">
          <Card class="bg-slate-900/50 backdrop-blur-sm border border-cyan-500/30 sticky top-24">
            <CardHeader>
              <CardTitle class="text-2xl text-white font-mono">BESTELLUNG</CardTitle>
            </CardHeader>
            <CardContent class="space-y-4">
              <div class="space-y-3 max-h-64 overflow-y-auto">
                <div
                  v-for="item in cartStore.enrichedItems"
                  :key="item.product_id"
                  class="flex justify-between text-sm border-b border-slate-700 pb-2"
                >
                  <div>
                    <span class="text-white">{{ item.product_name }}</span>
                    <span class="text-slate-400 ml-2">x{{ item.quantity }}</span>
                  </div>
                  <span class="text-cyan-400 font-mono">{{ formatPrice(item.price * item.quantity) }}</span>
                </div>
              </div>

              <div class="border-t border-cyan-500/20 pt-4 space-y-2">
                <div class="flex justify-between text-sm">
                  <span class="text-slate-400">Zwischensumme</span>
                  <span class="text-cyan-400 font-mono">{{ formatPrice(cartStore.total) }}</span>
                </div>
                <div class="flex justify-between text-sm">
                  <span class="text-slate-400">Lieferung</span>
                  <span class="text-cyan-400 font-mono">{{ orderType === 'delivery' ? '€2.50' : '€0.00' }}</span>
                </div>
                <div class="flex justify-between font-semibold text-lg pt-2 border-t border-slate-700">
                  <span class="text-white">Gesamt</span>
                  <span class="text-cyan-400 font-mono text-2xl">
                    {{ formatPrice(cartStore.total + (orderType === 'delivery' ? 2.50 : 0)) }}
                  </span>
                </div>
              </div>

              <div v-if="error" class="text-red-500 text-sm">
                {{ error }}
              </div>

              <Button
                @click="handleSubmitOrder"
                :disabled="loading || !isFormValid"
                class="w-full bg-gradient-to-r from-cyan-600 to-blue-600 hover:from-cyan-500 hover:to-blue-500 border border-cyan-500/50 shadow-lg shadow-cyan-500/30 disabled:opacity-50 disabled:cursor-not-allowed"
                size="lg"
              >
                {{ loading ? 'VERARBEITE...' : 'WEITER ZUR ZAHLUNG' }}
              </Button>

              <p class="text-xs text-slate-500 text-center">
                Schritt 1 von 2
              </p>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  </div>
</template>
