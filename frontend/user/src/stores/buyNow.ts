import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { CartItem } from './cart'

export const useBuyNowStore = defineStore('buyNow', () => {
    const item = ref<CartItem | null>(null)
    const plusHalfSession = ref('')

    const hasItem = computed(() => item.value !== null)

    const setItem = (newItem: CartItem) => {
        item.value = { ...newItem }
    }

    const clear = () => {
        item.value = null
        plusHalfSession.value = ''
    }

    const setPlusHalfSession = (session: string) => { plusHalfSession.value = session }

    return {
        item,
        hasItem,
        setItem,
        clear,
        plusHalfSession,
        setPlusHalfSession,
    }
})
