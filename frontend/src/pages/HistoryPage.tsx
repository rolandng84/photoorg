import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { jobsApi } from '@/api/client'
import { Clock, Eye, Undo2, Trash2, Loader2, FolderOpen } from 'lucide-react'
import type { Job } from '@/types'

export function HistoryPage() {
  const navigate = useNavigate()
  const [jobs, setJobs] = useState<Job[]>([])
  const [loading, setLoading] = useState(true)

  const fetchJobs = async () => {
    try {
      const data = await jobsApi.list()
      setJobs(data)
    } catch (err) {
      console.error('Failed to fetch jobs:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchJobs()
  }, [])

  const handleUndo = async (id: string) => {
    if (!confirm('Undo this job? Files will be moved back to their original locations.')) return
    try {
      await jobsApi.undo(id)
      fetchJobs()
    } catch (err) {
      console.error('Undo failed:', err)
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('Delete this job record?')) return
    try {
      await jobsApi.delete(id)
      fetchJobs()
    } catch (err) {
      console.error('Delete failed:', err)
    }
  }

  const formatDate = (ts: number) => {
    return new Date(ts).toLocaleString(undefined, {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  const statusColor = (status: string) => {
    switch (status) {
      case 'committed':
        return 'text-green-400 bg-green-500/10 border-green-500/30'
      case 'reviewing':
        return 'text-accent bg-accent/10 border-accent/30'
      case 'categorizing':
        return 'text-yellow-400 bg-yellow-500/10 border-yellow-500/30'
      case 'undone':
        return 'text-industrial-400 bg-industrial-700/30 border-industrial-600'
      case 'failed':
      case 'cancelled':
        return 'text-red-400 bg-red-500/10 border-red-500/30'
      default:
        return 'text-industrial-400 bg-industrial-800 border-industrial-700'
    }
  }

  if (loading) {
    return (
      <div className="flex-1 min-h-0 flex items-center justify-center">
        <Loader2 size={32} className="text-accent animate-spin" />
      </div>
    )
  }

  if (jobs.length === 0) {
    return (
      <div className="flex-1 min-h-0 flex items-center justify-center">
        <div className="terminal-panel p-12 text-center">
          <Clock size={48} className="text-industrial-600 mx-auto mb-4" />
          <h2 className="text-lg font-mono font-bold text-industrial-300 mb-2">No Jobs Yet</h2>
          <p className="text-sm text-industrial-500 font-mono">
            Completed categorization jobs will appear here.
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex-1 min-h-0 flex flex-col gap-3">
      <div className="terminal-panel shrink-0">
        <div className="terminal-header px-4 py-2">Job History</div>
      </div>

      <div className="flex-1 min-h-0 overflow-y-auto space-y-2 pr-1">
        {jobs.map((job) => (
          <div key={job.id} className="terminal-panel px-4 py-3 flex items-center gap-4">
            {/* Icon */}
            <FolderOpen size={18} className="text-industrial-600 shrink-0" />

            {/* Info */}
            <div className="flex-1 min-w-0">
              <div className="text-sm font-mono text-industrial-300 truncate">
                {job.input_path}
              </div>
              <div className="flex items-center gap-3 mt-1">
                <span className="text-[10px] font-mono text-industrial-500">
                  {formatDate(job.created_at)}
                </span>
                <span className="text-[10px] font-mono text-industrial-500">
                  {job.total_files} files
                </span>
                <span className="text-[10px] font-mono text-industrial-500">
                  {job.provider}/{job.model}
                </span>
                <span className="text-[10px] font-mono text-industrial-500">
                  {job.mode}
                </span>
              </div>
            </div>

            {/* Status badge */}
            <span className={`text-[10px] font-mono px-2 py-0.5 border shrink-0 ${statusColor(job.status)}`}>
              {job.status.toUpperCase()}
            </span>

            {/* Stats */}
            <div className="text-[10px] font-mono text-industrial-500 shrink-0 w-28 text-right">
              {job.categorized}/{job.total_files} cat
              {job.committed > 0 && ` / ${job.committed} mov`}
              {job.error_count > 0 && (
                <span className="text-yellow-500"> / {job.error_count} err</span>
              )}
            </div>

            {/* Actions */}
            <div className="flex gap-1 shrink-0">
              {(job.status === 'reviewing' || job.status === 'committed') && (
                <button
                  onClick={() => navigate(`/review/${job.id}`)}
                  className="p-1.5 text-industrial-500 hover:text-accent transition-colors"
                  title="View"
                >
                  <Eye size={14} />
                </button>
              )}
              {job.status === 'committed' && (
                <button
                  onClick={() => handleUndo(job.id)}
                  className="p-1.5 text-industrial-500 hover:text-yellow-400 transition-colors"
                  title="Undo"
                >
                  <Undo2 size={14} />
                </button>
              )}
              {job.status !== 'committed' && (
                <button
                  onClick={() => handleDelete(job.id)}
                  className="p-1.5 text-industrial-500 hover:text-red-400 transition-colors"
                  title="Delete"
                >
                  <Trash2 size={14} />
                </button>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
