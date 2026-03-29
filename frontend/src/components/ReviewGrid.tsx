import { useState, useCallback } from 'react'
import {
  DndContext,
  DragOverlay,
  useDraggable,
  useDroppable,
  type DragStartEvent,
  type DragEndEvent,
  PointerSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core'
import { ImageCard } from './ImageCard'
import { ImageLightbox } from './ImageLightbox'
import { filesApi } from '@/api/client'
import { ChevronDown, ChevronRight, GripVertical, Layers } from 'lucide-react'
import type { FileRecord, CategorySummary } from '@/types'

interface ReviewGridProps {
  jobId: string
  categories: string[]
  files: FileRecord[]
  summary: CategorySummary[]
  onRefresh: () => void
}

export function ReviewGrid({ jobId, categories, files, summary, onRefresh }: ReviewGridProps) {
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set())
  const [lightboxFileId, setLightboxFileId] = useState<number | null>(null)
  const [collapsedCats, setCollapsedCats] = useState<Set<string>>(new Set())
  const [dragFileId, setDragFileId] = useState<number | null>(null)

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 8 } })
  )

  // Group files by category
  const grouped = new Map<string, FileRecord[]>()
  for (const cat of categories) {
    grouped.set(cat, [])
  }
  for (const f of files) {
    const cat = f.final_category || 'misc'
    if (!grouped.has(cat)) grouped.set(cat, [])
    grouped.get(cat)!.push(f)
  }

  const handleSelect = useCallback((id: number, multi: boolean) => {
    setSelectedIds((prev) => {
      const next = new Set(multi ? prev : [])
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }, [])

  const toggleCollapse = (cat: string) => {
    setCollapsedCats((prev) => {
      const next = new Set(prev)
      if (next.has(cat)) next.delete(cat)
      else next.add(cat)
      return next
    })
  }

  const handleDragStart = (event: DragStartEvent) => {
    setDragFileId(event.active.id as number)
  }

  const handleDragEnd = async (event: DragEndEvent) => {
    setDragFileId(null)
    const { active, over } = event
    if (!over) return

    const targetCategory = over.id as string
    const fileId = active.id as number
    const file = files.find((f) => f.id === fileId)
    if (!file || file.final_category === targetCategory) return

    // If dragging a selected file, bulk move all selected
    const idsToMove = selectedIds.has(fileId)
      ? Array.from(selectedIds)
      : [fileId]

    try {
      if (idsToMove.length > 1) {
        await filesApi.bulkUpdateCategory(jobId, idsToMove, targetCategory)
      } else {
        await filesApi.updateCategory(jobId, fileId, targetCategory)
      }
      setSelectedIds(new Set())
      onRefresh()
    } catch (err) {
      console.error('Failed to update category:', err)
    }
  }

  const handleBulkReassign = async (category: string) => {
    if (selectedIds.size === 0) return
    try {
      await filesApi.bulkUpdateCategory(jobId, Array.from(selectedIds), category)
      setSelectedIds(new Set())
      onRefresh()
    } catch (err) {
      console.error('Bulk reassign failed:', err)
    }
  }

  const handleLightboxCategoryChange = async (fileId: number, category: string) => {
    try {
      await filesApi.updateCategory(jobId, fileId, category)
      onRefresh()
    } catch (err) {
      console.error('Failed to update category:', err)
    }
  }

  const lightboxFile = lightboxFileId !== null ? files.find((f) => f.id === lightboxFileId) : null
  const dragFile = dragFileId !== null ? files.find((f) => f.id === dragFileId) : null

  return (
    <div className="flex flex-col h-full min-h-0">
      {/* Bulk action bar */}
      {selectedIds.size > 0 && (
        <div className="shrink-0 flex items-center gap-3 px-4 py-2 bg-accent/5 border border-accent/20 mb-3">
          <span className="text-xs font-mono text-accent">
            {selectedIds.size} selected
          </span>
          <span className="text-[10px] font-mono text-industrial-500 uppercase">Move to:</span>
          <div className="flex gap-1 flex-wrap">
            {categories.map((cat) => (
              <button
                key={cat}
                onClick={() => handleBulkReassign(cat)}
                className="text-[10px] font-mono px-2 py-0.5 bg-industrial-800 border border-industrial-700 text-industrial-400 hover:text-accent hover:border-accent/40 transition-colors"
              >
                {cat}
              </button>
            ))}
          </div>
          <button
            onClick={() => setSelectedIds(new Set())}
            className="ml-auto text-[10px] font-mono text-industrial-500 hover:text-industrial-300"
          >
            Clear
          </button>
        </div>
      )}

      {/* Grid */}
      <div className="flex-1 min-h-0 overflow-y-auto space-y-4 pr-1">
        <DndContext sensors={sensors} onDragStart={handleDragStart} onDragEnd={handleDragEnd}>
          {Array.from(grouped.entries()).map(([category, catFiles]) => {
            const count = summary.find((s) => s.category === category)?.count ?? catFiles.length
            const isCollapsed = collapsedCats.has(category)

            return (
              <CategorySection
                key={category}
                category={category}
                count={count}
                isCollapsed={isCollapsed}
                onToggle={() => toggleCollapse(category)}
              >
                {!isCollapsed && (
                  <div className="grid grid-cols-3 sm:grid-cols-4 md:grid-cols-5 lg:grid-cols-6 xl:grid-cols-8 gap-2 p-3">
                    {catFiles.map((file) => (
                      <DraggableCard
                        key={file.id}
                        file={file}
                        jobId={jobId}
                        selected={selectedIds.has(file.id)}
                        onSelect={handleSelect}
                        onClick={() => setLightboxFileId(file.id)}
                      />
                    ))}
                    {catFiles.length === 0 && (
                      <div className="col-span-full py-6 text-center text-xs font-mono text-industrial-600">
                        Drop images here
                      </div>
                    )}
                  </div>
                )}
              </CategorySection>
            )
          })}

          <DragOverlay>
            {dragFile && (
              <div className="w-24 h-24 bg-industrial-900 border border-accent shadow-lg shadow-accent/20 flex items-center justify-center">
                <Layers size={20} className="text-accent" />
                {selectedIds.size > 1 && selectedIds.has(dragFile.id) && (
                  <span className="absolute -top-2 -right-2 bg-accent text-white text-[10px] font-bold rounded-full w-5 h-5 flex items-center justify-center">
                    {selectedIds.size}
                  </span>
                )}
              </div>
            )}
          </DragOverlay>
        </DndContext>
      </div>

      {/* Lightbox */}
      {lightboxFile && (
        <ImageLightbox
          file={lightboxFile}
          jobId={jobId}
          files={files}
          categories={categories}
          onClose={() => setLightboxFileId(null)}
          onNavigate={(id) => setLightboxFileId(id)}
          onChangeCategory={handleLightboxCategoryChange}
        />
      )}
    </div>
  )
}

// Droppable category section
function CategorySection({
  category,
  count,
  isCollapsed,
  onToggle,
  children,
}: {
  category: string
  count: number
  isCollapsed: boolean
  onToggle: () => void
  children: React.ReactNode
}) {
  const { setNodeRef, isOver } = useDroppable({ id: category })

  return (
    <div
      ref={setNodeRef}
      className={`terminal-panel transition-colors ${isOver ? 'border-accent/50 bg-accent/5' : ''}`}
    >
      <button
        onClick={onToggle}
        className="w-full flex items-center gap-2 px-3 py-2 hover:bg-industrial-800/50 transition-colors"
      >
        {isCollapsed ? (
          <ChevronRight size={14} className="text-industrial-500" />
        ) : (
          <ChevronDown size={14} className="text-industrial-500" />
        )}
        <span className="terminal-header text-xs !mb-0">{category}</span>
        <span className="text-[10px] font-mono text-industrial-500 ml-1">({count})</span>
      </button>
      {children}
    </div>
  )
}

// Draggable image card wrapper
function DraggableCard({
  file,
  jobId,
  selected,
  onSelect,
  onClick,
}: {
  file: FileRecord
  jobId: string
  selected: boolean
  onSelect: (id: number, multi: boolean) => void
  onClick: () => void
}) {
  const { attributes, listeners, setNodeRef, isDragging } = useDraggable({
    id: file.id,
  })

  return (
    <div ref={setNodeRef} className={isDragging ? 'opacity-30' : ''}>
      <ImageCard
        file={file}
        jobId={jobId}
        selected={selected}
        onSelect={onSelect}
        onClick={onClick}
        dragHandleProps={{ ...attributes, ...listeners }}
      />
    </div>
  )
}
