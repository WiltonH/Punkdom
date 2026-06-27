import { fetchAPI, jsonHeaders, readErrorMessage, requestJSON } from '@/lib/api-client'
import type { LayeredSettings, Settings, UpdateCheckResult, UpdateInstallResult } from './types'

export async function fetchSettings(): Promise<LayeredSettings> {
  return requestJSON('/api/settings')
}

export async function updateUserSettings(s: Settings): Promise<LayeredSettings> {
  return requestJSON('/api/settings/user', {
    method: 'PUT',
    headers: jsonHeaders,
    body: JSON.stringify(s),
  })
}

export async function updateWorkspaceSettings(s: Settings): Promise<LayeredSettings> {
  return requestJSON('/api/settings/workspace', {
    method: 'PUT',
    headers: jsonHeaders,
    body: JSON.stringify(s),
  })
}

export async function checkForUpdate(): Promise<UpdateCheckResult> {
  return requestJSON('/api/update/check')
}

export async function installUpdate(): Promise<UpdateInstallResult> {
  return requestJSON('/api/update/install', { method: 'POST' })
}

export async function exportDataBackup(): Promise<{ blob: Blob; filename: string }> {
  const res = await fetchAPI('/api/backup/export')
  if (!res.ok) throw new Error(await readErrorMessage(res))
  const disposition = res.headers.get('content-disposition') || ''
  const filename = filenameFromDisposition(disposition) || `Punkdom-${new Date().toISOString().replace(/[:.]/g, '-')}.zip`
  return { blob: await res.blob(), filename }
}

export async function restoreDataBackup(file: File): Promise<{ workspace: string; message: string }> {
  const form = new FormData()
  form.append('file', file)
  return requestJSON('/api/backup/restore', {
    method: 'POST',
    body: form,
  })
}

function filenameFromDisposition(disposition: string) {
  const utf8 = disposition.match(/filename\*=UTF-8''([^;]+)/i)
  if (utf8?.[1]) return decodeURIComponent(utf8[1])
  const ascii = disposition.match(/filename="?([^"]+)"?/i)
  return ascii?.[1] ? ascii[1] : ''
}

export async function testModelConfig(params: {
  openai_api_key: string
  openai_base_url: string
  openai_model: string
}): Promise<{ success: boolean; message?: string; error?: string }> {
  return requestJSON('/api/settings/test-model', {
    method: 'POST',
    headers: jsonHeaders,
    body: JSON.stringify(params),
  })
}
