export function browserPrint(url) {
  const win = window.open(url, '_blank')
  if (win) {
    win.focus()
  }
}

export function downloadPdf(url) {
  const link = document.createElement('a')
  link.href = url
  link.download = ''
  link.click()
}
