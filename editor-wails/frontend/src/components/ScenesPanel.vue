<script setup lang="ts">
import { useProjectStore } from '../stores/projectStore'

const store = useProjectStore()

function goToScene(sceneId: string) {
  store.selectScene(sceneId)
}

function addScene() {
  const sceneName = prompt('ID nowej sceny:', 'nowa_scena')
  if (!sceneName) return
  const exists = store.currentDayScenes.some(s => s.Id === sceneName)
  if (exists) {
    alert(`Scena "${sceneName}" już istnieje!`)
    return
  }
  store.addSceneToCurrentDay(sceneName)
}

function addEndDay() {
  // 1. Sprawdź czy już jest koniec dnia w tym dniu - ma być tylko jeden
  const existingEnd = store.currentDayScenes.find(s => s.IsEndDay || s.Id.startsWith('koniec_dnia_'))
  if (existingEnd) {
    alert(`Masz już koniec dnia w ${store.currentDay}: ${existingEnd.Id}. Edytuj go zamiast tworzyć nowy.`)
    goToScene(existingEnd.Id)
    return
  }

  // 2. Wyciągnij numer dnia z store.currentDay -> "day1" => 1
  const dayNum = parseInt(store.currentDay.replace('day','')) || 1
  const nextDayNum = dayNum + 1
  const endId = `koniec_dnia_${dayNum}`
  const nextId = `start_d${nextDayNum}`

  // 3. Użyj akcji ze store'a - przekazujemy pełny obiekt zgodny z Twoim JSONem
  store.addSceneToCurrentDay(endId, {
    SceneTitle: `KONIEC DNIA ${dayNum}`,
    Text: "Koniec dnia. Idziesz spać.",
    Background: "images/bg_sen_ jaanusza.jpg",
    IsEndDay: true,
    Type: "end_of_day",
    Day: dayNum,
    NextDayId: nextId,
    Transfers: {
      keep_flags: [],
      keep_stats: ["Cebula", "Wstyd", "Portfel", "Reputacja"],
      summary_text: ""
    },
    Choices: [
      {
        Text: "Śpij",
        Next: nextId,
        ReactionText: "",
        ReactionImage: "",
        SoundFile: ""
      }
    ]
  })

  goToScene(endId)
}

function deleteScene(sceneId: string) {
  store.deleteScene(sceneId)
}
</script>

<template>
  <aside class="scenes-panel">
    <div class="panel-header">
      <h3>Sceny [{{ store.currentDayScenes.length }}]</h3>
      <span class="day-badge">{{ store.currentDay }}</span>
    </div>

    <div class="btn-group">
      <button @click="addScene" class="btn-add">+ Scena</button>
      <button @click="addEndDay" class="btn-add-end">🌙 Koniec Dnia</button>
    </div>

    <div
      v-for="scene in store.currentDayScenes"
      :key="scene.Id"
      :class="['scene-item', { active: scene.Id === store.currentSceneId, 'is-end-day': scene.IsEndDay }]"
      @click="goToScene(scene.Id)"
    >
      <div class="scene-id">{{ scene.Id }}</div>
      <div class="scene-title">{{ scene.SceneTitle || '(bez tytułu)' }}</div>
      <div v-if="scene.IsEndDay" class="end-badge">🌙 KONIEC DNIA → {{ scene.NextDayId || scene.Choices?.[0]?.Next }}</div>
      <button @click.stop="deleteScene(scene.Id)" class="btn-delete-mini">✕</button>
    </div>
  </aside>
</template>

<style scoped>
.scenes-panel {
  background: rgba(10, 22, 40, 0.75);
  padding: 16px;
  overflow-y: auto;
  border: 1px solid rgba(51, 65, 85, 0.3);
  display: flex;
  flex-direction: column;
}
.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.scenes-panel h3 {
  margin: 0;
  color: #4ade80;
  font-size: 14px;
}
.day-badge {
  font-size: 11px;
  font-family: monospace;
  color: #94a3b8;
  background: rgba(51, 65, 85, 0.5);
  padding: 2px 6px;
  border-radius: 3px;
}
.btn-group {
  display: flex;
  gap: 6px;
  margin-bottom: 12px;
}
.btn-add {
  flex: 1;
  padding: 6px 10px;
  background: #16a34a;
  border: none;
  color: #fff;
  cursor: pointer;
  font-size: 12px;
}
.btn-add:hover { background: #22c55e; }
.btn-add-end {
  flex: 1;
  padding: 6px 10px;
  background: #1e293b;
  border: 1px solid #fbbf24;
  color: #fbbf24;
  cursor: pointer;
  font-size: 12px;
}
.btn-add-end:hover { background: #334155; }
.scene-item {
  padding: 10px;
  background: rgba(30, 41, 59, 0.8);
  margin-bottom: 6px;
  cursor: pointer;
  border-left: 3px solid transparent;
  position: relative;
  transition: all 0.2s;
}
.scene-item:hover {
  background: rgba(51, 65, 85, 0.9);
  border-left-color: #4ade80;
}
.scene-item.active {
  background: rgba(22, 163, 74, 0.8);
  border-left-color: #4ade80;
}
.scene-item.is-end-day {
  background: rgba(30, 41, 59, 0.9);
  border: 1px dashed #fbbf24;
  border-left: 3px solid #fbbf24;
}
.scene-item.is-end-day.active {
  background: rgba(251, 191, 36, 0.15);
}
.scene-id {
  font-family: monospace;
  font-size: 11px;
  color: #94a3b8;
}
.scene-title {
  font-size: 13px;
  margin-top: 2px;
  color: #fff;
  padding-right: 20px;
}
.end-badge {
  font-size: 9px;
  color: #fbbf24;
  margin-top: 4px;
  font-weight: bold;
}
.btn-delete-mini {
  position: absolute;
  right: 6px;
  top: 6px;
  background: #dc2626;
  border: none;
  color: #fff;
  font-size: 10px;
  padding: 2px 6px;
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.2s;
}
.scene-item:hover.btn-delete-mini {
  opacity: 1;
}
</style>