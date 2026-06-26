import client from './client'
import type { Currency } from '@/stores/currency'

export const getCurrenciesApi = () => {
  return client.get<Currency[]>('/currencies')
}

export const createCurrencyApi = (data: Partial<Currency>) => {
  return client.post<Currency>('/currencies', data)
}

export const updateCurrencyApi = (id: string, data: Partial<Currency>) => {
  return client.put<Currency>(`/currencies/${id}`, data)
}

export const deleteCurrencyApi = (id: string) => {
  return client.delete(`/currencies/${id}`)
}
