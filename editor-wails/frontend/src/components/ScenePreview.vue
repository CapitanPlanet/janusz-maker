<script setup>
import { ref, computed, watch } from 'vue'
import { useProjectStore } from '../stores/projectStore'
import { GetImageBase64 } from '../../wailsjs/go/main/App'

const store = useProjectStore()
const bgCache = ref({})
const reactionCache = ref({})
const hoveredChoice = ref(null)

const bgUrl = computed(() => {
  const bg = store.currentScene?.Background
  if (!bg ||!store.projectPath) return ''
  const key = bg.trim()
  return bgCache.value[key] || bgCache.value[bg] || ''
})

const choices = computed(() => {
  return store.currentScene?.Choices || []
})

const hoveredReactionUrl = computed(() => {
  const choice = hoveredChoice.value
  if (!choice?.ReactionImage) return ''
  const key = choice.ReactionImage.trim()
  return reactionCache.value[key] || reactionCache.value[choice.ReactionImage] || ''
})

function cleanPath(p) {
  if (!p) return ''
  return p.trim()
}

// Ładuj tło
watch(() => store.currentScene?.Background, async (newBg) => {
  const clean = cleanPath(newBg)
  if (!clean ||!store.projectPath) return
  if (bgCache.value[clean]) return

  try {
    // WAŻNE: 1 argument, nie 2!
    const b64 = await GetImageBase64(clean)
    bgCache.value[clean] = b64
    bgCache.value[newBg] = b64 // cache też pod oryginalnym kluczem
  } catch (e) {
    console.warn('[PREVIEW] Brak tła:', clean, e)
    bgCache.value[clean] = ''
    bgCache.value[newBg] = ''
  }
}, { immediate: true })

// Ładuj reakcje dla wyborów
watch(choices, async (newChoices) => {
  if (!newChoices?.length ||!store.projectPath) return

  for (const choice of newChoices) {
    const raw = cleanPath(choice.ReactionImage)
    if (!raw) continue
    if (reactionCache.value[raw]) continue

    try {
      const b64 = await GetImageBase64(raw)
      reactionCache.value[raw] = b64
      reactionCache.value[choice.ReactionImage] = b64
    } catch (e) {
      console.warn('[PREVIEW] Brak reakcji:', raw, e)
      reactionCache.value[raw] = ''
    }
  }
}, { immediate: true, deep: true })
</script>

<template>
  <div class="scene-preview" v-if="store.currentScene">
    <div class="preview-stage" :style="{ backgroundImage: bgUrl? `url(${bgUrl})` : 'none' }">
      <div v-if="!bgUrl" class="no-bg">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
          <circle cx="8.5" cy="8.5" r="1.5"></circle>
          <polyline points="21 15 16 10 5 21"></polyline>
        </svg>
        <span>Brak tła</span>
      </div>

      <!-- REAKCJA PODGLĄD NA HOVER -->
      <transition name="reaction">
        <div v-if="hoveredReactionUrl" class="reaction-overlay">
          <img :src="hoveredReactionUrl" alt="reakcja" />
          <div class="reaction-label">{{ hoveredChoice?.ReactionImage?.replace('images/re_', '').replace('images/', '') }}</div>
        </div>
      </transition>

      <div class="preview-text">
        <h2>{{ store.currentScene.SceneTitle }}</h2>
        <p>{{ store.currentScene.Text }}</p>
      </div>

      <!-- WYBORY NA DOLE EKRANU GRY -->
      <div v-if="choices.length" class="preview-choices">
        <div
          v-for="(choice, idx) in choices"
          :key="choice.id || idx"
          class="choice-btn"
          @mouseenter="hoveredChoice = choice"
          @mouseleave="hoveredChoice = null"
          :class="{ 'has-reaction': choice.ReactionImage }"
        >
          <span class="choice-num">{{ idx + 1 }}</span>
          <span class="choice-text">{{ choice.Text || 'Pusty wybór' }}</span>
          <span v-if="choice.ReactionImage" class="choice-reaction-icon">🖼</span>
        </div>
      </div>
    </div>
  </div>
  <div v-else class="no-scene">
    Wybierz scenę z listy
  </div>
</template>

<style scoped>
.scene-preview {
  background: var(--panel);
  border-radius: 8px;
  overflow: hidden;
  height: 100%;
}

.preview-stage {
  position: relative;
  width: 100%;
  height: 100%;
  background-size: contain;
  background-repeat: no-repeat;
  background-position: center;
  background-color: #000;
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
}

.no-bg {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: var(--text-dim);
  gap: 12px;
}

/* TEKST SCENY */
.preview-text {
  position: relative;
  z-index: 2;
  background: linear-gradient(180deg, transparent 0%, rgba(0,0,0,0.9) 100%);
  padding: 16px 24px 12px 24px;
}

.preview-text h2 {
  color: var(--accent);
  font-size: 12px;
  margin: 0 0 6px 0;
  text-transform: uppercase;
  letter-spacing: 1px;
  text-shadow: 0 2px 6px rgba(0,0,0,1);
  font-weight: 700;
}

.preview-text p {
  color: #fff;
  font-size: 14px;
  line-height: 1.5;
  margin: 0;
  white-space: pre-wrap;
  text-shadow: 0 2px 6px rgba(0,0,0,1);
}

/* REAKCJA OVERLAY */
.reaction-overlay {
  position: absolute;
  top: 50%;
  right: 20px;
  transform: translateY(-50%);
  z-index: 10;
  background: rgba(0,0,0,0.8);
  border: 2px solid var(--steel);
  border-radius: 8px;
  padding: 8px;
  box-shadow: 0 8px 24px rgba(0,0,0,0.8);
  max-width: 200px;
}

.reaction-overlay img {
  width: 180px;
  height: 180px;
  object-fit: cover;
  border-radius: 4px;
  display: block;
}

.reaction-label {
  font-size: 10px;
  color: var(--text-dim);
  font-family: monospace;
  margin-top: 6px;
  text-align: center;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.reaction-enter-active,
.reaction-leave-active {
  transition: all 0.2s ease;
}

.reaction-enter-from,
.reaction-leave-to {
  opacity: 0;
  transform: translateY(-50%) scale(0.9);
}

.preview-choices {
  position: relative;
  z-index: 2;
  background: rgba(0,0,0,0.85);
  border-top: 1px solid #1e293b;
  padding: 12px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  max-height: 40%;
  overflow-y: auto;
}

.preview-choices::-webkit-scrollbar { width: 6px; }
.preview-choices::-webkit-scrollbar-track { background: rgba(30, 41, 59, 0.4); }
.preview-choices::-webkit-scrollbar-thumb { background: #475569; border-radius: 3px; }

/* jak jest nieparzysta liczba to ostatni na całą szerokość */
.choice-btn:last-child:nth-child(odd) {
  grid-column: 1 / -1;
}

.choice-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  background: linear-gradient(180deg, #2d3748 0%, #1a202c 100%);
  border: 1px solid #4a5568;
  border-top-color: #718096;
  border-radius: 4px;
  color: #cbd5e0;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s;
  box-shadow: 0 2px 0 #0f1419, inset 0 1px 0 rgba(255,255,255,0.05);
  text-align: left;
}

.choice-btn:hover {
  background: linear-gradient(180deg, #4a5568 0%, #2d3748 100%);
  border-color: var(--accent);
  color: #fff;
  transform: translateY(-1px);
  box-shadow: 0 3px 0 #0f1419, 0 0 12px rgba(0, 255, 148, 0.2), inset 0 1px 0 rgba(255,255,255,0.1);
}

.choice-btn:active {
  transform: translateY(2px);
  box-shadow: inset 0 2px 4px rgba(0,0,0,0.5);
}

.choice-btn.has-reaction {
  border-left: 3px solid var(--accent);
}

.choice-num {
  width: 22px;
  height: 22px;
  background: rgba(0,0,0,0.5);
  border: 1px solid #334155;
  border-radius: 50%;
  display: grid;
  place-items: center;
  font-size: 11px;
  font-weight: 700;
  flex-shrink: 0;
}

.choice-btn:hover.choice-num {
  background: var(--accent);
  color: #000;
  border-color: var(--accent);
}

.choice-text {
  flex: 1;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.choice-reaction-icon {
  font-size: 12px;
  opacity: 0.6;
  flex-shrink: 0;
}

.choice-btn:hover.choice-reaction-icon {
  opacity: 1;
}

.no-scene {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-dim);
}
</style>