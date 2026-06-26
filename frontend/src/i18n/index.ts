import { createI18n } from 'vue-i18n'
import tr from './tr.json'
import en from './en.json'

const i18n = createI18n({
  legacy: false,
  locale: 'tr',
  fallbackLocale: 'en',
  messages: {
    tr,
    en,
  },
})

export default i18n
