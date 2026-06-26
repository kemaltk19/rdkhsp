import { ref } from 'vue'

const isDark = ref(localStorage.getItem('theme') === 'dark')

export function useTheme() {
  const toggleTheme = () => {
    isDark.value = !isDark.value
    const theme = isDark.value ? 'dark' : 'light'
    localStorage.setItem('theme', theme)
    applyTheme()
  }

  const applyTheme = () => {
    const el = document.documentElement
    if (isDark.value) {
      el.classList.add('p-dark')
    } else {
      el.classList.remove('p-dark')
    }
  }

  return {
    isDark,
    toggleTheme,
    applyTheme,
  }
}
