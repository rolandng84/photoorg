import { useToastStore } from '@/stores/useToastStore'
import { X, CheckCircle2, AlertTriangle, Info } from 'lucide-react'

export function Toasts() {
  const { toasts, removeToast } = useToastStore()

  if (toasts.length === 0) return null

  return (
    <div className="fixed bottom-12 right-4 z-50 flex flex-col gap-2 max-w-sm">
      {toasts.map((toast) => {
        const icon =
          toast.type === 'success' ? <CheckCircle2 size={14} className="text-green-400" /> :
          toast.type === 'error' ? <AlertTriangle size={14} className="text-red-400" /> :
          <Info size={14} className="text-accent" />

        const borderColor =
          toast.type === 'success' ? 'border-green-500/30' :
          toast.type === 'error' ? 'border-red-500/30' :
          'border-accent/30'

        return (
          <div
            key={toast.id}
            className={`terminal-panel px-3 py-2 flex items-center gap-2 ${borderColor} animate-in slide-in-from-right`}
          >
            {icon}
            <span className="text-xs font-mono text-industrial-300 flex-1">{toast.message}</span>
            <button
              onClick={() => removeToast(toast.id)}
              className="text-industrial-600 hover:text-industrial-400"
            >
              <X size={12} />
            </button>
          </div>
        )
      })}
    </div>
  )
}
