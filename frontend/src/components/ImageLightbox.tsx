import { useEffect, useCallback } from 'react'
import { X, ChevronLeft, ChevronRight, Tag } from 'lucide-react'
import type { FileRecord } from '@/types'

interface ImageLightboxProps {
  file: FileRecord
  jobId: string
  files: FileRecord[]
  categories: string[]
  onClose: () => void
  onNavigate: (fileId: number) => void
  onChangeCategory: (fileId: number, category: string) => void
}

export function ImageLightbox({
  file,
  jobId,
  files,
  categories,
  onClose,
  onNavigate,
  onChangeCategory,
}: ImageLightboxProps) {
  const currentIndex = files.findIndex((f) => f.id === file.id)
  const hasPrev = currentIndex > 0
  const hasNext = currentIndex < files.length - 1

  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
      if (e.key === 'ArrowLeft' && hasPrev) onNavigate(files[currentIndex - 1].id)
      if (e.key === 'ArrowRight' && hasNext) onNavigate(files[currentIndex + 1].id)
    },
    [onClose, hasPrev, hasNext, onNavigate, files, currentIndex]
  )

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])

  const imageUrl = `/api/image?path=${encodeURIComponent(file.original_path)}&job_id=${jobId}`

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/90" onClick={onClose}>
      <div
        className="relative max-w-[90vw] max-h-[90vh] flex flex-col"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-2 bg-industrial-900 border border-industrial-800">
          <div className="flex items-center gap-3">
            <span className="text-xs font-mono text-industrial-400 truncate max-w-md">
              {file.filename}
            </span>
            <span className="text-[10px] font-mono text-accent px-2 py-0.5 bg-accent/10 border border-accent/20">
              {file.final_category}
            </span>
          </div>
          <button onClick={onClose} className="text-industrial-500 hover:text-industrial-300 transition-colors">
            <X size={18} />
          </button>
        </div>

        {/* Image */}
        <div className="relative flex-1 flex items-center justify-center bg-black min-h-0">
          <img
            src={imageUrl}
            alt={file.filename}
            className="max-w-full max-h-[70vh] object-contain"
          />

          {/* Navigation arrows */}
          {hasPrev && (
            <button
              onClick={() => onNavigate(files[currentIndex - 1].id)}
              className="absolute left-2 top-1/2 -translate-y-1/2 p-2 bg-industrial-900/80 border border-industrial-700 text-industrial-300 hover:text-white transition-colors"
            >
              <ChevronLeft size={20} />
            </button>
          )}
          {hasNext && (
            <button
              onClick={() => onNavigate(files[currentIndex + 1].id)}
              className="absolute right-2 top-1/2 -translate-y-1/2 p-2 bg-industrial-900/80 border border-industrial-700 text-industrial-300 hover:text-white transition-colors"
            >
              <ChevronRight size={20} />
            </button>
          )}
        </div>

        {/* Category reassignment */}
        <div className="flex items-center gap-2 px-4 py-2 bg-industrial-900 border border-industrial-800 border-t-0">
          <Tag size={12} className="text-industrial-500" />
          <span className="text-[10px] font-mono text-industrial-500 uppercase">Reassign:</span>
          <div className="flex gap-1 flex-wrap">
            {categories.map((cat) => (
              <button
                key={cat}
                onClick={() => onChangeCategory(file.id, cat)}
                className={`text-[10px] font-mono px-2 py-0.5 border transition-colors ${
                  cat === file.final_category
                    ? 'bg-accent/20 border-accent/40 text-accent'
                    : 'bg-industrial-800 border-industrial-700 text-industrial-400 hover:text-industrial-200 hover:border-industrial-500'
                }`}
              >
                {cat}
              </button>
            ))}
          </div>
        </div>

        {/* Counter */}
        <div className="text-center py-1 text-[10px] font-mono text-industrial-600">
          {currentIndex + 1} / {files.length}
        </div>
      </div>
    </div>
  )
}
