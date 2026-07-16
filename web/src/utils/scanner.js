import { Html5Qrcode } from 'html5-qrcode'

export async function startCameraScan(elementId, onResult) {
  const scanner = new Html5Qrcode(elementId)
  await scanner.start(
    { facingMode: 'environment' },
    { fps: 10, qrbox: { width: 260, height: 260 } },
    (decodedText) => {
      onResult(decodedText)
    }
  )
  return scanner
}

export function normalizeScanCode(value) {
  const text = String(value || '').trim()
  if (/^\d{15}$/.test(text)) return { type: 'imei', value: text }
  if (/^SN[:：]?/i.test(text)) return { type: 'sn', value: text.replace(/^SN[:：]?/i, '') }
  if (/^https?:\/\//i.test(text) || text.includes('{')) return { type: 'qrcode', value: text }
  return { type: 'barcode', value: text }
}
