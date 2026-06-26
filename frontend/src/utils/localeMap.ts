// Ülke (Türkçe ad) -> para birimi / dil / saat dilimi eşlemesi.
// Kayıt formunda ülke seçilince currency + locale + timezone otomatik doldurulur.
// Desteklenen para birimleri seed ile sınırlı (TRY/USD/EUR); diğer ülkeler en yakın
// makul para birimine veya USD'ye düşer. Dil sadece tr/en (UI iki dilli).

export interface CountryMeta {
  currency: string // ISO 4217 (seed: TRY/USD/EUR)
  locale: string // 'tr' | 'en'
  timezone: string // IANA tz
}

// Sadece öne çıkan ülkeler tek tek tanımlı; gerisi için DEFAULT kullanılır.
const map: Record<string, CountryMeta> = {
  'Türkiye': { currency: 'TRY', locale: 'tr', timezone: 'Europe/Istanbul' },

  // Euro bölgesi
  'Almanya': { currency: 'EUR', locale: 'en', timezone: 'Europe/Berlin' },
  'Fransa': { currency: 'EUR', locale: 'en', timezone: 'Europe/Paris' },
  'İtalya': { currency: 'EUR', locale: 'en', timezone: 'Europe/Rome' },
  'İspanya': { currency: 'EUR', locale: 'en', timezone: 'Europe/Madrid' },
  'Hollanda': { currency: 'EUR', locale: 'en', timezone: 'Europe/Amsterdam' },
  'Belçika': { currency: 'EUR', locale: 'en', timezone: 'Europe/Brussels' },
  'Avusturya': { currency: 'EUR', locale: 'en', timezone: 'Europe/Vienna' },
  'Portekiz': { currency: 'EUR', locale: 'en', timezone: 'Europe/Lisbon' },
  'İrlanda': { currency: 'EUR', locale: 'en', timezone: 'Europe/Dublin' },
  'Yunanistan': { currency: 'EUR', locale: 'en', timezone: 'Europe/Athens' },
  'Finlandiya': { currency: 'EUR', locale: 'en', timezone: 'Europe/Helsinki' },
  'Slovakya': { currency: 'EUR', locale: 'en', timezone: 'Europe/Bratislava' },
  'Slovenya': { currency: 'EUR', locale: 'en', timezone: 'Europe/Ljubljana' },
  'Estonya': { currency: 'EUR', locale: 'en', timezone: 'Europe/Tallinn' },
  'Letonya': { currency: 'EUR', locale: 'en', timezone: 'Europe/Riga' },
  'Litvanya': { currency: 'EUR', locale: 'en', timezone: 'Europe/Vilnius' },
  'Lüksemburg': { currency: 'EUR', locale: 'en', timezone: 'Europe/Luxembourg' },
  'Malta': { currency: 'EUR', locale: 'en', timezone: 'Europe/Malta' },
  'Kıbrıs': { currency: 'EUR', locale: 'en', timezone: 'Asia/Nicosia' },

  // USD / dolar bazlı veya USD'ye düşürülenler
  'Amerika Birleşik Devletleri': { currency: 'USD', locale: 'en', timezone: 'America/New_York' },
  'Kanada': { currency: 'USD', locale: 'en', timezone: 'America/Toronto' },
  'Birleşik Krallık': { currency: 'GBP', locale: 'en', timezone: 'Europe/London' },
  'İsviçre': { currency: 'EUR', locale: 'en', timezone: 'Europe/Zurich' },
  'Birleşik Arap Emirlikleri': { currency: 'USD', locale: 'en', timezone: 'Asia/Dubai' },
  'Suudi Arabistan': { currency: 'USD', locale: 'en', timezone: 'Asia/Riyadh' },
  'Katar': { currency: 'USD', locale: 'en', timezone: 'Asia/Qatar' },
  'Kuveyt': { currency: 'USD', locale: 'en', timezone: 'Asia/Kuwait' },
  'Mısır': { currency: 'USD', locale: 'en', timezone: 'Africa/Cairo' },
  'Rusya': { currency: 'RUB', locale: 'en', timezone: 'Europe/Moscow' },
  'Çin': { currency: 'USD', locale: 'en', timezone: 'Asia/Shanghai' },
  'Japonya': { currency: 'USD', locale: 'en', timezone: 'Asia/Tokyo' },
  'Güney Kore': { currency: 'USD', locale: 'en', timezone: 'Asia/Seoul' },
  'Hindistan': { currency: 'USD', locale: 'en', timezone: 'Asia/Kolkata' },
  'Avustralya': { currency: 'USD', locale: 'en', timezone: 'Australia/Sydney' },
  'Brezilya': { currency: 'USD', locale: 'en', timezone: 'America/Sao_Paulo' },
  'Azerbaycan': { currency: 'USD', locale: 'tr', timezone: 'Asia/Baku' },
}

const DEFAULT: CountryMeta = { currency: 'USD', locale: 'en', timezone: 'Europe/Istanbul' }

export function metaForCountry(country: string | null | undefined): CountryMeta {
  if (!country) return DEFAULT
  return map[country.trim()] || DEFAULT
}

// Tarayıcıdan IANA timezone (Intl). Hata olursa Europe/Istanbul.
export function browserTimezone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || 'Europe/Istanbul'
  } catch {
    return 'Europe/Istanbul'
  }
}
