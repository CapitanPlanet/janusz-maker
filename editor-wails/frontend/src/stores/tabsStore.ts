import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useTabsStore = defineStore('tabs', () => {
  const openFiles = ref<string[]>([]) // ścieżki do .janusz
  const activeFile = ref<string | null>(null)

  function openFile(path: string) {
    if (!openFiles.value.includes(path)) {
      openFiles.value.push(path)
    }
    activeFile.value = path
  }

  function closeFile(path: string) {
    openFiles.value = openFiles.value.filter(f => f !== path)
    if (activeFile.value === path) {
    activeFile.value = openFiles.value[openFiles.value.length - 1] ?? null
    }
  }

  return { openFiles, activeFile, openFile, closeFile }
})