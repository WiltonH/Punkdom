import { afterEach, describe, expect, it } from 'vitest'
import { applyFontSettings, fontSettingsFromEffective } from './font-variables'

describe('font variables', () => {
  afterEach(() => {
    document.documentElement.removeAttribute('style')
  })

  it('applies effective font settings to root css variables', () => {
    applyFontSettings(fontSettingsFromEffective({
      ui_font_family: 'system-sans',
      ui_font_size: 13,
      reading_font_family: 'lxgw-wenkai',
      reading_font_size: 21,
    }))

    const style = document.documentElement.style
    expect(style.getPropertyValue('--punkdom-ui-font-size')).toBe('13px')
    expect(style.getPropertyValue('--punkdom-ui-sm-font-size')).toBe('15px')
    expect(style.getPropertyValue('--punkdom-reading-font-size')).toBe('21px')
    expect(style.getPropertyValue('--punkdom-content-font-size')).toBe('21px')
    expect(style.getPropertyValue('--punkdom-reading-font-family')).toContain('LXGW WenKai')
  })

  it('does not overwrite a local content font size override', () => {
    document.documentElement.style.setProperty('--punkdom-content-font-size', '24px')

    applyFontSettings(fontSettingsFromEffective({
      reading_font_size: 18,
    }))

    expect(document.documentElement.style.getPropertyValue('--punkdom-reading-font-size')).toBe('18px')
    expect(document.documentElement.style.getPropertyValue('--punkdom-content-font-size')).toBe('24px')
  })

  it('clamps out-of-range sizes before writing variables', () => {
    applyFontSettings({
      uiFontSize: 99,
      readingFontSize: 2,
    })

    const style = document.documentElement.style
    expect(style.getPropertyValue('--punkdom-ui-font-size')).toBe('16px')
    expect(style.getPropertyValue('--punkdom-reading-font-size')).toBe('14px')
  })
})
