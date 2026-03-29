import { useState } from 'react'
import { GripVertical, AlertTriangle } from 'lucide-react'
import type { FileRecord } from '@/types'

interface ImageCardProps {
  file: FileRecord
  jobId: string
  selected: boolean
  onSelect: (id: number, multi: boolean) => void
  onClick: () => void
  dragHandleProps?: Record<string, unknown>
}

export function ImageCard({ file, jobId, selected, onSelect, onClick, dragHandleProps }: ImageCardProps) {
  const [imgError, setImgError] = useState(false)
  const thumbUrl = `/api/thumbnail?path=${encodeURIComponent(file.original_path)}&job_id=${jobId}`

  return (
    <div
      className={`group relative bg-industrial-900 border overflow-hidden cursor-pointer transition-all ${
        selected
          ? 'border-accent ring-1 ring-accent/30'
          : 'border-industrial-800 hover:border-industrial-600'
      }`}
      onClick={onClick}
    >
      {/* Selection checkbox */}
      <div
        className="absolute top-1.5 left-1.5 z-10"
        onClick={(e) => {
          e.stopPropagation()
          onSelect(file.id, e.ctrlKey || e.metaKey)
        }}
      >
        <div
          className={`w-5 h-5 border flex items-center justify-center text-[10px] font-bold transition-all ${
            selected
              ? 'bg-accent border-accent text-white'
              : 'bg-industrial-950/80 border-industrial-600 text-transparent group-hover:text-industrial-500'
          }`}
        >
          {selected ? '✓' : '·'}
        </div>
      </div>

      {/* Drag handle */}
      {dragHandleProps && (
        <div
          className="absolute top-1.5 right-1.5 z-10 opacity-0 group-hover:opacity-100 transition-opacity cursor-grab active:cursor-grabbing"
          {...dragHandleProps}
          onClick={(e) => e.stopPropagation()}
        >
          <GripVertical size={14} className="text-industrial-400" />
        </div>
      )}

      {/* Thumbnail */}
      <div className="aspect-square bg-industrial-800 flex items-center justify-center overflow-hidden">
        {imgError || file.status === 'error' ? (
          <AlertTriangle size={24} className="text-yellow-500/50" />
        ) : (
          <img
            src={thumbUrl}
            alt={file.filename}
            loading="lazy"
            className="w-full h-full object-cover"
            onError={() => setImgError(true)}
          />
        )}
      </div>

      {/* Filename */}
      <div className="px-2 py-1.5">
        <div className="text-[10px] font-mono text-industrial-400 truncate" title={file.filename}>
          {file.filename}
        </div>
      </div>
    </div>
  )
}
