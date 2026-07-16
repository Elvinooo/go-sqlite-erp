export function isMobileWidth() {
  return window.innerWidth <= 767
}

export function getBreakpoint() {
  const width = window.innerWidth
  if (width <= 767) return 'mobile'
  if (width <= 1199) return 'pad'
  return 'pc'
}

export function vibrate(duration = 20) {
  if (navigator.vibrate) navigator.vibrate(duration)
}

export function getLocation() {
  return new Promise((resolve, reject) => {
    if (!navigator.geolocation) {
      reject(new Error('当前浏览器不支持定位'))
      return
    }
    navigator.geolocation.getCurrentPosition(resolve, reject, {
      enableHighAccuracy: true,
      timeout: 10000,
      maximumAge: 30000
    })
  })
}
