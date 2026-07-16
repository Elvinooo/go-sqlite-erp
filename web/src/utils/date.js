const dateFields = new Set([
  'createdAt',
  'updatedAt',
  'deletedAt',
  'orderDate',
  'purchaseDate',
  'stockTime',
  'occurredAt',
  'registeredAt',
  'startDate',
  'endDate',
  'invoiceDate',
  'dueDate',
  'takenAt'
])

export function formatDateTime(value) {
  if (!value) return ''
  if (typeof value === 'string') {
    const normalized = value.replace(' ', 'T')
    const match = normalized.match(/^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})/)
    if (match) return match[1]
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return String(value)
  const pad = (number) => String(number).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

export function formatDateFields(row) {
  if (!row || typeof row !== 'object') return row
  for (const [key, value] of Object.entries(row)) {
    if (Array.isArray(value)) {
      value.forEach(formatDateFields)
    } else if (value && typeof value === 'object') {
      formatDateFields(value)
    } else if (dateFields.has(key)) {
      row[key] = formatDateTime(value)
    }
  }
  return row
}

export function formatDateRows(rows = []) {
  return rows.map((row) => formatDateFields(row))
}
