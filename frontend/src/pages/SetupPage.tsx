import { FolderBrowser } from '@/components/FolderBrowser'
import { ProviderConfig } from '@/components/ProviderConfig'
import { CategoryEditor } from '@/components/CategoryEditor'
import { CategorizeProgress } from '@/components/CategorizeProgress'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { useJobStore } from '@/stores/useJobStore'
import { jobsApi } from '@/api/client'
import { useToastStore } from '@/stores/useToastStore'
import { Eye, Zap } from 'lucide-react'

export function SetupPage() {
  const { inputPath, provider, model, endpoint, apiKey, concurrency, mode, categories, customPrompt, saveToServer } = useSettingsStore()
  const { activeJobId, jobStatus, setActiveJob, setJobStatus, setInstantMove, updateProgress } = useJobStore()
  const addToast = useToastStore((s) => s.addToast)

  const isRunning = activeJobId && jobStatus !== 'idle'
  const canStart = !!inputPath && categories.length > 0

  const handleStart = async (instantMove: boolean) => {
    if (!canStart) return

    await saveToServer()

    try {
      const job = await jobsApi.create({
        input_path: inputPath,
        provider,
        model,
        endpoint,
        api_key: apiKey,
        concurrency,
        mode,
        categories,
        custom_prompt: customPrompt,
        instant_move: instantMove,
      })

      setActiveJob(job.id)
      setInstantMove(instantMove)
      setJobStatus('categorizing')
      updateProgress({ categorized: 0, total: job.total_files, errorCount: 0, committed: 0 })
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to create job'
      addToast(msg, 'error')
    }
  }

  return (
    <div className="flex-1 min-h-0 grid grid-cols-12 gap-4">
      {/* Left: Folder Browser */}
      <div className="col-span-12 lg:col-span-4 min-h-0">
        <FolderBrowser />
      </div>

      {/* Right: Config + Categories + Progress/Start */}
      <div className="col-span-12 lg:col-span-8 min-h-0 flex flex-col gap-4">
        {!isRunning && (
          <div className="flex-1 min-h-0 grid grid-cols-2 gap-4">
            <ProviderConfig />
            <CategoryEditor />
          </div>
        )}

        {isRunning ? (
          <CategorizeProgress />
        ) : (
          <div className="terminal-panel p-6 flex items-center justify-between shrink-0">
            <div>
              <div className="text-[10px] font-mono text-industrial-500 uppercase tracking-wider mb-1">
                Target Directory
              </div>
              <div className="font-mono text-sm text-accent truncate max-w-md">
                {inputPath || 'No folder selected'}
              </div>
            </div>
            <div className="flex items-center gap-3">
              <button
                onClick={() => handleStart(false)}
                disabled={!canStart}
                className="btn-secondary flex items-center gap-2 px-6 py-3 disabled:opacity-30"
              >
                <Eye size={16} />
                Review First
              </button>
              <button
                onClick={() => handleStart(true)}
                disabled={!canStart}
                className="btn-primary flex items-center gap-2 px-6 py-3 disabled:opacity-30"
              >
                <Zap size={16} fill="white" />
                Organize Now
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
