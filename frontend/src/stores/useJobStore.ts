import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { FileRecord } from '@/types'

interface JobStore {
  activeJobId: string | null
  jobStatus: 'idle' | 'categorizing' | 'reviewing' | 'committing'
  instantMove: boolean
  progress: { categorized: number; total: number; errorCount: number; committed: number }
  commitProgress: { committed: number; total: number }
  categoryCounts: Record<string, number>
  lastProcessedFile: { filename: string; folder: string } | null

  sseConnected: boolean
  recentFiles: FileRecord[]

  setActiveJob: (id: string | null) => void
  setJobStatus: (status: 'idle' | 'categorizing' | 'reviewing' | 'committing') => void
  setInstantMove: (v: boolean) => void
  updateProgress: (progress: { categorized: number; total: number; errorCount: number; committed?: number }) => void
  updateCommitProgress: (p: { committed: number; total: number }) => void
  incrementCategoryCount: (category: string) => void
  setLastProcessedFile: (filename: string, folder: string) => void
  addRecentFile: (file: FileRecord) => void
  reset: () => void
  setSseConnected: (connected: boolean) => void
}

export const useJobStore = create<JobStore>()(
  persist(
    (set) => ({
      activeJobId: null,
      jobStatus: 'idle',
      instantMove: false,
      progress: { categorized: 0, total: 0, errorCount: 0, committed: 0 },
      commitProgress: { committed: 0, total: 0 },
      categoryCounts: {},
      lastProcessedFile: null,
      sseConnected: false,
      recentFiles: [],

      setActiveJob: (id) => set({ activeJobId: id }),
      setJobStatus: (status) => set({ jobStatus: status }),
      setInstantMove: (v) => set({ instantMove: v }),
      updateProgress: (progress) =>
        set((state) => ({
          progress: { ...state.progress, ...progress },
        })),
      updateCommitProgress: (p) => set({ commitProgress: p }),
      incrementCategoryCount: (category) =>
        set((state) => ({
          categoryCounts: {
            ...state.categoryCounts,
            [category]: (state.categoryCounts[category] || 0) + 1,
          },
        })),
      setLastProcessedFile: (filename, folder) => set({ lastProcessedFile: { filename, folder } }),
      addRecentFile: (file) =>
        set((state) => ({
          recentFiles: [file, ...state.recentFiles].slice(0, 50),
        })),
      reset: () =>
        set({
          activeJobId: null,
          jobStatus: 'idle',
          instantMove: false,
          progress: { categorized: 0, total: 0, errorCount: 0, committed: 0 },
          commitProgress: { committed: 0, total: 0 },
          categoryCounts: {},
          lastProcessedFile: null,
          recentFiles: [],
        }),
      setSseConnected: (connected) => set({ sseConnected: connected }),
    }),
    {
      name: 'photoorg-job',
      partialize: (state) => ({
        activeJobId: state.activeJobId,
        jobStatus: state.jobStatus,
        instantMove: state.instantMove,
        progress: state.progress,
      }),
    }
  )
)
