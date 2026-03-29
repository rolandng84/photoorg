import { BrowserRouter, Routes, Route, NavLink } from 'react-router-dom'
import { Camera, Zap, Layout, Clock } from 'lucide-react'
import { SetupPage } from '@/pages/SetupPage'
import { ReviewPage } from '@/pages/ReviewPage'
import { HistoryPage } from '@/pages/HistoryPage'
import { useEffect } from 'react'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { useJobStore } from '@/stores/useJobStore'
import { useSSE } from '@/hooks/useSSE'
import { Toasts } from '@/components/Toasts'
import { jobsApi } from '@/api/client'

function App() {
  const loadFromServer = useSettingsStore((s) => s.loadFromServer)
  const sseConnected = useJobStore((s) => s.sseConnected)

  useEffect(() => {
    loadFromServer()
  }, [loadFromServer])

  // Hydrate job state from API after localStorage restore
  useEffect(() => {
    const { activeJobId, setJobStatus, updateProgress, reset } = useJobStore.getState()
    if (!activeJobId) return

    jobsApi.get(activeJobId).then((job) => {
      const statusMap: Record<string, 'idle' | 'categorizing' | 'reviewing' | 'committing'> = {
        categorizing: 'categorizing',
        reviewing: 'reviewing',
      }
      const frontendStatus = statusMap[job.status] || 'idle'
      setJobStatus(frontendStatus)
      updateProgress({
        categorized: job.categorized,
        total: job.total_files,
        errorCount: job.error_count,
      })

      // Terminal states: clear the active job
      if (!statusMap[job.status]) {
        reset()
      }
    }).catch(() => {
      reset()
    })
  }, [])

  useSSE()

  return (
    <BrowserRouter>
      <div className="h-screen w-screen flex flex-col p-4 bg-industrial-950 gap-4 overflow-hidden">
        {/* Top Navigation Bar */}
        <header className="flex justify-between items-center px-4 py-2 border border-industrial-800 bg-industrial-900 shrink-0">
          <div className="flex items-center gap-4">
            <div className="bg-accent p-2">
              <Camera size={18} className="text-white" />
            </div>
            <div>
              <h1 className="text-lg font-mono font-bold tracking-tighter leading-none flex items-center gap-2">
                RNIO{' '}
                <span className="text-accent underline decoration-2 underline-offset-4">PHOTOORG</span>
              </h1>
              <p className="text-[10px] text-industrial-500 font-mono uppercase tracking-widest mt-0.5">
                Vision-Driven Photo Organizer v2
              </p>
            </div>
          </div>

          <nav className="flex items-center gap-1">
            <AppNavLink to="/" icon={<Zap size={14} />} label="Setup" />
            <AppNavLink to="/review" icon={<Layout size={14} />} label="Review" />
            <AppNavLink to="/history" icon={<Clock size={14} />} label="History" />
            <div className="h-6 w-px bg-industrial-800 mx-2" />
            <div className="flex flex-col items-end">
              <span className="text-[10px] text-industrial-500 font-mono">v2.0.0</span>
              <span className="text-[10px] text-accent font-mono animate-pulse">SYSTEM_ONLINE</span>
            </div>
          </nav>
        </header>

        {/* Main Content */}
        <Routes>
          <Route path="/" element={<SetupPage />} />
          <Route path="/review" element={<ReviewPage />} />
          <Route path="/review/:jobId" element={<ReviewPage />} />
          <Route path="/history" element={<HistoryPage />} />
        </Routes>

        <Toasts />

        {/* Bottom Status Bar */}
        <footer className="h-7 shrink-0 flex items-center justify-between px-4 bg-industrial-900 border border-industrial-800 font-mono text-[10px] text-industrial-500">
          <div className="flex gap-4">
            <span>ENGINE: GO + REACT</span>
            <span>STORE: SQLITE</span>
          </div>
          <div className="flex gap-4 items-center">
            <span>PORT: 8012</span>
            <span className={sseConnected ? 'text-green-500' : 'text-industrial-600'}>
              SSE: {sseConnected ? 'CONNECTED' : 'IDLE'}
            </span>
          </div>
        </footer>
      </div>
    </BrowserRouter>
  )
}

function AppNavLink({ to, icon, label }: { to: string; icon: React.ReactNode; label: string }) {
  return (
    <NavLink
      to={to}
      end={to === '/'}
      className={({ isActive }) =>
        `flex items-center gap-2 px-3 py-1.5 text-[10px] uppercase font-bold font-mono transition-colors ${
          isActive ? 'text-accent bg-accent/5 border border-accent/20' : 'text-industrial-500 hover:text-industrial-300 border border-transparent'
        }`
      }
    >
      {icon}
      <span>{label}</span>
    </NavLink>
  )
}

export default App
