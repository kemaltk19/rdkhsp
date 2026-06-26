<script setup lang="ts">
import { computed } from 'vue'
import { useCurrencyStore } from '@/stores/currency'

const props = defineProps({
  value: {
    type: [Number, String],
    default: 0,
  },
  currency: {
    type: String,
    default: null, // Allow falling back to default
  },
})

const currencyStore = useCurrencyStore()

import Decimal from 'decimal.js'

const formatted = computed(() => {
  const raw = props.value || '0'
  let d: Decimal;
  try {
    d = new Decimal(raw)
  } catch(e) {
    d = new Decimal(0)
  }

  // Find the exact currency or the default one
  const currDef = props.currency 
    ? currencyStore.getCurrencyByCode(props.currency) 
    : currencyStore.defaultCurrency

  const decimals = currDef ? (currDef.format_decimals || 2) : 2
  const decimalSep = currDef ? (currDef.format_decimal_sep || ',') : ','
  const thousandSep = currDef ? (currDef.format_thousand_sep || '.') : '.'
  const symbol = currDef ? (currDef.symbol || '₺') : '₺'
  const position = currDef ? (currDef.format_position || 'RightSpace') : 'RightSpace'

  // Format the number manually using Decimal.js fixed length without standard JS exponentiation loss
  const valStr = d.toFixed(decimals)
  const parts = valStr.split('.')
  // Add thousand separators manually to integer part
  parts[0] = parts[0].replace(/\B(?=(\d{3})+(?!\d))/g, thousandSep)
  const formattedVal = parts.join(decimalSep)
  
  switch(position) {
    case 'Left': return `${symbol}${formattedVal}`
    case 'LeftSpace': return `${symbol} ${formattedVal}`
    case 'Right': return `${formattedVal}${symbol}`
    case 'RightSpace': return `${formattedVal} ${symbol}`
    default: return `${formattedVal} ${symbol}`
  }
})
</script>

<template>
  <span class="tabular-nums">{{ formatted }}</span>
</template>
