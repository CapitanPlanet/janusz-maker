<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import SidePanel from './SidePanel.vue'
import ScenesPanel from './ScenesPanel.vue'
import ScenePreview from './ScenePreview.vue'
import ChoicesPanel from './ChoicesPanel.vue'
import AvatarEditor from './AvatarEditor.vue'
import { useProjectStore } from '../stores/projectStore'
import { OpenProjectFolder } from '../../wailsjs/go/main/App'
import tutorialBg from '../assets/bg_tutorial.webp'

const store = useProjectStore()
const emit = defineEmits(['go-to-menu'])
const showSavedToast = ref(false)
const showTutorial = ref(false)

const isLoaded = computed(() => !!store.meta && !!store.projectPath)
const rulesCount = computed(() => store.avatarSystem?.rules?.length ?? 0)

onMounted(() => {
  console.log('[EDITOR] Mounted, meta:', store.meta?.gameName, 'path:', store.projectPath)
})

async function saveProject() {
  try {
    await store.saveProject()
    showSavedToast.value = true
    setTimeout(() => { showSavedToast.value = false }, 2000)
  } catch (e) {
    alert('Błąd zapisu: ' + e)
  }
}

async function openFolder() {
  if (!store.projectPath) return
  try { await OpenProjectFolder(store.projectPath) } catch (e) { console.warn(e) }
}
</script>

<template>
  <div class="editor">
    <header class="top-bar">
      <div class="top-bar-left">
        <h1>{{ store.meta?.gameName || 'HTFFY Editor' }}</h1>
        <div class="project-path clickable" @click="openFolder" title="Otwórz w eksploratorze plików">
          <span class="path-icon">📁</span>
          <span class="path-text">{{ store.projectPath || '...' }}</span>
        </div>
      </div>
      <div class="top-bar-right">
        <button @click="store.ui.showAvatarEditor = true" class="btn-top-janusz">
          🧠 JANUSZ [{{ rulesCount }}]
        </button>
      </div>
    </header>

    <div v-if="isLoaded" class="main-grid">
      <SidePanel class="panel" />
      <ScenesPanel class="panel" />
      <ScenePreview class="panel panel-preview" />
      <ChoicesPanel class="panel" />
    </div>
    <div v-else class="loading">Ładowanie projektu... {{ store.projectPath || '' }}</div>

    <div class="bottom-bar">
      <button @click="saveProject" class="btn-footer save"><span class="icon">💾</span><span>Zapisz projekt</span></button>
      <button @click="emit('go-to-menu')" class="btn-footer menu"><span class="icon">📁</span><span>Menu</span></button>
      <button @click="showTutorial = true" class="btn-footer tutorial"><span class="icon">📖</span><span>Pokaż Tutorial</span></button>
    </div>

    <transition name="toast"><div v-if="showSavedToast" class="toast-saved">✓ Zapisano</div></transition>

    <div v-if="showTutorial" class="modal-overlay" @click.self="showTutorial = false">
      <div class="modal-content tutorial-content">
        <button class="close-btn" @click="showTutorial = false">✕</button>
        <img :src="tutorialBg" alt="Instrukcja" class="tutorial-img" />
      </div>
    </div>

    <AvatarEditor />
  </div>
</template>

<style>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap');
:root { --bg:#0D1117;--panel:#161B22;--card:#21262D;--input:#21262D;--border:#30363D;--accent:#00FF94;--text:#E6EDF3;--text-dim:#7D8590;--danger:#F85149;--steel:#94a3b8; }
* { box-sizing:border-box; font-family:'Inter',system-ui,sans-serif; }
body { background:var(--bg); margin:0; color:var(--text); overflow-x:hidden; }
</style>

<style scoped>
.editor { height:100vh; display:flex; flex-direction:column; background:var(--bg); }
.main-grid { display:grid; grid-template-columns:280px 280px minmax(400px,1fr) 380px; gap:12px; padding:12px; flex:1; overflow:hidden; min-height:0; padding-bottom:72px; min-width:1376px; }
.top-bar { padding:12px 20px; border-bottom:1px solid var(--border); background:var(--panel); flex-shrink:0; display:flex; align-items:center; justify-content:space-between; }
.top-bar-left { display:flex; flex-direction:column; }
.top-bar h1 { margin:0; font-size:16px; font-weight:600; }
.project-path { font-size:11px; color:var(--text-dim); font-family:monospace; margin-top:2px; }
.project-path.clickable { display:inline-flex; gap:6px; cursor:pointer; padding:2px 6px; margin-left:-6px; border-radius:4px; }
.project-path.clickable:hover { background:var(--card); color:var(--accent); }
.top-bar-right { display:flex; align-items:center; }
.btn-top-janusz { background:#161B22; border:1px solid var(--accent); color:var(--accent); padding:6px 12px; border-radius:6px; font-size:12px; font-weight:700; cursor:pointer; }
.panel { background:var(--panel); border:1px solid var(--border); border-radius:8px; overflow:hidden; display:flex; flex-direction:column; min-height:0; }
.loading { flex:1; display:flex; align-items:center; justify-content:center; color:var(--text-dim); }
.bottom-bar { position:fixed; bottom:0; left:0; right:0; height:56px; background:#1e293b; border-top:2px solid #0f1419; display:flex; gap:12px; padding:0 20px; align-items:center; z-index:100; }
.btn-footer { height:40px; padding:0 20px; background:#2d3748; border:1px solid #4a5568; color:#cbd5e0; cursor:pointer; font-size:12px; font-weight:700; border-radius:3px; display:flex; gap:8px; align-items:center; }
.toast-saved { position:fixed; bottom:80px; right:20px; background:var(--card); border:1px solid var(--accent); color:var(--accent); padding:10px 16px; border-radius:6px; z-index:200; }
.modal-overlay { position:fixed; inset:0; background:rgba(0,0,0,0.9); display:flex; align-items:center; justify-content:center; z-index:300; }
.modal-content { background:#0d1117; border-radius:8px; border:2px solid var(--steel); max-width:95vw; max-height:95vh; overflow:auto; }
.close-btn { position:absolute; top:16px; right:16px; background:#dc2626; color:white; border:none; width:40px; height:40px; border-radius:50%; cursor:pointer; }
.tutorial-img { width:100%; display:block; }
</style>