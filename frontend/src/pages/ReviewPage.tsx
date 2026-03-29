import { useEffect, useState, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { ReviewGrid } from '@/components/ReviewGrid'
import { jobsApi, filesApi } from '@/api/client'
import { useJobStore } from '@/stores/useJobStore'
import { Layout, CheckCircle2, Loader2, ArrowLeft } from 'lucide-react'
import type { Job, FileRecord, CategorySummary } from '@/types'

export function ReviewPage() {
  const { jobId: paramJobId } = useParams<{ jobId: string }>()
  const navigate = useNavigate()
  const activeJobId = useJobStore((s) => s.activeJobId)

  const resolvedJobId = paramJobId || activeJobId

  const [job, setJob] = useState<Job | null>(null)
  const [files, setFiles] = useState<FileRecord[]>([])
  const [summary, setSummary] = useState<CategorySummary[]>([])
  const [loading, setLoading] = useState(true)
  const [committing, setCommitting] = useState(false)
  const [showConfirm, setShowConfirm] = useState(false)

  const fetchData = useCallback(async () => {
    if (!resolvedJobId) return
    try {
      const [jobData, filesData, summaryData] = await Promise.all([
        jobsApi.get(resolvedJobId),
        filesApi.list(resolvedJobId, { per_page: 10000 }),
        filesApi.summary(resolvedJobId),
      ])
      setJob(jobData)
      setFiles(filesData.files)
      setSummary(summaryData)
    } catch (err) {
      console.error('Failed to fetch review data:', err)
    } finally {
      setLoading(false)
    }
  }, [resolvedJobId])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  const handleCommit = async () => {
    if (!resolvedJobId) return
    setShowConfirm(false)
    setCommitting(true)
    try {
      await jobsApi.commit(resolvedJobId)
      // Poll for completion
      const poll = setInterval(async () => {
        const updated = await jobsApi.get(resolvedJobId)
        if (updated.status === 'committed' || updated.status === 'failed') {
          clearInterval(poll)
          setCommitting(false)
          setJob(updated)
          fetchData()
        }
      }, 1000)
    } catch (err) {
      console.error('Commit failed:', err)
      setCommitting(false)
    }
  }

  if (!resolvedJobId) {
    return (
      <div className="flex-1 min-h-0 flex items-center justify-center">
        <div className="terminal-panel p-12 text-center">
          <Layout size={48} className="text-industrial-600 mx-auto mb-4" />
          <h2 className="text-lg font-mono font-bold text-industrial-300 mb-2">No Job Selected</h2>
          <p className="text-sm text-industrial-500 font-mono">
            Run a categorization job first, then review results here.
          </p>
          <button onClick={() => navigate('/')} className="btn-secondary mt-4 px-4 py-2 text-sm">
            Go to Setup
          </button>
        </div>
      </div>
    )
  }

  if (loading) {
    return (
      <div className="flex-1 min-h-0 flex items-center justify-center">
        <Loader2 size={32} className="text-accent animate-spin" />
      </div>
    )
  }

  if (!job) {
    return (
      <div className="flex-1 min-h-0 flex items-center justify-center">
        <div className="terminal-panel p-8 text-center">
          <p className="text-sm text-industrial-500 font-mono">Job not found</p>
        </div>
      </div>
    )
  }

  const isReviewing = job.status === 'reviewing'
  const isCommitted = job.status === 'committed'

  return (
    <div className="flex-1 min-h-0 flex flex-col gap-3">
      {/* Review header */}
      <div className="terminal-panel px-4 py-3 flex items-center justify-between shrink-0">
        <div className="flex items-center gap-4">
          <button onClick={() => navigate('/')} className="text-industrial-500 hover:text-industrial-300">
            <ArrowLeft size={16} />
          </button>
          <div>
            <div className="text-xs font-mono text-industrial-300">
              {job.input_path}
            </div>
            <div className="text-[10px] font-mono text-industrial-500">
              {job.total_files} files / {job.categorized} categorized / {job.error_count} errors
            </div>
          </div>
          <span className={`text-[10px] font-mono px-2 py-0.5 border ${
            isCommitted
              ? 'text-green-400 border-green-500/30 bg-green-500/10'
              : 'text-accent border-accent/30 bg-accent/10'
          }`}>
            {job.status.toUpperCase()}
          </span>
        </div>

        <div className="flex gap-2">
          {isReviewing && !committing && (
            <button
              onClick={() => setShowConfirm(true)}
              className="btn-primary px-6 py-2 flex items-center gap-2"
            >
              <CheckCircle2 size={14} />
              Commit Changes
            </button>
          )}
          {committing && (
            <div className="flex items-center gap-2 text-sm font-mono text-accent">
              <Loader2 size={14} className="animate-spin" />
              Committing...
            </div>
          )}
        </div>
      </div>

      {/* Grid */}
      <div className="flex-1 min-h-0">
        <ReviewGrid
          jobId={resolvedJobId}
          categories={job.categories}
          files={files}
          summary={summary}
          onRefresh={fetchData}
        />
      </div>

      {/* Commit confirmation dialog */}
      {showConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70" onClick={() => setShowConfirm(false)}>
          <div className="terminal-panel p-6 max-w-md w-full" onClick={(e) => e.stopPropagation()}>
            <h3 className="terminal-header mb-4">Confirm Commit</h3>
            <p className="text-sm text-industrial-400 font-mono mb-2">
              This will {job.mode === 'copy' ? 'copy' : 'move'}{' '}
              <span className="text-accent">{files.filter((f) => f.status === 'categorized').length}</span> files
              into category subfolders inside:
            </p>
            <p className="text-sm text-accent font-mono mb-4 break-all">{job.input_path}</p>
            <p className="text-[10px] text-industrial-500 font-mono mb-6">
              {job.mode === 'move'
                ? 'Files will be moved. You can undo this from the History page.'
                : 'Files will be copied. Originals remain in place.'}
            </p>
            <div className="flex justify-end gap-3">
              <button onClick={() => setShowConfirm(false)} className="btn-secondary px-4 py-2">
                Cancel
              </button>
              <button onClick={handleCommit} className="btn-primary px-6 py-2">
                Commit
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
