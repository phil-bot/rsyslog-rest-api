import { computed } from 'vue'
import { language } from '@/stores/preferences'
import en from '@/i18n/en.json'
import de from '@/i18n/de.json'

const translations = { en, de }

// Locale string for toLocaleString (number formatting)
const LOCALE_MAP = { en: 'en-US', de: 'de-DE' }

export function useLocale() {
  const locale = computed(() => language.value)

  function t(key, vars) {
    const dict = translations[language.value] ?? translations.en
    let str = dict[key] ?? translations.en[key] ?? key
    if (vars) {
      Object.entries(vars).forEach(([k, v]) => {
        str = str.replace(new RegExp(`\\{${k}\\}`, 'g'), v)
      })
    }
    return str
  }

  // Format a number with the correct locale thousands separator
  function fmtNumber(n) {
    const lc = LOCALE_MAP[language.value] ?? 'en-US'
    return Number(n).toLocaleString(lc)
  }

  return { t, locale, fmtNumber }
}
