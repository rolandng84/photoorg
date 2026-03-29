import { useEffect } from 'react'
import { Folder, ChevronRight, CornerUpLeft, Home, HardDrive as DriveIcon } from 'lucide-react'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { clsx } from 'clsx'

export function FolderBrowser() {
  const { currentPath, browserItems, fetchBrowserItems, setInputPath, inputPath } = useSettingsStore()

  useEffect(() => {
    fetchBrowserItems('')
  }, [fetchBrowserItems])

  return (
    <div className="flex flex-col h-full min-h-0 overflow-hidden terminal-panel">
      <div className="terminal-header shrink-0">
        <div className="flex items-center gap-2">
          <DriveIcon size={14} />
          <span>Local Navigator</span>
        </div>
        <div className="flex items-center gap-4">
          <button onClick={() => fetchBrowserItems('')} title="Go Home">
            <Home size={12} className="hover:text-accent transition-colors" />
          </button>
          <span className="opacity-50 text-[10px] truncate max-w-[200px]">{currentPath || 'HOME'}</span>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto min-h-0 p-2 space-y-0.5 bg-industrial-950/50">
        {browserItems.map((item, idx) => (
          <button
            key={`${item.path}-${idx}`}
            onClick={() => {
              if (item.is_dir) fetchBrowserItems(item.path)
            }}
            onDoubleClick={() => {
              if (item.is_dir && item.name !== '..') setInputPath(item.path)
            }}
            className={clsx(
              'w-full flex items-center gap-3 px-3 py-2 text-xs font-mono transition-colors group text-left shrink-0',
              'hover:bg-industrial-800',
              inputPath === item.path
                ? 'text-accent bg-accent/5 border-l-2 border-accent'
                : 'text-industrial-400'
            )}
          >
            {item.name === '..' ? (
              <CornerUpLeft size={14} className="text-industrial-500 shrink-0" />
            ) : item.is_drive ? (
              <DriveIcon size={14} className="text-industrial-500 shrink-0" />
            ) : (
              <Folder
                size={14}
                className={clsx(
                  'transition-colors shrink-0',
                  inputPath === item.path ? 'text-accent' : 'text-industrial-600 group-hover:text-industrial-300'
                )}
              />
            )}
            <span className="truncate flex-1">{item.name}</span>

            {item.image_count !== null && item.image_count > 0 && (
              <span className="shrink-0 px-1.5 py-0.5 bg-industrial-800 text-industrial-300 text-[9px] rounded-full border border-industrial-700">
                {item.image_count}
              </span>
            )}

            {item.is_dir && item.name !== '..' && (
              <ChevronRight size={12} className="opacity-0 group-hover:opacity-100 transition-opacity shrink-0 text-industrial-600" />
            )}
          </button>
        ))}
      </div>

      <div className="p-4 border-t border-industrial-800 bg-industrial-900 shrink-0">
        <div className="flex items-center justify-between mb-2">
          <span className="text-[10px] uppercase tracking-wider text-industrial-500 font-bold font-mono">Selected Target</span>
          <button
            onClick={() => setInputPath(currentPath || '')}
            className="text-[10px] text-accent hover:underline uppercase font-mono"
          >
            Use Current
          </button>
        </div>
        <div className="bg-industrial-950 p-2 border border-industrial-700 font-mono text-[11px] text-accent truncate">
          {inputPath || 'No path selected'}
        </div>
      </div>
    </div>
  )
}
