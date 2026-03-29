import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { settingsApi, modelsApi, browseApi } from '@/api/client'
import type { BrowseItem } from '@/types'

interface SettingsStore {
  // LLM Config
  provider: string
  model: string
  endpoint: string
  apiKey: string
  concurrency: number
  mode: 'move' | 'copy'
  categories: string[]
  customPrompt: string
  instantMove: boolean
  availableModels: string[]

  // Browser state
  currentPath: string
  inputPath: string
  browserItems: BrowseItem[]

  // Actions
  updateSetting: (key: string, value: string | number | boolean | string[]) => void
  setInputPath: (path: string) => void
  fetchModels: () => Promise<void>
  fetchBrowserItems: (path: string) => Promise<void>
  loadFromServer: () => Promise<void>
  saveToServer: () => Promise<void>
}

export const useSettingsStore = create<SettingsStore>()(
  persist(
    (set, get) => ({
      provider: 'ollama',
      model: 'llava:7b',
      endpoint: 'http://localhost:11434',
      apiKey: '',
      concurrency: 4,
      mode: 'move',
      categories: ['people', 'food', 'landscape', 'animals', 'documents', 'misc'],
      customPrompt: '',
      instantMove: false,
      availableModels: [],

      currentPath: '',
      inputPath: '',
      browserItems: [],

      updateSetting: (key, value) => {
        set({ [key]: value })
        // Refresh models when provider or endpoint changes
        if (key === 'provider' || key === 'endpoint') {
          get().fetchModels()
        }
      },

      setInputPath: (path) => set({ inputPath: path }),

      fetchModels: async () => {
        const { provider, endpoint, apiKey } = get()
        try {
          const models = await modelsApi.list(provider, endpoint, apiKey)
          set({ availableModels: models })
          // Auto-select first model if current not in list
          if (models.length > 0 && !models.includes(get().model)) {
            set({ model: models[0] })
          }
        } catch {
          set({ availableModels: [] })
        }
      },

      fetchBrowserItems: async (path) => {
        try {
          const items = await browseApi.list(path)
          set({ browserItems: items, currentPath: path })
        } catch {
          console.error('Failed to fetch browser items')
        }
      },

      loadFromServer: async () => {
        try {
          const settings = await settingsApi.get()
          set({
            provider: settings.provider || 'ollama',
            model: settings.model || 'llava:7b',
            endpoint: settings.endpoint || 'http://localhost:11434',
            apiKey: settings.api_key || '',
            concurrency: parseInt(settings.concurrency || '4'),
            mode: (settings.mode as 'move' | 'copy') || 'move',
            categories: JSON.parse(settings.categories || '[]'),
            customPrompt: settings.custom_prompt || '',
            instantMove: settings.instant_move === 'true',
          })
        } catch {
          console.error('Failed to load settings from server')
        }
      },

      saveToServer: async () => {
        const { provider, model, endpoint, apiKey, concurrency, mode, categories, customPrompt, instantMove } = get()
        try {
          await settingsApi.update({
            provider,
            model,
            endpoint,
            api_key: apiKey,
            concurrency: String(concurrency),
            mode,
            categories: JSON.stringify(categories),
            custom_prompt: customPrompt,
            instant_move: String(instantMove),
          })
        } catch {
          console.error('Failed to save settings to server')
        }
      },
    }),
    {
      name: 'photoorg-settings',
      partialize: (state) => ({
        provider: state.provider,
        model: state.model,
        endpoint: state.endpoint,
        apiKey: state.apiKey,
        concurrency: state.concurrency,
        mode: state.mode,
        categories: state.categories,
        customPrompt: state.customPrompt,
        instantMove: state.instantMove,
        inputPath: state.inputPath,
      }),
    }
  )
)
