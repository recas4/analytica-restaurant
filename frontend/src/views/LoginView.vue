<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'

const authStore = useAuthStore()
const router = useRouter()

const email = ref('')
const password = ref('')
const isRegister = ref(false)

async function handleSubmit() {
  if (isRegister.value) {
    const success = await authStore.register(email.value, password.value, 'admin')
    if (success) router.push('/products')
  } else {
    const success = await authStore.login(email.value, password.value)
    if (success) router.push('/products')
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-slate-950 px-4">
    <Card class="w-full max-w-md bg-slate-900/50 backdrop-blur-sm border border-cyan-500/30 shadow-2xl shadow-cyan-500/20">
      <CardHeader class="space-y-3">
        <div class="text-center">
          <h1 class="text-4xl font-bold bg-gradient-to-r from-cyan-400 to-blue-500 bg-clip-text text-transparent mb-2">
            ANALYTICA
          </h1>
          <p class="text-slate-400 text-sm font-mono">// cloud kitchen interface</p>
        </div>
        <CardTitle class="text-center text-2xl text-white font-mono">
          {{ isRegister ? 'REGISTER' : 'LOGIN' }}
        </CardTitle>
        <CardDescription class="text-center text-slate-400">
          {{ isRegister ? 'Initialize new user credentials' : 'Authenticate to continue' }}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form @submit.prevent="handleSubmit" class="space-y-4">
          <div class="space-y-2">
            <Label for="email" class="text-slate-300">Email</Label>
            <Input
              id="email"
              v-model="email"
              type="email"
              placeholder="you@example.com"
              class="text-white bg-slate-800/50 border-cyan-500/30 focus-visible:border-cyan-500"
              required
            />
          </div>
          <div class="space-y-2">
            <Label for="password" class="text-slate-300">Password</Label>
            <Input
              id="password"
              v-model="password"
              type="password"
              placeholder="••••••••"
              class="text-white bg-slate-800/50 border-cyan-500/30 focus-visible:border-cyan-500"
              required
            />
          </div>

          <div v-if="authStore.error" class="text-sm text-red-600">
            {{ authStore.error }}
          </div>

          <Button
            type="submit"
            class="w-full bg-gradient-to-r from-cyan-600 to-blue-600 hover:from-cyan-500 hover:to-blue-500 border border-cyan-500/50 shadow-lg shadow-cyan-500/30 font-mono"
            :disabled="authStore.loading"
          >
            {{ authStore.loading ? 'PROCESSING...' : (isRegister ? 'REGISTER' : 'LOGIN') }}
          </Button>

          <div class="text-center text-sm">
            <button
              type="button"
              @click="isRegister = !isRegister"
              class="text-blue-600 hover:underline"
            >
              {{ isRegister ? 'Already have an account? Login' : "Don't have an account? Register" }}
            </button>
          </div>
        </form>
      </CardContent>
    </Card>
  </div>
</template>
