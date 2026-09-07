<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useProjectStore } from '../stores/projectStore'
import {
  SelectImageFile, ImportAsset, GetImageBase64, DeleteAsset,
  SelectAudioFile, ImportAudioAsset, ListAudioAssets, DeleteAudioAsset,
  SetProjectPath
} from '../../wailsjs/go/main/App'

const store = useProjectStore()
const assetCache = ref<Record<string,string>>({})
const sectionsOpen = ref({ bg: true, re: true, av: true, sfx: true, vo: true, mu: true })

const backgroundAssets = computed(() => store.backgroundAssets?? [])
const reactionAssets = computed(() => store.reactionAssets?? [])
const avatarAssets = computed(() => store.avatarAssets?? [])
const avatarRulesCount = computed(() => store.avatarSystem?.rules?.length?? 0)
const avatarCount = computed(() => avatarAssets.value.length?? 0)
const totalFiles = computed(() =>
  (backgroundAssets.value.length + reactionAssets.value.length + avatarAssets.value.length + sfxAssets.value.length + voiceAssets.value.length + musicAssets.value.length)
)

const sfxAssets = ref<string[]>([])
const voiceAssets = ref<string[]>([])
const musicAssets = ref<string[]>([])

function norm(p: string) {
  return (p || '').replace(/\\/g, '/').trim()
}

async function refreshAssets() {
  if (!store.projectPath) return
  try {
    await SetProjectPath(store.projectPath)
    await store.loadAssets()
    for (const raw of (store.assets.all?? [])) {
      const key = norm(raw)
      if (!key) continue
      if (assetCache.value[key] || assetCache.value[raw]) continue
      try {
        const b64 = await GetImageBase64(key)
        assetCache.value[key] = b64
        assetCache.value[raw] = b64
        assetCache.value[raw.trim()] = b64
      } catch (e) {
        console.warn('[ASSETS] thumb fail', key, e)
      }
    }
    await refreshAudio()
  } catch(e) {
    console.warn('[ASSETS] refresh error', e)
  }
}

async function refreshAudio() {
  if (!store.projectPath) return
  try {
    await SetProjectPath(store.projectPath)
    const allAudio = (await ListAudioAssets() as any)?? [] as string[]
    const lower = (s:string) => s.toLowerCase()
    const normalized = allAudio.map((a:string)=>norm(a))
    sfxAssets.value = normalized.filter((a:string) => a.includes('sfx/') || lower(a).endsWith('.mp3') || lower(a).endsWith('.wav') || lower(a).endsWith('.ogg'))
    voiceAssets.value = normalized.filter((a:string) => a.includes('voice/'))
    musicAssets.value = normalized.filter((a:string) => a.includes('music/'))
  } catch (e) {
    console.warn('[AUDIO] list error', e)
    sfxAssets.value = []
    voiceAssets.value = []
    musicAssets.value = []
  }
}

const getAssetUrl = (p: string) => {
  if (!p) return ''
  const n = norm(p)
  return assetCache.value[p] || assetCache.value[n] || assetCache.value[p.trim()] || ''
}

async function importAsset(type: 'bg' | 're') {
  const filePath = await SelectImageFile()
  if (!filePath ||!store.projectPath) return
  await SetProjectPath(store.projectPath)
  let newRel = ""
  try { newRel = await ImportAsset(filePath, type) as string } catch(e) { return }
  const n = norm(newRel)
  for (let i = 0; i < 6; i++) {
    try {
      const b64 = await GetImageBase64(n)
      if (b64) { assetCache.value[n]=b64; assetCache.value[newRel]=b64; break }
    } catch { await new Promise(r=>setTimeout(r,150)) }
  }
  await new Promise(r => setTimeout(r, 200))
  await refreshAssets()
}

async function importAudio(type: 'sfx' | 'voice' | 'music') {
  const filePath = await SelectAudioFile()
  if (!filePath ||!store.projectPath) return
  await SetProjectPath(store.projectPath)
  await ImportAudioAsset(filePath, type)
  await new Promise(r => setTimeout(r, 150))
  await refreshAudio()
}

async function deleteAsset(asset: string) {
  if (!confirm(`Usunąć ${norm(asset).replace('images/', '')}?`)) return
  await SetProjectPath(store.projectPath!)
  await DeleteAsset(norm(asset))
  const n = norm(asset)
  delete assetCache.value[asset]
  delete assetCache.value[n]
  delete assetCache.value[n.trim()]
  await refreshAssets()
}

async function deleteAudioAsset(asset: string) {
  if (!confirm(`Usunąć ${norm(asset)}?`)) return
  await SetProjectPath(store.projectPath!)
  await DeleteAudioAsset(norm(asset))
  await refreshAudio()
}

function toggleSection(t: 'bg'|'re'|'av'|'sfx'|'vo'|'mu') { (sectionsOpen.value as any)[t] =!(sectionsOpen.value as any)[t] }
function selectDay(day: string) {
  store.currentDay = day
  store.currentSceneId = store.days[day]?.[0]?.Id || null
}
function handleAddDay() {
  const base = `day${(store.dayFileList?.length?? 0) + 1}`
  const name = prompt('Nazwa nowego dnia? (np. day2)', base)
  if (!name) return
  const id = name.trim().toLowerCase().replace(/\s+/g,'_')
  if (!id) return
  store.addDay(id)
  store.saveProject()
}
async function handleDeleteDay(e: Event, day: string) {
  e.stopPropagation()
  if ((store.dayFileList?.length?? 0) <= 1) { alert('Musisz zostawić minimum 1 dzień!'); return }
  if (!confirm(`Na pewno usunąć dzień "${day}"?`)) return
  await store.deleteDay(day)
  await store.saveProject()
}

onMounted(async () => {
  if (store.projectPath) {
    await SetProjectPath(store.projectPath)
  }
  await refreshAssets()
})

watch(() => store.projectPath, async (newPath) => {
  if (newPath) {
    await SetProjectPath(newPath)
  }
  await refreshAssets()
})

// po zamknięciu edytora Janusza odśwież listę avatarów
watch(() => store.ui.showAvatarEditor, async (open) => {
  if (!open) {
    await refreshAssets()
  }
})
</script>

<template>
  <aside v-if="store.meta" class="side-panel">
    <div class="side-scroll">
      <div class="panel-section highlight">
        <div class="label">PROJEKT JANUSZA</div>
        <div class="project-name" :title="store.meta?.gameName?? ''">{{ store.meta?.gameName?? '...' }}</div>
        <button @click="store.ui.showAvatarEditor = true" class="btn-janusz">
          <span class="icon">🧠</span>
          <span class="text">
            <b>KONFIGURUJ JANUSZA</b>
            <small>{{ avatarRulesCount }} rules • {{ avatarCount }} avatars</small>
          </span>
        </button>
      </div>

      <div class="panel-section">
        <div class="section-header"><h4>ASSETY</h4></div>
          <div class="asset-buttons">
            <button @click="importAsset('bg')" class="btn-asset">+ Tło</button>
            <button @click="importAsset('re')" class="btn-asset">+ Reakcja</button>
            <button @click="handleAddDay" class="btn-asset">+ Dzień</button>
            <button @click="importAudio('sfx')" class="btn-asset">+ Dźwięk</button>
            <button @click="importAudio('voice')" class="btn-asset">+ Narrator</button>
            <button @click="importAudio('music')" class="btn-asset">+ Muzyka</button>
          </div>
        <div class="asset-count">Łącznie: {{ totalFiles }} plików</div>
      </div>

      <div class="panel-section days-section">
        <div class="section-header"><h4>DNI [{{ store.dayFileList?.length?? 0 }}]</h4></div>
        <div v-for="day in (store.dayFileList?? [])" :key="day" :class="['day-item', { active: day === store.currentDay }]" @click="selectDay(day)">
          <span>📁 {{ day }}</span>
          <div class="day-right"><span class="badge">{{ store.days[day]?.length?? 0 }}</span><button @click="handleDeleteDay($event, day)" class="btn-del">✕</button></div>
        </div>
      </div>

      <div class="panel-section assets-section">
        <div class="asset-category">
          <div class="category-header" @click="toggleSection('bg')"><span class="arrow">{{ sectionsOpen.bg? '▼':'▶' }}</span><span>TŁA</span><span class="count">[{{ backgroundAssets.length }}]</span></div>
          <div v-if="sectionsOpen.bg" class="asset-list">
            <div v-for="a in backgroundAssets" :key="a" class="asset-item"><img v-if="getAssetUrl(a)" :src="getAssetUrl(a)" /><div v-else class="thumb-placeholder"></div><span class="name">{{ norm(a).replace('images/bg_','').replace('images/','') }}</span><button @click.stop="deleteAsset(a)" class="btn-del">✕</button></div>
            <div v-if="!backgroundAssets.length" class="empty">Brak teł</div>
          </div>
        </div>

        <div class="asset-category">
          <div class="category-header" @click="toggleSection('re')"><span class="arrow">{{ sectionsOpen.re? '▼':'▶' }}</span><span>REAKCJE</span><span class="count">[{{ reactionAssets.length }}]</span></div>
          <div v-if="sectionsOpen.re" class="asset-list">
            <div v-for="a in reactionAssets" :key="a" class="asset-item"><img v-if="getAssetUrl(a)" :src="getAssetUrl(a)" /><div v-else class="thumb-placeholder"></div><span class="name">{{ norm(a).replace('images/re_','').replace('images/','') }}</span><button @click.stop="deleteAsset(a)" class="btn-del">✕</button></div>
            <div v-if="!reactionAssets.length" class="empty">Brak reakcji</div>
          </div>
        </div>

        <div class="asset-category">
          <div class="category-header" @click="toggleSection('av')"><span class="arrow">{{ sectionsOpen.av? '▼':'▶' }}</span><span>AVATARY</span><span class="count">[{{ avatarAssets.length }}]</span></div>
          <div v-if="sectionsOpen.av" class="asset-list av-list">
            <div v-for="a in avatarAssets" :key="a" class="asset-item"><img v-if="getAssetUrl(a)" :src="getAssetUrl(a)" /><div v-else class="thumb-placeholder"></div><span class="name">{{ norm(a).replace('images/av_','').replace('images/','') }}</span><button @click.stop="deleteAsset(a)" class="btn-del">✕</button></div>
            <div v-if="!avatarAssets.length" class="empty">Brak avatarów</div>
          </div>
        </div>

        <div class="asset-category">
          <div class="category-header" @click="toggleSection('sfx')"><span class="arrow">{{ sectionsOpen.sfx? '▼':'▶' }}</span><span>DŹWIĘKI</span><span class="count">[{{ sfxAssets.length }}]</span></div>
          <div v-if="sectionsOpen.sfx" class="asset-list">
            <div v-for="a in sfxAssets" :key="a" class="asset-item audio-item">
              <span class="audio-icon">🔊</span>
              <span class="name">{{ norm(a).replace('sounds/sfx/','').replace('sounds/','') }}</span>
              <button @click.stop="deleteAudioAsset(a)" class="btn-del">✕</button>
            </div>
            <div v-if="!sfxAssets.length" class="empty">Brak dźwięków</div>
          </div>
        </div>

        <div class="asset-category">
          <div class="category-header" @click="toggleSection('vo')"><span class="arrow">{{ sectionsOpen.vo? '▼':'▶' }}</span><span>NARRATOR</span><span class="count">[{{ voiceAssets.length }}]</span></div>
          <div v-if="sectionsOpen.vo" class="asset-list">
            <div v-for="a in voiceAssets" :key="a" class="asset-item audio-item">
              <span class="audio-icon">🎙</span>
              <span class="name">{{ norm(a).replace('sounds/voice/','').replace('sounds/','') }}</span>
              <button @click.stop="deleteAudioAsset(a)" class="btn-del">✕</button>
            </div>
            <div v-if="!voiceAssets.length" class="empty">Brak nagrań</div>
          </div>
        </div>

        <div class="asset-category">
          <div class="category-header" @click="toggleSection('mu')"><span class="arrow">{{ sectionsOpen.mu? '▼':'▶' }}</span><span>MUZYKA</span><span class="count">[{{ musicAssets.length }}]</span></div>
          <div v-if="sectionsOpen.mu" class="asset-list">
            <div v-for="a in musicAssets" :key="a" class="asset-item audio-item">
              <span class="audio-icon">🎵</span>
              <span class="name">{{ norm(a).replace('sounds/music/','').replace('sounds/','') }}</span>
              <button @click.stop="deleteAudioAsset(a)" class="btn-del">✕</button>
            </div>
            <div v-if="!musicAssets.length" class="empty">Brak muzyki</div>
          </div>
        </div>

      </div>
    </div>
  </aside>
  <aside v-else class="side-panel">
    <div class="side-scroll"><div class="empty">Ładowanie projektu...</div></div>
  </aside>
</template>

<style scoped>
.side-panel{width:280px;min-width:280px;height:100%;background:#0D1117;border-right:1px solid #21262D;display:flex;flex-direction:column;overflow:hidden;box-sizing:border-box}
.side-scroll{flex:1;overflow-y:auto;overflow-x:hidden;display:flex;flex-direction:column;gap:16px;padding:12px;box-sizing:border-box}
.side-scroll::-webkit-scrollbar{width:6px}
.side-scroll::-webkit-scrollbar-thumb{background:#30363D;border-radius:3px}
.panel-section{flex-shrink:0;display:flex;flex-direction:column;gap:8px}
.panel-section.highlight{background:rgba(0,255,148,0.06);border:1px solid rgba(0,255,148,0.25);border-radius:8px;padding:10px}
.label{font-size:10px;font-weight:700;letter-spacing:1px;color:#7D8590}
.project-name{font-size:12px;font-family:monospace;color:#E6EDF3;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.btn-janusz{margin-top:6px;width:100%;display:flex;gap:10px;align-items:center;background:#161B22;border:1px solid #00FF94;border-radius:6px;padding:10px 12px;color:#E6EDF3;cursor:pointer;text-align:left;box-sizing:border-box}
.btn-janusz.icon{font-size:18px;flex:0 0 18px;line-height:1}
.btn-janusz.text{display:flex;flex-direction:column;gap:2px;flex:1;min-width:0;line-height:1.15}
.btn-janusz.text b{display:block;font-size:11px;letter-spacing:.4px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.btn-janusz.text small{display:block;font-size:10px;color:#7D8590;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;font-weight:400}
.btn-janusz:hover{background:#1a2e25}
.days-section{padding-bottom:8px;border-bottom:1px solid #21262D}
.section-header h4{margin:0;font-size:11px;color:#00FF94;letter-spacing:1px}
.day-item{display:flex;justify-content:space-between;align-items:center;padding:6px 8px;background:#161B22;border-radius:4px;font-size:12px;cursor:pointer;border-left:2px solid transparent;gap:8px;flex-shrink:0}
.day-item:hover{border-left-color:#00FF94}.day-item.active{background:#1a2e25;border-left-color:#00FF94}
.day-right{display:flex;align-items:center;gap:6px}
.badge{background:#000;padding:1px 6px;border-radius:10px;font-size:10px;color:#7D8590}
.asset-buttons{display:grid;grid-template-columns:1fr 1fr 1fr;gap:6px}
.btn-asset{height:34px;background:#21262D;border:1px solid #30363D;color:#cbd5e1;border-radius:4px;font-size:11px;font-weight:700;cursor:pointer;display:flex;align-items:center;justify-content:center}
.btn-asset:hover{border-color:#8B949E;color:white;background:#30363D}
.asset-count{font-size:10px;color:#484F58;text-align:center}
.asset-category{margin-top:4px;flex-shrink:0}
.category-header{display:flex;gap:6px;align-items:center;padding:6px 8px;background:#161B22;border:1px solid #21262D;border-radius:4px;cursor:pointer;font-size:11px;font-weight:700;color:#E6EDF3}
.count{margin-left:auto;color:#7D8590;font-weight:400}
.asset-list{display:flex;flex-direction:column;gap:4px;margin-top:6px;max-height:160px;overflow-y:auto;flex-shrink:0}
.asset-list.av-list{max-height:200px}
.asset-item{display:flex;align-items:center;gap:6px;background:#161B22;border-radius:4px;padding:4px;flex-shrink:0}
.asset-item img{width:28px;height:28px;object-fit:cover;border-radius:3px;flex-shrink:0}
.thumb-placeholder{width:28px;height:28px;background:#21262D;border-radius:3px;flex-shrink:0}
.audio-item.audio-icon{width:28px;height:28px;display:flex;align-items:center;justify-content:center;background:#21262D;border-radius:3px;font-size:12px;flex-shrink:0}
.name{flex:1;font-size:10px;font-family:monospace;color:#94a3b8;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.btn-del{width:18px;height:18px;flex:0 0 18px;background:#dc2626;color:white;border:0;border-radius:3px;cursor:pointer;font-size:10px;display:flex;align-items:center;justify-content:center;line-height:1}
.btn-del:hover{background:#ef4444}
.empty{font-size:11px;color:#484F58;font-style:italic;padding:8px;text-align:center}
</style>