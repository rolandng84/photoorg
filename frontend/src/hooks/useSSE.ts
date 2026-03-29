import { useEffect, useRef } from 'react'
import { useJobStore } from '@/stores/useJobStore'
import { useToastStore } from '@/stores/useToastStore'
import { jobsApi } from '@/api/client'

export function useSSE() {
  const sourceRef = useRef<EventSource | null>(null)
  const activeJobId = useJobStore((s) => s.activeJobId)
  const sseConnected = useJobStore((s) => s.sseConnected)
  const addToast = useToastStore((s) => s.addToast)

  useEffect(() => {
    const store = useJobStore.getState()

    if (!activeJobId) {
      if (sourceRef.current) {
        sourceRef.current.close()
        sourceRef.current = null
        store.setSseConnected(false)
      }
      return
    }

    const es = new EventSource('/api/events')
    sourceRef.current = es

    es.onopen = () => useJobStore.getState().setSseConnected(true)
    es.onerror = () => useJobStore.getState().setSseConnected(false)

    // Standard mode: separate categorize + progress events
    es.addEventListener('file_categorized', (e) => {
      const data = JSON.parse(e.data)
      const s = useJobStore.getState()
      if (data.job_id !== s.activeJobId) return
      s.addRecentFile({
        id: data.file_id,
        job_id: data.job_id,
        original_path: data.original_path || '',
        filename: data.filename,
        file_size: 0,
        ai_category: data.category,
        final_category: data.category,
        new_path: '',
        status: 'categorized',
        error_message: null,
        categorized_at: Date.now(),
        committed_at: null,
      })
    })

    // Instant move: single combined event with everything
    es.addEventListener('file_organized', (e) => {
      const data = JSON.parse(e.data)
      const s = useJobStore.getState()
      if (data.job_id !== s.activeJobId) return

      // Update progress inline
      s.updateProgress({
        categorized: data.categorized,
        total: data.total,
        errorCount: data.error_count,
        committed: data.committed ?? 0,
      })

      // Update category tally
      if (data.new_path) {
        s.incrementCategoryCount(data.category)
      }

      // Track current file/folder
      s.setLastProcessedFile(data.filename, data.folder || '')

      // Add recent file with new_path for thumbnail
      s.addRecentFile({
        id: data.file_id,
        job_id: data.job_id,
        original_path: data.new_path || data.original_path || '',
        filename: data.filename,
        file_size: 0,
        ai_category: data.category,
        final_category: data.category,
        new_path: data.new_path || '',
        status: data.new_path ? 'committed' : 'categorized',
        error_message: null,
        categorized_at: Date.now(),
        committed_at: data.new_path ? Date.now() : null,
      })
    })

    es.addEventListener('job_progress', (e) => {
      const data = JSON.parse(e.data)
      const s = useJobStore.getState()
      if (data.job_id !== s.activeJobId) return
      s.updateProgress({
        categorized: data.categorized,
        total: data.total,
        errorCount: data.error_count,
        committed: data.committed ?? 0,
      })
    })

    es.addEventListener('job_completed', (e) => {
      const data = JSON.parse(e.data)
      const s = useJobStore.getState()
      if (data.job_id !== s.activeJobId) return
      s.updateProgress({
        categorized: data.categorized,
        total: data.total,
        errorCount: data.errors,
        committed: data.committed ?? 0,
      })

      if (data.instant_move) {
        s.setJobStatus('idle')
        addToast(`Done! ${data.committed} files organized into folders`, 'success')
      } else {
        s.setJobStatus('reviewing')
        addToast(`Categorization complete: ${data.categorized}/${data.total} files`, 'success')
      }
    })

    es.addEventListener('job_cancelled', (e) => {
      const data = JSON.parse(e.data)
      const s = useJobStore.getState()
      if (data.job_id !== s.activeJobId) return
      s.setJobStatus('idle')
      addToast('Job cancelled', 'info')
    })

    es.addEventListener('job_failed', (e) => {
      const data = JSON.parse(e.data)
      const s = useJobStore.getState()
      if (data.job_id !== s.activeJobId) return
      s.setJobStatus('idle')
      addToast(`Job failed: ${data.error || 'unknown error'}`, 'error')
    })

    es.addEventListener('commit_progress', (e) => {
      const data = JSON.parse(e.data)
      const s = useJobStore.getState()
      if (data.job_id !== s.activeJobId) return
      s.updateCommitProgress({ committed: data.committed, total: data.total })
    })

    es.addEventListener('commit_completed', (e) => {
      const data = JSON.parse(e.data)
      const s = useJobStore.getState()
      if (data.job_id !== s.activeJobId) return
      s.setJobStatus('idle')
      addToast('Files committed successfully', 'success')
    })

    es.addEventListener('commit_failed', (e) => {
      const data = JSON.parse(e.data)
      const s = useJobStore.getState()
      if (data.job_id !== s.activeJobId) return
      s.setJobStatus('idle')
      addToast(`Commit failed: ${data.error || 'unknown error'}`, 'error')
    })

    return () => {
      es.close()
      sourceRef.current = null
      useJobStore.getState().setSseConnected(false)
    }
  }, [activeJobId, addToast])

  // Fallback: poll job status when SSE is disconnected
  useEffect(() => {
    if (!activeJobId || sseConnected) return

    const poll = setInterval(async () => {
      try {
        const job = await jobsApi.get(activeJobId)
        const s = useJobStore.getState()
        s.updateProgress({
          categorized: job.categorized,
          total: job.total_files,
          errorCount: job.error_count,
          committed: job.committed ?? 0,
        })
        // Handle terminal states
        if (['cancelled', 'committed', 'failed', 'undone'].includes(job.status)) {
          s.setJobStatus('idle')
        }
      } catch { /* ignore */ }
    }, 3000)

    return () => clearInterval(poll)
  }, [activeJobId, sseConnected])
}
