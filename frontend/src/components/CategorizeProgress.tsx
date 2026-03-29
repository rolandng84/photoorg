import { useState } from 'react'
import { useJobStore } from '@/stores/useJobStore'
import { jobsApi } from '@/api/client'
import { XCircle, CheckCircle2, Loader2, AlertTriangle, FolderOutput, ImageOff, Zap, FolderOpen } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { AnimatePresence, motion } from 'framer-motion'

export function CategorizeProgress() {
  const navigate = useNavigate()
  const { activeJobId, jobStatus, instantMove, progress, commitProgress, categoryCounts, lastProcessedFile, recentFiles, reset, setJobStatus } = useJobStore()

  if (!activeJobId || jobStatus === 'idle') return null

  const pct = progress.total > 0 ? Math.round((progress.categorized / progress.total) * 100) : 0
  const isComplete = jobStatus === 'reviewing'
  const isCommitting = jobStatus === 'committing'
  const isCategorizing = jobStatus === 'categorizing'

  const handleCancel = async () => {
    if (!activeJobId) return
    try {
      await jobsApi.cancel(activeJobId)
      // Poll until terminal state (SSE may be down)
      let attempts = 0
      const poll = setInterval(async () => {
        attempts++
        try {
          const job = await jobsApi.get(activeJobId)
          if (['cancelled', 'committed', 'failed'].includes(job.status)) {
            clearInterval(poll)
            setJobStatus('idle')
          }
        } catch { /* ignore */ }
        if (attempts >= 5) clearInterval(poll)
      }, 2000)
    } catch (err) {
      console.error('Failed to cancel job:', err)
    }
  }

  const handleReview = () => {
    navigate(`/review/${activeJobId}`)
  }

  const handleNewJob = () => {
    reset()
  }

  const headerText = isCommitting
    ? 'Committing Files...'
    : isComplete
      ? 'Categorization Complete'
      : instantMove
        ? 'Organizing...'
        : 'Categorizing...'

  const sortedCategories = Object.entries(categoryCounts).sort((a, b) => b[1] - a[1])

  return (
    <div className="terminal-panel p-6 space-y-4">
      <div className="terminal-header flex items-center gap-2">
        {instantMove && isCategorizing && <Zap size={16} className="text-accent" />}
        {headerText}
      </div>

      {/* Progress Bar */}
      <div className="space-y-2">
        <div className="flex justify-between text-xs font-mono text-industrial-400">
          <span>{progress.categorized} / {progress.total} files</span>
          <span>{pct}%</span>
        </div>
        <div className="h-2 bg-industrial-800 rounded-full overflow-hidden">
          <div
            className="h-full bg-accent transition-all duration-300 rounded-full"
            style={{ width: `${pct}%` }}
          />
        </div>
      </div>

      {/* Current file/folder info */}
      {isCategorizing && lastProcessedFile && (
        <div className="flex items-center gap-2 text-xs font-mono text-industrial-400">
          <FolderOpen size={12} className="text-industrial-500 shrink-0" />
          <span className="text-industrial-500">{lastProcessedFile.folder}/</span>
          <span className="text-industrial-300 truncate">{lastProcessedFile.filename}</span>
        </div>
      )}

      {/* Stats */}
      <div className="flex gap-6 text-sm font-mono">
        <div className="flex items-center gap-1.5 text-industrial-300">
          <CheckCircle2 size={14} className="text-green-500" />
          {progress.categorized} categorized
        </div>
        {instantMove && progress.committed > 0 && (
          <div className="flex items-center gap-1.5 text-accent">
            <FolderOutput size={14} />
            {progress.committed} moved
          </div>
        )}
        {progress.errorCount > 0 && (
          <div className="flex items-center gap-1.5 text-industrial-300">
            <AlertTriangle size={14} className="text-yellow-500" />
            {progress.errorCount} errors
          </div>
        )}
        {isCommitting && (
          <div className="flex items-center gap-1.5 text-accent">
            <FolderOutput size={14} />
            {commitProgress.committed} / {commitProgress.total} moved
          </div>
        )}
        {isCategorizing && (
          <div className="flex items-center gap-1.5 text-industrial-400">
            <Loader2 size={14} className="animate-spin" />
            Processing...
          </div>
        )}
      </div>

      {/* Category Tally (instant move only) */}
      {instantMove && sortedCategories.length > 0 && (
        <div className="space-y-1.5">
          <div className="text-[10px] font-mono text-industrial-500 uppercase tracking-wider">
            Categories
          </div>
          <div className="flex flex-wrap gap-2">
            {sortedCategories.map(([cat, count]) => (
              <span key={cat} className="px-2 py-0.5 text-xs font-mono bg-industrial-800 border border-industrial-700 rounded">
                {cat}: <span className="text-accent">{count}</span>
              </span>
            ))}
          </div>
        </div>
      )}

      {/* Recent Thumbnails */}
      {recentFiles.length > 0 && !isCommitting && (
        <div className="space-y-1.5">
          <div className="text-[10px] font-mono text-industrial-500 uppercase tracking-wider">
            Recent
          </div>
          <div className="flex gap-2 overflow-x-auto pb-1">
            <AnimatePresence initial={false}>
              {recentFiles.slice(0, 8).map((f) => (
                <RecentThumb key={f.id} file={f} jobId={activeJobId!} />
              ))}
            </AnimatePresence>
          </div>
        </div>
      )}

      {/* Completion banner */}
      {isComplete && (
        <div className="border border-accent/30 bg-accent/5 p-3 flex items-center gap-3">
          <CheckCircle2 size={18} className="text-accent shrink-0" />
          <div className="text-sm font-mono text-industrial-300">
            Ready to commit. Review the results, then commit to move files into folders.
          </div>
        </div>
      )}

      {/* Actions */}
      <div className="flex gap-3 pt-2">
        {isCommitting ? (
          <div className="flex items-center gap-2 text-sm font-mono text-accent">
            <Loader2 size={14} className="animate-spin" />
            Moving files to category folders...
          </div>
        ) : isComplete ? (
          <>
            <button onClick={handleReview} className="btn-primary px-6 py-2 flex items-center gap-2">
              <CheckCircle2 size={14} />
              Review Results
            </button>
            <button onClick={handleNewJob} className="btn-secondary px-4 py-2">
              New Job
            </button>
          </>
        ) : (
          <button onClick={handleCancel} className="btn-secondary px-4 py-2 flex items-center gap-2">
            <XCircle size={14} />
            Cancel
          </button>
        )}
      </div>
    </div>
  )
}

function RecentThumb({ file, jobId }: { file: { id: number; original_path: string; final_category: string }; jobId: string }) {
  const [error, setError] = useState(false)
  const thumbUrl = `/api/thumbnail?path=${encodeURIComponent(file.original_path)}&job_id=${jobId}`

  return (
    <motion.div
      layout
      initial={{ opacity: 0, scale: 0.8, x: 20 }}
      animate={{ opacity: 1, scale: 1, x: 0 }}
      exit={{ opacity: 0, scale: 0.8 }}
      transition={{ duration: 0.25 }}
      className="relative shrink-0 w-20 h-20 rounded overflow-hidden bg-industrial-800 border border-industrial-700"
    >
      {error ? (
        <div className="w-full h-full flex items-center justify-center">
          <ImageOff size={16} className="text-industrial-600" />
        </div>
      ) : (
        <img
          src={thumbUrl}
          alt=""
          className="w-full h-full object-cover"
          loading="lazy"
          onError={() => setError(true)}
        />
      )}
      <div className="absolute bottom-0 inset-x-0 bg-black/70 px-1 py-0.5">
        <span className="text-[9px] font-mono text-accent truncate block">{file.final_category}</span>
      </div>
    </motion.div>
  )
}
