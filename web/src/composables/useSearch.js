import { reactive } from 'vue'

function cloneDefaults(value) {
  return JSON.parse(JSON.stringify(value))
}

export function useSearch(defaults, loader) {
  const initial = cloneDefaults(defaults)
  const searchForm = reactive(cloneDefaults(initial))

  function resetSearch(overrides = {}) {
    for (const key of Object.keys(searchForm)) {
      delete searchForm[key]
    }
    Object.assign(searchForm, cloneDefaults(initial), overrides)
  }

  function clearFilter(overrides = {}) {
    resetSearch(overrides)
  }

  async function loadData(...args) {
    if (typeof loader !== 'function') {
      return null
    }
    return loader(searchForm, ...args)
  }

  return {
    searchForm,
    resetSearch,
    clearFilter,
    loadData
  }
}
