import { useAuthStore } from '@/stores/auth'
import { useSettingsStore } from '@/stores/settings'

export function getCompanyTimezone(): string {
  const authStore = useAuthStore()
  const settingsStore = useSettingsStore()
  
  return settingsStore.settings['timezone'] 
    || authStore.company?.timezone 
    || Intl.DateTimeFormat().resolvedOptions().timeZone 
    || 'Europe/Istanbul'
}

export function formatDate(dateStr: string | Date | null | undefined): string {
  return formatDateTime(dateStr)
}

export function formatDateTime(dateStr: string | Date | null | undefined): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  try {
    return date.toLocaleString('tr-TR', {
      timeZone: getCompanyTimezone(),
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      hour12: false
    })
  } catch (e) {
    return date.toLocaleString('tr-TR', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      hour12: false
    })
  }
}

export function getCurrentCompanyDatetimeLocal(addDays: number = 0): string {
  const d = new Date()
  if (addDays) d.setDate(d.getDate() + addDays)
  
  try {
    const formatter = new Intl.DateTimeFormat('sv-SE', {
      timeZone: getCompanyTimezone(),
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      hour12: false
    })
    return formatter.format(d).replace(' ', 'T')
  } catch(e) {
    const pad = (n: number) => n.toString().padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
  }
}

export function toCompanyDatetimeLocal(dateStr: string | null | undefined): string {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  try {
    const formatter = new Intl.DateTimeFormat('sv-SE', {
      timeZone: getCompanyTimezone(),
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      hour12: false
    })
    return formatter.format(d).replace(' ', 'T')
  } catch(e) {
    const pad = (n: number) => n.toString().padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
  }
}

export function toBackendDate(datetimeLocalStr: string | null | undefined): string {
  if (!datetimeLocalStr) return ''
  if (datetimeLocalStr.includes('Z') || datetimeLocalStr.match(/[+-]\d{2}:\d{2}$/)) {
    return datetimeLocalStr 
  }
  
  try {
    const d = new Date(datetimeLocalStr)
    const formatter = new Intl.DateTimeFormat('en-US', { 
      timeZone: getCompanyTimezone(), 
      timeZoneName: 'longOffset' 
    })
    const formatted = formatter.format(d)
    const match = formatted.match(/GMT([+-]\d{2}:\d{2})/)
    let offset = 'Z'
    if (match) {
      offset = match[1]
    }
    
    let finalStr = datetimeLocalStr
    if (finalStr.length === 16) {
      finalStr += ':00'
    }
    return finalStr + offset
  } catch(e) {
    let finalStr = datetimeLocalStr
    if (finalStr.length === 16) finalStr += ':00'
    return finalStr + 'Z'
  }
}
