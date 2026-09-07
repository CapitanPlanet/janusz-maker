<script setup lang="ts">
import { computed } from 'vue'
import { useProjectStore } from './stores/projectStore'
import LauncherView from './components/LauncherView.vue'
import EditorView from './components/EditorView.vue'

const projectStore = useProjectStore()

// MUSI sprawdzać meta, nie tylko path!
const isProjectLoaded = computed(() => !!projectStore.projectPath && !!projectStore.meta)

function closeProject() {
  projectStore.closeProject()
}
</script>

<template>
  <div v-if="!isProjectLoaded">
    <LauncherView />
    <!-- debug żebyś widział co się dzieje -->
    <div v-if="projectStore.projectPath && !projectStore.meta" style="position:fixed; bottom:10px; left:10px; color:#7D8590; font-size:11px; font-family:monospace;">
      Ładowanie {{ projectStore.projectPath }}...
    </div>
  </div>
  <EditorView v-else @go-to-menu="closeProject" />
</template>

<style>
html, body, #app { margin:0; padding:0; width:100%; height:100%; overflow:auto; background:#0D1117; }
* { box-sizing:border-box; }
::-webkit-scrollbar { width:10px; height:10px; }
::-webkit-scrollbar-track { background:#161B22; }
::-webkit-scrollbar-thumb { background:#30363D; border-radius:5px; }
::-webkit-scrollbar-thumb:hover { background:#484F58; }
</style>