<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { SelectFolder, CreateProject, AddRecentProject, GetRecentProjects, GetDefaultProjectPath } from '../../wailsjs/go/main/App'
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime'
import { useProjectStore } from '../stores/projectStore'

import januszLogo from '../assets/janusz_logo_clean.png'
import bgImage from '../assets/bg.png'
import tutorialImage from '../assets/bg_tutorial.webp'

const store = useProjectStore()
const recentProjects = ref<string[]>([])
const showNewProjectModal = ref(false)
const showTutorialModal = ref(false)
const newProjectName = ref('')
const selectedFolder = ref('')
const isCreating = ref(false)

async function loadRecentProjects() {
  try {
    const all = await GetRecentProjects()
    recentProjects.value = all.slice(0, 5)
  } catch (e) { recentProjects.value = [] }
}
async function openProjectDialog() {
  try {
    const selected = await SelectFolder()
    if (selected) { await AddRecentProject(selected); await store.loadProjectFromPath(selected) }
  } catch (e) { alert('Błąd otwierania: ' + e) }
}
async function pickFolder() {
  try { const folder = await SelectFolder(); if (folder) selectedFolder.value = folder } catch {}
}
async function createProject() {
  if (!newProjectName.value.trim()) { alert('Wpisz nazwę projektu'); return }
  if (!selectedFolder.value) { alert('Wybierz folder docelowy'); return }
  isCreating.value = true
  try {
    const fullPath = `${selectedFolder.value}/${newProjectName.value}`
    await CreateProject(fullPath, newProjectName.value)
    await AddRecentProject(fullPath)
    await store.loadProjectFromPath(fullPath)
    showNewProjectModal.value = false; newProjectName.value = ''; selectedFolder.value = ''
    await loadRecentProjects()
  } catch (e: any) { alert('Błąd tworzenia: ' + (e.message || e.toString())) } finally { isCreating.value = false }
}
async function openNewProjectModal() { selectedFolder.value = await GetDefaultProjectPath(); newProjectName.value = ''; showNewProjectModal.value = true }
async function openRecent(path: string) {
  try { await store.loadProjectFromPath(path); await AddRecentProject(path) } catch (e) { alert('Nie można otworzyć: ' + e); await loadRecentProjects() }
}
function getProjectName(path: string) { return path.split('\\').pop() || path.split('/').pop() || path }
function donate() { BrowserOpenURL('https://tipply.pl/@kapitanplaneta') }
onMounted(loadRecentProjects)
</script>

<template>
  <div class="launcher" :style="{ backgroundImage: `linear-gradient(rgba(0,0,0,0.3), rgba(0,0,0,0.3)), url(${bgImage})` }">
    <div class="launcher-content">
      <img :src="januszLogo" alt="Janusz Logo" class="menu-header" />
      <h1>Edytor Janusza V2.0</h1>
      <div class="launcher-buttons">
        <button @click="openNewProjectModal" class="btn-primary">+ Nowy Projekt</button>
        <button @click="openProjectDialog" class="btn-secondary">📂 Otwórz Projekt</button>
        <button @click="showTutorialModal = true" class="btn-help">❓ Instrukcja Obsługi</button>
      </div>
      <div v-if="recentProjects.length" class="recent">
        <h3>OSTATNIE PROJEKTY</h3>
        <div class="recent-list">
          <div v-for="p in recentProjects" :key="p" @click="openRecent(p)" class="recent-item">
            📁 {{ getProjectName(p) }}<span class="recent-path">{{ p }}</span>
          </div>
        </div>
      </div>
      <button @click="donate" class="beer-tip">DAJ 3zł NA PIWO 🍺</button>
    </div>
  </div>

  <div v-if="showNewProjectModal" class="modal" @click.self="showNewProjectModal = false">
    <div class="modal-content">
      <h3>Nowy projekt Janusza</h3>
      <div class="form-group"><label>NAZWA PROJEKTU:</label><input v-model="newProjectName" placeholder="MojaSuperGra" @keyup.enter="createProject" @keyup.esc="showNewProjectModal = false" autofocus maxlength="50" /></div>
      <div class="form-group"><label>LOKALIZACJA:</label><div class="folder-picker"><input :value="selectedFolder" type="text" readonly placeholder="Kliknij Przeglądaj..." /><button @click="pickFolder">Przeglądaj...</button></div><small v-if="selectedFolder && newProjectName" class="hint">Utworzony zostanie: {{ selectedFolder }}/{{ newProjectName }}</small></div>
      <div class="modal-buttons"><button @click="createProject" :disabled="isCreating || !newProjectName.trim()" class="btn-primary">{{ isCreating ? 'Tworzenie...' : 'Stwórz' }}</button><button @click="showNewProjectModal = false" class="btn-cancel">Anuluj</button></div>
    </div>
  </div>

  <div v-if="showTutorialModal" class="modal tutorial-modal" @click.self="showTutorialModal = false">
    <div class="modal-content tutorial-content"><button class="close-btn" @click="showTutorialModal = false">✕</button><img :src="tutorialImage" alt="Instrukcja" class="tutorial-img" /></div>
  </div>
</template>

<style scoped>
.launcher{position:fixed;inset:0;display:flex;align-items:center;justify-content:center;z-index:200;background-size:cover;background-position:center}
.launcher h1{color:#4ade80;font-size:32px;margin:0 0 32px 0;text-shadow:0 2px 8px rgba(0,0,0,1),0 0 20px rgba(74,222,128,0.3);font-weight:700;letter-spacing:1px}
.launcher-content{text-align:center;position:relative;z-index:1;background:rgba(13,17,23,0.85);padding:40px 50px;border-radius:8px;border:2px solid #16a34a;box-shadow:0 8px 32px rgba(0,0,0,0.8),0 0 40px rgba(22,163,74,0.2);backdrop-filter:blur(8px)}
.menu-header{width:100px;height:100px;margin-bottom:24px;border-radius:8px;border:3px solid #16a34a;box-shadow:0 0 16px rgba(74,222,128,0.6)}
.launcher-buttons{display:flex;flex-direction:column;gap:12px;width:320px;margin:0 auto}
.launcher-buttons button{padding:14px 24px;font-size:15px;font-weight:700;border:none;cursor:pointer;transition:all 0.15s;border-radius:4px;text-transform:uppercase;letter-spacing:0.5px}
.btn-primary{background:linear-gradient(135deg,#16a34a,#22c55e);color:#fff;box-shadow:0 4px 12px rgba(22,163,74,0.4)}
.btn-primary:hover{background:linear-gradient(135deg,#22c55e,#4ade80);transform:translateY(-2px);box-shadow:0 6px 20px rgba(34,197,94,0.6)}
.btn-secondary{background:#1e293b;color:#e2e8f0;border:2px solid #334155}
.btn-secondary:hover{background:#334155;border-color:#4ade80;transform:translateY(-2px)}
.btn-help{background:#7c3aed;color:#fff;border:2px solid #8b5cf6}
.btn-help:hover{background:#8b5cf6;transform:translateY(-2px);box-shadow:0 4px 12px rgba(139,92,246,0.5)}
.recent{margin-top:32px}.recent h3{color:#94a3b8;font-size:12px;margin:0 0 12px 0;letter-spacing:1px;font-weight:600;text-transform:uppercase}
.recent-list{max-height:280px;overflow-y:auto;overflow-x:hidden;padding-right:8px;margin-right:-8px}
.recent-list::-webkit-scrollbar{width:8px}.recent-list::-webkit-scrollbar-track{background:#0a1628;border-radius:4px}.recent-list::-webkit-scrollbar-thumb{background:#334155;border-radius:4px}.recent-list::-webkit-scrollbar-thumb:hover{background:#4ade80}
.recent-item{padding:12px 16px;background:#1e293b;border:2px solid #334155;border-radius:4px;margin-bottom:8px;cursor:pointer;transition:all 0.15s;color:#cbd5e1;font-size:13px;text-align:left;width:320px;margin-left:auto;margin-right:auto}
.recent-item:hover{background:#334155;border-color:#4ade80;color:#e2e8f0;transform:translateX(4px)}
.recent-path{display:block;font-size:10px;color:#64748b;font-family:'Consolas',monospace;margin-top:4px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.beer-tip{margin-top:32px;padding:12px;background:#fbbf24;color:#78350f;border-radius:4px;font-weight:700;font-size:13px;border:2px solid #f59e0b;box-shadow:0 4px 12px rgba(251,191,36,0.4);cursor:pointer;width:100%;text-transform:uppercase;transition:all 0.15s}
.beer-tip:hover{background:#fcd34d;transform:translateY(-2px) scale(1.02);box-shadow:0 6px 20px rgba(251,191,36,0.6)}
.modal{position:fixed;inset:0;background:rgba(0,0,0,0.9);display:flex;align-items:center;justify-content:center;z-index:300;backdrop-filter:blur(8px)}
.modal-content{background:#1e293b;padding:32px;border-radius:8px;border:2px solid #16a34a;min-width:500px;max-width:90%;width:500px;box-shadow:0 8px 32px rgba(0,0,0,0.8);box-sizing:border-box}
.modal-content h3{margin:0 0 24px 0;color:#4ade80;font-size:22px;font-weight:700}
.form-group{margin-bottom:20px}.form-group label{display:block;margin-bottom:8px;color:#94a3b8;font-size:12px;font-weight:700;text-align:left;text-transform:uppercase;letter-spacing:0.5px}
.form-group input{width:100%;padding:12px;background:#0a1628;border:2px solid #334155;color:#fff;font-size:14px;border-radius:4px;font-family:'Consolas',monospace;box-sizing:border-box;overflow:hidden;text-overflow:ellipsis}
.form-group input:focus{outline:none;border-color:#16a34a;box-shadow:0 0 0 3px rgba(22,163,74,0.1)}
.folder-picker{display:flex;gap:8px}.folder-picker input{flex:1;margin:0;min-width:0}.folder-picker button{padding:12px 20px;background:#334155;border:2px solid #475569;color:white;border-radius:4px;cursor:pointer;white-space:nowrap;font-weight:600;flex-shrink:0}.folder-picker button:hover{background:#475569}
.hint{display:block;margin-top:8px;color:#64748b;font-size:11px;text-align:left;font-family:'Consolas',monospace;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.modal-buttons{display:flex;gap:12px;justify-content:flex-end;margin-top:28px}.modal-buttons button{padding:12px 28px;cursor:pointer;border:none;border-radius:4px;font-weight:700;font-size:14px;text-transform:uppercase}
.btn-cancel{background:#334155;color:#cbd5e1}.btn-cancel:hover{background:#475569}
.tutorial-modal .modal-content{padding:0;max-width:95vw;max-height:95vh;overflow:auto;background:#0d1117}
.tutorial-content{position:relative}.close-btn{position:absolute;top:16px;right:16px;background:#dc2626;color:white;border:none;width:40px;height:40px;border-radius:50%;font-size:20px;cursor:pointer;z-index:10;font-weight:700}.close-btn:hover{background:#ef4444;transform:scale(1.1)}
.tutorial-img{width:100%;height:auto;display:block}
</style>