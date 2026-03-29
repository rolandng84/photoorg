export interface BrowseItem {
  name: string
  path: string
  is_dir: boolean
  is_drive?: boolean
  image_count: number | null
}

export interface Job {
  id: string
  input_path: string
  status: 'categorizing' | 'reviewing' | 'committing' | 'committed' | 'undone' | 'failed' | 'cancelled'
  mode: 'move' | 'copy'
  categories: string[]
  provider: string
  model: string
  endpoint: string
  concurrency: number
  custom_prompt: string
  instant_move: boolean
  total_files: number
  categorized: number
  committed: number
  error_count: number
  created_at: number
  updated_at: number
}

export interface FileRecord {
  id: number
  job_id: string
  original_path: string
  filename: string
  file_size: number
  ai_category: string
  final_category: string
  new_path: string
  status: 'pending' | 'categorized' | 'error' | 'committed' | 'undone' | 'skipped'
  error_message: string | null
  categorized_at: number | null
  committed_at: number | null
}

export interface CategorySummary {
  category: string
  count: number
}

export interface Settings {
  provider: string
  model: string
  endpoint: string
  api_key: string
  concurrency: string
  mode: string
  categories: string
  custom_prompt: string
  [key: string]: string
}
