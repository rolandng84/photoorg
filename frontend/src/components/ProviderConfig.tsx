import { useEffect } from 'react'
import { Settings, Database, Globe, ChevronDown } from 'lucide-react'
import { useSettingsStore } from '@/stores/useSettingsStore'

const providers = [
  { id: 'ollama', name: 'Ollama', icon: Database, desc: 'Local models' },
  { id: 'openai-compatible', name: 'OpenAI-Compatible', icon: Globe, desc: 'Any compatible API' },
]

export function ProviderConfig() {
  const {
    provider, model, endpoint, apiKey, concurrency, mode,
    availableModels, updateSetting, fetchModels,
  } = useSettingsStore()

  useEffect(() => {
    fetchModels()
  }, [fetchModels])

  return (
    <div className="flex flex-col h-full terminal-panel">
      <div className="terminal-header">
        <div className="flex items-center gap-2">
          <Settings size={14} />
          <span>System Config</span>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto p-5 space-y-6">
        {/* Provider Selection */}
        <section>
          <label className="text-[10px] uppercase font-bold text-industrial-500 mb-3 block font-mono tracking-wider">
            Extraction Engine
          </label>
          <div className="grid grid-cols-2 gap-2">
            {providers.map((p) => {
              const Icon = p.icon
              return (
                <button
                  key={p.id}
                  onClick={() => updateSetting('provider', p.id)}
                  className={`flex items-center gap-2 p-3 border text-xs font-mono transition-all ${
                    provider === p.id
                      ? 'border-accent bg-accent/5 text-accent'
                      : 'border-industrial-800 bg-industrial-950 text-industrial-500 hover:border-industrial-600'
                  }`}
                >
                  <Icon size={14} />
                  <div className="text-left">
                    <div>{p.name}</div>
                    <div className="text-[9px] opacity-60">{p.desc}</div>
                  </div>
                </button>
              )
            })}
          </div>
        </section>

        {/* Model Selection */}
        <section>
          <label className="text-[10px] uppercase font-bold text-industrial-500 mb-1.5 block font-mono tracking-wider">
            Target Model
          </label>
          <div className="relative">
            <select
              className="input-field appearance-none pr-8"
              value={model}
              onChange={(e) => updateSetting('model', e.target.value)}
            >
              {availableModels.length > 0 ? (
                availableModels.map((m) => (
                  <option key={m} value={m}>{m}</option>
                ))
              ) : (
                <option value={model}>{model} (Manual)</option>
              )}
            </select>
            <ChevronDown size={14} className="absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none text-industrial-500" />
          </div>
        </section>

        {/* Endpoint */}
        <section>
          <label className="text-[10px] uppercase font-bold text-industrial-500 mb-1.5 block font-mono tracking-wider">
            {provider === 'ollama' ? 'Host Endpoint' : 'API Base URL'}
          </label>
          <input
            className="input-field"
            placeholder={provider === 'ollama' ? 'http://localhost:11434' : 'https://api.openai.com/v1'}
            value={endpoint}
            onChange={(e) => updateSetting('endpoint', e.target.value)}
          />
        </section>

        {/* API Key (only for OpenAI-compatible) */}
        {provider !== 'ollama' && (
          <section>
            <label className="text-[10px] uppercase font-bold text-industrial-500 mb-1.5 block font-mono tracking-wider">
              Access Token
            </label>
            <input
              type="password"
              className="input-field"
              placeholder="sk-..."
              value={apiKey}
              onChange={(e) => updateSetting('apiKey', e.target.value)}
            />
          </section>
        )}

        {/* Concurrency */}
        <section>
          <div className="flex justify-between items-center mb-1.5">
            <label className="text-[10px] uppercase font-bold text-industrial-500 block font-mono tracking-wider">
              Processing Threads
            </label>
            <span className="text-xs font-mono text-accent">{concurrency}</span>
          </div>
          <input
            type="range"
            min="1"
            max="8"
            step="1"
            value={concurrency}
            onChange={(e) => updateSetting('concurrency', parseInt(e.target.value))}
            className="w-full h-1 bg-industrial-800 rounded-lg appearance-none cursor-pointer accent-accent"
          />
          <div className="flex justify-between mt-1 text-[9px] text-industrial-600 font-mono">
            <span>SAFE (1)</span>
            <span>TURBO (8)</span>
          </div>
        </section>

        {/* File Strategy */}
        <section>
          <label className="text-[10px] uppercase font-bold text-industrial-500 mb-3 block font-mono tracking-wider">
            File Strategy
          </label>
          <div className="flex gap-2">
            {(['move', 'copy'] as const).map((m) => (
              <button
                key={m}
                onClick={() => updateSetting('mode', m)}
                className={`flex-1 py-2 border text-[10px] uppercase font-mono tracking-wider transition-all ${
                  mode === m
                    ? 'border-accent bg-accent text-white glow-accent'
                    : 'border-industrial-800 bg-industrial-950 text-industrial-500 hover:border-industrial-600'
                }`}
              >
                {m} Files
              </button>
            ))}
          </div>
        </section>
      </div>
    </div>
  )
}
