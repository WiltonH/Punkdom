import { fontStackFor } from './font-options'
import type { Settings } from './types'

export interface FontSettingsInput {
  uiFont?: string | null
  uiFontSize?: number | null
  readingFont?: string | null
  readingFontSize?: number | null
}

export function fontSettingsFromEffective(effective?: Settings | null): FontSettingsInput {
  return {
    uiFont: effective?.ui_font_family,
    uiFontSize: effective?.ui_font_size,
    readingFont: effective?.reading_font_family,
    readingFontSize: effective?.reading_font_size,
  }
}

export function applyFontSettings(settings: FontSettingsInput) {
  if (typeof document === 'undefined') return
  const baseSize = clampFontSize(settings.uiFontSize, 11, 16, 14)
  const smSize = clampFontSize(baseSize + 2, 12, 18, 14)
  const compactSize = clampFontSize(baseSize - 1, 10, 15, 11)
  const microSize = clampFontSize(baseSize - 2, 10, 14, 10)

  document.documentElement.style.setProperty('--punkdom-ui-font-family', fontStackFor(settings.uiFont, 'apple-system'))
  document.documentElement.style.setProperty('--punkdom-reading-font-family', fontStackFor(settings.readingFont, 'source-han-serif'))
  document.documentElement.style.setProperty('--punkdom-ui-font-size', `${baseSize}px`)
  document.documentElement.style.setProperty('--punkdom-ui-line-height', `${baseSize + 6}px`)
  document.documentElement.style.setProperty('--punkdom-ui-sm-font-size', `${smSize}px`)
  document.documentElement.style.setProperty('--punkdom-ui-sm-line-height', `${smSize + 6}px`)
  document.documentElement.style.setProperty('--punkdom-ui-caption-size', `${compactSize}px`)
  document.documentElement.style.setProperty('--punkdom-ui-compact-font-size', `${compactSize}px`)
  document.documentElement.style.setProperty('--punkdom-ui-compact-line-height', `${compactSize + 5}px`)
  document.documentElement.style.setProperty('--punkdom-ui-micro-font-size', `${microSize}px`)
  document.documentElement.style.setProperty('--punkdom-ui-micro-line-height', `${microSize + 4}px`)
  const readingFontSize = clampFontSize(settings.readingFontSize, 14, 28, 18)
  document.documentElement.style.setProperty('--punkdom-reading-font-size', `${readingFontSize}px`)
  if (!document.documentElement.style.getPropertyValue('--punkdom-content-font-size')) {
    document.documentElement.style.setProperty('--punkdom-content-font-size', `${readingFontSize}px`)
  }
}

function clampFontSize(value: unknown, min: number, max: number, fallback: number) {
  const numberValue = typeof value === 'number' ? value : Number(value)
  if (!Number.isFinite(numberValue)) return fallback
  return Math.min(max, Math.max(min, Math.round(numberValue)))
}
