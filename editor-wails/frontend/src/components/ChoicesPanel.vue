<script setup lang="ts">
import { computed, onMounted, ref, watch, onBeforeUnmount } from 'vue'
import { useProjectStore } from '../stores/projectStore'
import { ListAudioAssets, GetAudioBase64 } from '../../wailsjs/go/main/App'

const store = useProjectStore()
const scene = computed(() => store.currentScene)
const sceneIds = computed(() => store.sceneIdsInCurrentDay)
const backgrounds = computed(() => store.backgroundAssets)
const reactions = computed(() => store.reactionAssets)
const statDefs = computed(() => store.statsSystem?.stats || [])

const sfxAssets = ref<string[]>([])
const audioCache = ref<Record<string, string>>({})
const playingFile = ref<string | null>(null)
const idError = ref('')
let currentAudio: HTMLAudioElement | null = null

async function refreshSfx() {
  if (!store.projectPath) return
  try {
    sfxAssets.value = await ListAudioAssets(store.projectPath, 'sfx') as any
  } catch (e) {
    console.warn('[EDITOR] sfx list error', e)
  }
}

async function ensureAudio(file: string) {
  if (!file) return ''
  if (audioCache.value[file]) return audioCache.value[file]
  if (!store.projectPath) return ''
  try {
    const b64 = await GetAudioBase64(store.projectPath, file)
    audioCache.value[file] = b64
    return b64
  } catch (e) {
    console.warn('[EDITOR] audio load error', file, e)
    return ''
  }
}

async function togglePlay(file: string) {
  if (!file) return
  if (playingFile.value === file) {
    stopPlay()
    return
  }
  stopPlay()
  const url = await ensureAudio(file)
  if (!url) return
  currentAudio = new Audio(url)
  currentAudio.volume = 0.7
  currentAudio.onended = () => { playingFile.value = null }
  playingFile.value = file
  await currentAudio.play().catch(e => console.warn('play err', e))
}

function stopPlay() {
  if (currentAudio) {
    currentAudio.pause()
    currentAudio.src = ''
    currentAudio = null
  }
  playingFile.value = null
}

function updateScene(field: string, value: any) {
  store.updateCurrentScene(field, value)
}

function ensureStatsForChoice(choice: any) {
  if (!choice.Stats) choice.Stats = {}
  statDefs.value.forEach(s => {
    if (choice.Stats[s.id] === undefined) choice.Stats[s.id] = 0
  })
}

function setStat(choice: any, statId: string, e: Event) {
  const val = Number((e.target as HTMLInputElement).value) || 0
  if (!choice.Stats) choice.Stats = {}
  choice.Stats[statId] = val
}

function renameSceneId(e: Event){
  const input = e.target as HTMLInputElement
  let newId = input.value.trim().toLowerCase().replace(/\s+/g,'_').replace(/[^a-z0-9_]/g,'').slice(0,40)
  if(!newId){
    idError.value='ID nie może być puste'
    input.value = store.currentScene?.Id || ''
    return
  }
  if(newId.length < 2){
    idError.value='Min 2 znaki'
    return
  }
  const oldId = store.currentScene?.Id
  if(!oldId || oldId === newId) { idError.value=''; return }
  
  if(store.sceneIdsInCurrentDay.includes(newId)){
    idError.value=`ID "${newId}" już istnieje w ${store.currentDay}`
    input.value = oldId
    return
  }

  // FIX ZNIKANIA
  store.currentScene!.Id = newId
  store.currentSceneId = newId
  
  let fixed = 0
  Object.values(store.days).forEach(scenes => {
    scenes.forEach(s => {
      s.Choices?.forEach(c => {
        if(c.Next === oldId){ c.Next = newId; fixed++ }
      })
    })
  })

  if(store.meta?.startScene === oldId && store.currentDay === store.meta.startDay){
    store.meta.startScene = newId
  }

  store.saveProject()
  idError.value = ''
  console.log(`[RENAME] ${oldId} -> ${newId}, fixed ${fixed} links`)
}

onMounted(() => {
  scene.value?.Choices?.forEach((c: any) => {
    if (!c.id) c.id = crypto.randomUUID()
    ensureStatsForChoice(c)
  })
  refreshSfx()
})

onBeforeUnmount(() => stopPlay())

watch(() => store.projectPath, refreshSfx)
watch(() => store.currentScene?.Id, () => {
  stopPlay()
  refreshSfx()
  scene.value?.Choices?.forEach((c: any) => ensureStatsForChoice(c))
})

watch(statDefs, () => {
  scene.value?.Choices?.forEach((c: any) => ensureStatsForChoice(c))
}, { deep: true })

function addChoice() {
  if (!scene.value) return
  if (!scene.value.Choices) scene.value.Choices = []
  const initStats: Record<string, number> = {}
  statDefs.value.forEach(s => initStats[s.id] = 0)
  scene.value.Choices.push({
    id: crypto.randomUUID(),
    Text: 'Nowy wybór',
    Next: scene.value.Id,
    Stats: initStats,
    ReactionText: '',
    ReactionImage: '',
    SoundFile: ''
  } as any)
}

function deleteChoice(index: number) {
  scene.value?.Choices?.splice(index, 1)
}
</script>

<template>
  <div v-if="scene" class="choices-panel">
    <div class="panel-content">
      <div class="id-edit-block">
        <div class="field">
          <label>ID SCENY * <span class="hint-inline">(unikalne, bez spacji)</span></label>
          <input 
            class="input-id"
            :value="scene.Id"
            @blur="renameSceneId"
            @keydown.enter="(e) => (e.target as HTMLInputElement).blur()"
            placeholder="np. kuchnia_start"
          />
          <div v-if="idError" class="field-error">{{ idError }}</div>
        </div>
      </div>

      <h3>Edycja: {{ scene.Id }}</h3>
      <div class="field">
        <label>Tytuł sceny</label>
        <input :value="scene.SceneTitle" @input="updateScene('SceneTitle', ($event.target as HTMLInputElement).value)" placeholder="Nazwa wyświetlana w grze" />
      </div>
      <div class="field">
        <label>Tło</label>
        <select :value="scene.Background" @change="updateScene('Background', ($event.target as HTMLSelectElement).value)">
          <option value="">— Brak —</option>
          <option v-for="bg in backgrounds" :key="bg" :value="bg">{{ bg.replace('images/','') }}</option>
        </select>
      </div>
      <div class="field">
        <label>Tekst sceny</label>
        <textarea :value="scene.Text" @input="updateScene('Text', ($event.target as HTMLTextAreaElement).value)" rows="5" placeholder="Co widzi gracz..."></textarea>
      </div>
      <div class="choices-header">
        <h4>WYBORY [{{ scene.Choices?.length || 0 }}]</h4>
        <button @click="addChoice" class="btn-add">+ Wybór</button>
      </div>
      <div class="choices-list">
        <div v-for="(choice, idx) in scene.Choices || []" :key="choice.id" class="choice-card">
          <div class="choice-top">
            <input v-model="choice.Text" placeholder="Tekst wyboru" class="choice-input" />
            <button @click="deleteChoice(idx)" class="btn-delete">✕</button>
          </div>
          <div class="row-2">
            <div class="inline-field">
              <span>Przejdź do:</span>
              <select v-model="choice.Next">
                <option v-for="id in sceneIds" :key="id" :value="id">{{ id }}</option>
              </select>
            </div>
            <div class="inline-field">
              <span>Reakcja:</span>
              <select v-model="choice.ReactionImage">
                <option value="">— Brak —</option>
                <option v-for="re in reactions" :key="re" :value="re">{{ re.replace('images/','') }}</option>
              </select>
            </div>
          </div>
          <div class="row-2">
            <div class="inline-field full-span sfx-field">
              <span>🔊 SFX:</span>
              <select v-model="choice.SoundFile" @change="stopPlay()">
                <option value="">— Brak —</option>
                <option v-for="sfx in sfxAssets" :key="sfx" :value="sfx">{{ sfx.replace('sounds/sfx/','') }}</option>
              </select>
              <button v-if="choice.SoundFile" @click="togglePlay(choice.SoundFile)" class="btn-play" :class="{ playing: playingFile === choice.SoundFile }">{{ playingFile === choice.SoundFile? '■' : '▶' }}</button>
              <button v-if="choice.SoundFile" @click="choice.SoundFile=''; stopPlay()" class="btn-clear">✕</button>
            </div>
          </div>
          <div class="stats-title">Statystyki ({{ statDefs.length }}) - edytuj w 🧠 JANUSZ</div>
          <div class="stats-grid" :style="{ gridTemplateColumns: `repeat(${Math.min(Math.max(statDefs.length,2),4)}, 1fr)` }">
            <div v-for="st in statDefs" :key="st.id" class="stat">
              <label :title="st.id">{{ st.name }}</label>
              <input type="number" :value="choice.Stats?.[st.id]?? 0" @input="setStat(choice, st.id, $event)" />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
  <div v-else class="empty">Wybierz scenę</div>
</template>

<style scoped>
.choices-panel{background:#0D1117;height:100%;display:flex;flex-direction:column;min-height:0;width:100%;box-sizing:border-box}
.panel-content{padding:16px;overflow-y:auto;overflow-x:hidden;flex:1;min-height:0;display:flex;flex-direction:column;gap:12px;box-sizing:border-box}
.choices-panel h3{margin:0;color:#00FF94;font-size:11px;font-family:monospace;text-transform:uppercase;letter-spacing:1px;opacity:.9}
.id-edit-block{background:#161B22;border:1px solid #00FF94;border-radius:8px;padding:10px;margin-bottom:4px}
.input-id{font-family:monospace;font-weight:700;letter-spacing:0.5px}
.hint-inline{font-weight:400;color:#484F58;text-transform:none;letter-spacing:0}
.field-error{font-size:11px;color:#F85149;margin-top:4px}
.field{display:flex;flex-direction:column;gap:6px}
.field label{font-size:10px;color:#7D8590;text-transform:uppercase;letter-spacing:.6px;font-weight:700}
.field input,.field select,.field textarea{width:100%;box-sizing:border-box;padding:8px 10px;background:#161B22;border:1px solid #21262D;color:#E6EDF3;font-size:12px;border-radius:6px;color-scheme:dark}
.field input:focus,.field select:focus,.field textarea:focus{outline:none;border-color:#00FF94}
.field textarea{resize:vertical;line-height:1.5;min-height:80px}
.field select option,.inline-field select option{background:#161B22;color:#E6EDF3}
.choices-header{width:100%;box-sizing:border-box;display:flex;flex-direction:row;justify-content:space-between;align-items:center;gap:12px;min-height:38px;padding:0;padding-top:14px;margin-top:10px;border-top:1px solid #21262D}
.choices-header h4{margin:0;color:#E6EDF3;font-size:11px;letter-spacing:1px;line-height:1;white-space:nowrap}
.btn-add{flex:0 0 auto;height:30px;padding:0 14px;background:#00FF94;border:0;color:#000;font-size:12px;font-weight:800;border-radius:6px;cursor:pointer;white-space:nowrap}
.btn-add:hover{filter:brightness(1.1)}
.choices-list{display:flex;flex-direction:column;gap:10px;width:100%;box-sizing:border-box}
.choice-card{background:#161B22;border:1px solid #21262D;border-radius:8px;padding:10px;display:flex;flex-direction:column;gap:10px;box-sizing:border-box;width:100%}
.choice-top{display:flex;gap:8px;align-items:center;width:100%}
.choice-input{flex:1;min-width:0;height:34px;padding:0 10px;background:#0D1117;border:1px solid #2A313C;color:#fff;font-size:12px;border-radius:6px;box-sizing:border-box}
.choice-input:focus{outline:none;border-color:#00FF94}
.btn-delete{width:34px;height:34px;flex:0 0 34px;background:#2A1215;border:1px solid #3A1A20;color:#FF8A9B;border-radius:6px;cursor:pointer;font-size:14px;font-weight:700}
.btn-delete:hover{background:#3A1A20;color:#fff}
.row-2{display:grid;grid-template-columns:1fr 1fr;gap:8px;width:100%}
.inline-field{display:flex;align-items:center;gap:8px;background:#0D1117;border:1px solid #21262D;border-radius:6px;padding:0 8px;height:34px;box-sizing:border-box;min-width:0}
.inline-field.full-span{grid-column:1 / -1}
.inline-field span{font-size:10px;color:#7D8590;white-space:nowrap;flex:0 0 auto}
.inline-field select{flex:1;min-width:0;background:#0D1117;border:0;color:#E6EDF3;font-size:11px;font-family:monospace;outline:none;padding:0;color-scheme:dark}
.btn-clear{width:18px;height:18px;flex:0 0 18px;background:#21262D;border:0;border-radius:3px;color:#7D8590;cursor:pointer;font-size:10px;display:flex;align-items:center;justify-content:center}
.btn-clear:hover{background:#30363D;color:#fff}
.btn-play{width:26px;height:26px;flex:0 0 26px;background:#21262D;border:1px solid #30363D;border-radius:6px;color:#00FF94;cursor:pointer;font-size:11px;display:flex;align-items:center;justify-content:center}
.btn-play:hover{background:#30363D;border-color:#00FF94}
.btn-play.playing{background:#00FF94;color:#000;border-color:#00FF94;animation:pulse 1s infinite}
@keyframes pulse{0%{box-shadow:0 0 0 0 rgba(0,255,148,.4)}70%{box-shadow:0 0 0 6px rgba(0,255,148,0)}100%{box-shadow:0 0 0 0 rgba(0,255,148,0)}}
.stats-title{font-size:9px;color:#7D8590;text-transform:uppercase;letter-spacing:1px;margin-top:2px}
.stats-grid{display:grid;gap:8px;width:100%}
.stat{display:flex;flex-direction:column;gap:4px;min-width:0}
.stat label{font-size:10px;color:#7D8590;text-align:center;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;font-weight:700}
.stat input{width:100%;box-sizing:border-box;height:32px;background:#0D1117;border:1px solid #21262D;color:#E6EDF3;font-size:12px;border-radius:6px;text-align:center}
.stat input:focus{outline:none;border-color:#00FF94}
.empty{padding:60px 20px;text-align:center;color:#484F58;font-size:12px}
@media(max-width:600px){.row-2{grid-template-columns:1fr}.stats-grid{grid-template-columns:1fr 1fr!important}}
</style>
