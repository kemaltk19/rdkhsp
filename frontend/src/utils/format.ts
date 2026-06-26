export function toTurkishUpperCase(str: string): string {
  if (!str) return ''
  return str.replace(/i/g, 'İ').replace(/ı/g, 'I').toLocaleUpperCase('tr-TR')
}

export function toTurkishTitleCase(str: string): string {
  if (!str) return ''
  return str.split(' ').map(word => {
    if (word.length === 0) return ''
    const firstChar = word.charAt(0).replace(/i/g, 'İ').replace(/ı/g, 'I').toLocaleUpperCase('tr-TR')
    const rest = word.slice(1).replace(/İ/g, 'i').replace(/I/g, 'ı').toLocaleLowerCase('tr-TR')
    return firstChar + rest
  }).join(' ')
}

export function setupGlobalInputFormatting() {
  document.addEventListener('focusout', (e) => {
    const el = e.target as HTMLInputElement | HTMLTextAreaElement
    if (!el) return
    if ((el.tagName === 'INPUT' && (el.type === 'text' || el.type === 'search')) || el.tagName === 'TEXTAREA') {
      const isEmailLike = el.type === 'email' || 
                          (el.name && el.name.toLowerCase().includes('email')) || 
                          (el.id && el.id.toLowerCase().includes('email')) || 
                          (el.placeholder && (el.placeholder.toLowerCase().includes('eposta') || el.placeholder.toLowerCase().includes('e-posta') || el.placeholder.toLowerCase().includes('mail')));
      
      if (el.classList.contains('no-format') || isEmailLike || el.type === 'password' || el.type === 'url') return
      
      if (el.value && typeof el.value === 'string') {
        const isUpper = el.classList.contains('uppercase-input') || el.dataset.uppercase === 'true'
        let newVal = ''
        if (isUpper) {
           newVal = toTurkishUpperCase(el.value)
        } else {
           newVal = toTurkishTitleCase(el.value)
        }

        if (newVal !== el.value) {
          el.value = newVal
          el.dispatchEvent(new Event('input', { bubbles: true }))
        }
      }
    }
  })
  
  document.addEventListener('input', (e) => {
    const el = e.target as HTMLInputElement | HTMLTextAreaElement
    if (!el) return
    if ((el.tagName === 'INPUT' && (el.type === 'text' || el.type === 'search')) || el.tagName === 'TEXTAREA') {
      const isUpper = el.classList.contains('uppercase-input') || el.dataset.uppercase === 'true'
      if (isUpper && el.value) {
        const newVal = toTurkishUpperCase(el.value)
        if (newVal !== el.value) {
          const start = el.selectionStart
          const end = el.selectionEnd
          el.value = newVal
          el.dispatchEvent(new Event('input', { bubbles: true }))
          if (start !== null && end !== null) {
            el.setSelectionRange(start, end)
          }
        }
      }
    }
  })
}
