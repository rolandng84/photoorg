import { useState } from 'react'
import { useSettingsStore } from '@/stores/useSettingsStore'

export function CategoryEditor() {
  const { categories, customPrompt, updateSetting } = useSettingsStore()
  const [newCat, setNewCat] = useState('')

  const addCategory = () => {
    const val = newCat.trim().toLowerCase()
    if (val && !categories.includes(val)) {
      updateSetting('categories', [...categories, val])
      setNewCat('')
    }
  }

  const removeCategory = (cat: string) => {
    updateSetting('categories', categories.filter((c) => c !== cat))
  }

  return (
    <div className="terminal-panel flex flex-col">
      <div className="terminal-header">
        <span>Categories & Prompt</span>
      </div>

      <div className="p-5 space-y-4 flex-1 overflow-y-auto">
        {/* Categories Tag Cloud */}
        <div>
          <span className="text-[9px] text-industrial-500 block mb-1.5 font-mono uppercase tracking-wider">
            Active Categories
          </span>
          <div className="flex flex-wrap gap-1.5 mb-2">
            {categories.map((cat) => (
              <span
                key={cat}
                className="group flex items-center gap-1 px-2 py-1 bg-industrial-800 text-industrial-300 text-[10px] font-mono border border-industrial-700 hover:border-red-500/50 hover:bg-red-500/10 hover:text-red-400 transition-colors cursor-pointer"
                onClick={() => removeCategory(cat)}
                title="Click to remove"
              >
                {cat}
                <span className="opacity-0 group-hover:opacity-100 text-red-400">&times;</span>
              </span>
            ))}
          </div>
          <div className="flex gap-2">
            <input
              type="text"
              placeholder="+ Add category"
              className="input-field flex-1 text-[11px]"
              value={newCat}
              onChange={(e) => setNewCat(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') addCategory()
              }}
            />
            <button onClick={addCategory} className="btn-secondary text-[10px] px-3">
              Add
            </button>
          </div>
        </div>

        {/* Custom Prompt */}
        <div>
          <label className="text-[9px] text-industrial-500 block mb-1.5 font-mono uppercase tracking-wider">
            Vision System Prompt
          </label>
          <div className="text-[9px] text-industrial-600 mb-2 leading-tight">
            Customize how the AI categorizes images. Use{' '}
            <span className="text-accent font-bold">[categories]</span> as a placeholder.
          </div>
          <textarea
            className="input-field min-h-[80px] resize-none text-[11px] leading-relaxed border-dashed"
            placeholder="Default: Analyze this image and categorize it exactly as one of the following: [categories]. Respond with only the category name..."
            value={customPrompt}
            onChange={(e) => updateSetting('customPrompt', e.target.value)}
          />
        </div>
      </div>
    </div>
  )
}
