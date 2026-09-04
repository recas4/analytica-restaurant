<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useRouter, useRoute } from 'vue-router'
import {
  checkoutApi,
  type PaymentRequest,
  type StripePaymentDetails,
  type PayPalPaymentDetails,
  type BankTransferDetails
} from '@/api/checkout'
import { faker } from '@faker-js/faker/locale/de'

const authStore = useAuthStore()
const router = useRouter()
const route = useRoute()

const loading = ref(false)
const error = ref<string | null>(null)
const success = ref(false)

// Get order details from URL params
const orderId = computed(() => route.query.order_id as string || '')
const amount = computed(() => parseFloat(route.query.amount as string) || 0)
const currency = computed(() => route.query.currency as string || 'EUR')

const provider = ref<'stripe' | 'paypal' | 'bank_transfer'>('stripe')

const stripeData = ref<StripePaymentDetails>({
  card_number: '',
  expiry_month: '',
  expiry_year: '',
  cvv: '',
  cardholder_name: ''
})

const paypalData = ref<PayPalPaymentDetails>({
  email: '',
  password: ''
})

const bankData = ref<BankTransferDetails>({
  iban: '',
  account_holder: '',
  bank_name: ''
})

onMounted(() => {
  if (!orderId.value || amount.value <= 0) {
    router.push('/checkout')
  }
})

function formatPrice(price: number) {
  return `€${price.toFixed(2)}`
}

function generateFakeData() {
  switch (provider.value) {
    case 'stripe':
      stripeData.value = {
        cardholder_name: faker.person.fullName(),
        card_number: faker.finance.creditCardNumber('visa'),
        expiry_month: String(faker.number.int({ min: 1, max: 12 })).padStart(2, '0'),
        expiry_year: String(faker.date.future({ years: 3 }).getFullYear()).slice(-2),
        cvv: faker.finance.creditCardCVV()
      }
      break
    case 'paypal':
      paypalData.value = {
        email: faker.internet.email(),
        password: faker.internet.password({ length: 10 })
      }
      break
    case 'bank_transfer':
      bankData.value = {
        account_holder: faker.person.fullName(),
        iban: faker.finance.iban({ countryCode: 'DE', formatted: true }),
        bank_name: faker.helpers.arrayElement(['Deutsche Bank', 'Commerzbank', 'Sparkasse', 'DKB', 'ING'])
      }
      break
  }
}

const isFormValid = computed(() => {
  switch (provider.value) {
    case 'stripe':
      return stripeData.value.card_number !== '' &&
             stripeData.value.expiry_month !== '' &&
             stripeData.value.expiry_year !== '' &&
             stripeData.value.cvv !== '' &&
             stripeData.value.cardholder_name !== ''
    case 'paypal':
      return paypalData.value.email !== '' && paypalData.value.password !== ''
    case 'bank_transfer':
      return bankData.value.iban !== '' && bankData.value.account_holder !== ''
  }
})

async function handleSubmitPayment() {
  if (!isFormValid.value) return

  loading.value = true
  error.value = null

  try {
    let payment_details: StripePaymentDetails | PayPalPaymentDetails | BankTransferDetails

    switch (provider.value) {
      case 'stripe':
        payment_details = stripeData.value
        break
      case 'paypal':
        payment_details = paypalData.value
        break
      case 'bank_transfer':
        payment_details = bankData.value
        break
    }

    const request: PaymentRequest = {
      order_id: orderId.value,
      user_id: authStore.user?.id || 'unknown',
      provider: provider.value,
      amount: amount.value,
      currency: currency.value,
      payment_details
    }

    await checkoutApi.processPayment(request)

    success.value = true

    // Redirect to order status after 2.5 seconds
    setTimeout(() => {
      router.push({ name: 'order-status', params: { orderId: orderId.value } })
    }, 2500)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Zahlung fehlgeschlagen'
    console.error('Payment failed:', e)
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
            <Button @click="router.push('/checkout')" class="bg-slate-800 hover:bg-slate-700 border border-cyan-500/30 text-cyan-400">
              Zurück zur Bestellung
            </Button>
          </div>
        </div>
      </div>
    </div>

    <div class="container mx-auto px-4 pt-28 pb-12 max-w-2xl">
      <!-- Success State -->
      <div v-if="success" class="text-center space-y-6">
        <div class="text-6xl">✅</div>
        <h2 class="text-3xl font-bold text-green-400">Zahlung erfolgreich!</h2>
        <p class="text-slate-400">
          Deine Zahlung wurde verarbeitet. Du wirst zur Bestellverfolgung weitergeleitet...
        </p>
        <div class="bg-slate-900/50 border border-cyan-500/20 rounded-lg p-4">
          <p class="text-sm text-slate-400">Bestellung #{{ orderId.slice(-8) }}</p>
          <p class="text-2xl font-bold text-cyan-400 font-mono">{{ formatPrice(amount) }} bezahlt</p>
        </div>
      </div>

      <!-- Payment Form -->
      <template v-else>
        <div class="mb-8">
          <h2 class="text-5xl font-bold bg-gradient-to-r from-white to-slate-300 bg-clip-text text-transparent mb-2 tracking-tight">
            ZAHLUNG
          </h2>
          <p class="text-cyan-400 font-mono text-sm">// schritt 2 von 2 - zahlungsdetails</p>
        </div>

        <div class="space-y-6">
          <!-- Order Summary -->
          <Card class="bg-slate-900/50 backdrop-blur-sm border border-cyan-500/20">
            <CardHeader>
              <CardTitle class="text-xl text-white font-mono">BESTELLÜBERSICHT</CardTitle>
            </CardHeader>
            <CardContent>
              <div class="flex justify-between items-center">
                <div>
                  <p class="text-sm text-slate-400">Bestellung #{{ orderId.slice(-8) }}</p>
                </div>
                <p class="text-2xl font-bold text-cyan-400 font-mono">{{ formatPrice(amount) }}</p>
              </div>
            </CardContent>
          </Card>

          <!-- Payment Method -->
          <Card class="bg-slate-900/50 backdrop-blur-sm border border-cyan-500/20">
            <CardHeader>
              <CardTitle class="text-xl text-white font-mono">ZAHLUNGSMETHODE</CardTitle>
            </CardHeader>
            <CardContent class="space-y-3">
              <div
                @click="provider = 'stripe'"
                :class="[
                  'p-4 rounded-lg border-2 cursor-pointer transition-all flex items-center gap-3',
                  provider === 'stripe'
                    ? 'border-cyan-500 bg-cyan-500/10'
                    : 'border-slate-600 bg-slate-800/50 hover:border-slate-500'
                ]"
              >
                <span class="text-2xl">💳</span>
                <span class="text-white font-medium">Kreditkarte (Stripe)</span>
              </div>
              <div
                @click="provider = 'paypal'"
                :class="[
                  'p-4 rounded-lg border-2 cursor-pointer transition-all flex items-center gap-3',
                  provider === 'paypal'
                    ? 'border-cyan-500 bg-cyan-500/10'
                    : 'border-slate-600 bg-slate-800/50 hover:border-slate-500'
                ]"
              >
                <span class="text-2xl">🅿️</span>
                <span class="text-white font-medium">PayPal</span>
              </div>
              <div
                @click="provider = 'bank_transfer'"
                :class="[
                  'p-4 rounded-lg border-2 cursor-pointer transition-all flex items-center gap-3',
                  provider === 'bank_transfer'
                    ? 'border-cyan-500 bg-cyan-500/10'
                    : 'border-slate-600 bg-slate-800/50 hover:border-slate-500'
                ]"
              >
                <span class="text-2xl">🏦</span>
                <span class="text-white font-medium">Banküberweisung (SEPA)</span>
              </div>
            </CardContent>
          </Card>

          <!-- Payment Details -->
          <Card class="bg-slate-900/50 backdrop-blur-sm border border-cyan-500/20">
            <CardHeader class="flex flex-row items-center justify-between">
              <CardTitle class="text-xl text-white font-mono">ZAHLUNGSDETAILS</CardTitle>
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
              <!-- Stripe Form -->
              <template v-if="provider === 'stripe'">
                <div class="space-y-2">
                  <Label for="cardholder" class="text-slate-300">Karteninhaber *</Label>
                  <Input
                    id="cardholder"
                    v-model="stripeData.cardholder_name"
                    placeholder="Max Mustermann"
                    class="bg-slate-800 border-slate-600 text-white placeholder:text-slate-500 focus:border-cyan-500"
                  />
                </div>
                <div class="space-y-2">
                  <Label for="cardnumber" class="text-slate-300">Kartennummer *</Label>
                  <Input
                    id="cardnumber"
                    v-model="stripeData.card_number"
                    placeholder="4242 4242 4242 4242"
                    class="bg-slate-800 border-slate-600 text-white placeholder:text-slate-500 focus:border-cyan-500"
                  />
                </div>
                <div class="grid grid-cols-3 gap-4">
                  <div class="space-y-2">
                    <Label for="month" class="text-slate-300">Monat *</Label>
                    <Input
                      id="month"
                      v-model="stripeData.expiry_month"
                      placeholder="MM"
                      maxlength="2"
                      class="bg-slate-800 border-slate-600 text-white placeholder:text-slate-500 focus:border-cyan-500"
                    />
                  </div>
                  <div class="space-y-2">
                    <Label for="year" class="text-slate-300">Jahr *</Label>
                    <Input
                      id="year"
                      v-model="stripeData.expiry_year"
                      placeholder="YY"
                      maxlength="2"
                      class="bg-slate-800 border-slate-600 text-white placeholder:text-slate-500 focus:border-cyan-500"
                    />
                  </div>
                  <div class="space-y-2">
                    <Label for="cvv" class="text-slate-300">CVV *</Label>
                    <Input
                      id="cvv"
                      v-model="stripeData.cvv"
                      placeholder="123"
                      maxlength="3"
                      class="bg-slate-800 border-slate-600 text-white placeholder:text-slate-500 focus:border-cyan-500"
                    />
                  </div>
                </div>
              </template>

              <!-- PayPal Form -->
              <template v-if="provider === 'paypal'">
                <div class="space-y-2">
                  <Label for="paypal_email" class="text-slate-300">PayPal E-Mail *</Label>
                  <Input
                    id="paypal_email"
                    v-model="paypalData.email"
                    type="email"
                    placeholder="dein@paypal.de"
                    class="bg-slate-800 border-slate-600 text-white placeholder:text-slate-500 focus:border-cyan-500"
                  />
                </div>
                <div class="space-y-2">
                  <Label for="paypal_password" class="text-slate-300">PayPal Passwort *</Label>
                  <Input
                    id="paypal_password"
                    v-model="paypalData.password"
                    type="password"
                    placeholder="••••••••"
                    class="bg-slate-800 border-slate-600 text-white placeholder:text-slate-500 focus:border-cyan-500"
                  />
                </div>
              </template>

              <!-- Bank Transfer Form -->
              <template v-if="provider === 'bank_transfer'">
                <div class="space-y-2">
                  <Label for="account_holder" class="text-slate-300">Kontoinhaber *</Label>
                  <Input
                    id="account_holder"
                    v-model="bankData.account_holder"
                    placeholder="Max Mustermann"
                    class="bg-slate-800 border-slate-600 text-white placeholder:text-slate-500 focus:border-cyan-500"
                  />
                </div>
                <div class="space-y-2">
                  <Label for="iban" class="text-slate-300">IBAN *</Label>
                  <Input
                    id="iban"
                    v-model="bankData.iban"
                    placeholder="DE89 3704 0044 0532 0130 00"
                    class="bg-slate-800 border-slate-600 text-white placeholder:text-slate-500 focus:border-cyan-500"
                  />
                </div>
                <div class="space-y-2">
                  <Label for="bank_name" class="text-slate-300">Bank (optional)</Label>
                  <Input
                    id="bank_name"
                    v-model="bankData.bank_name"
                    placeholder="Deutsche Bank"
                    class="bg-slate-800 border-slate-600 text-white placeholder:text-slate-500 focus:border-cyan-500"
                  />
                </div>
              </template>
            </CardContent>
          </Card>

          <div v-if="error" class="text-red-500 text-sm bg-red-500/10 border border-red-500/30 rounded-lg p-3">
            {{ error }}
          </div>

          <Button
            @click="handleSubmitPayment"
            :disabled="loading || !isFormValid"
            class="w-full bg-gradient-to-r from-cyan-600 to-blue-600 hover:from-cyan-500 hover:to-blue-500 border border-cyan-500/50 shadow-lg shadow-cyan-500/30 disabled:opacity-50 disabled:cursor-not-allowed"
            size="lg"
          >
            {{ loading ? 'VERARBEITE...' : `JETZT BEZAHLEN - ${formatPrice(amount)}` }}
          </Button>

          <p class="text-xs text-slate-500 text-center">
            🔒 Deine Zahlungsdaten sind sicher verschlüsselt
          </p>
        </div>
      </template>
    </div>
  </div>
</template>
