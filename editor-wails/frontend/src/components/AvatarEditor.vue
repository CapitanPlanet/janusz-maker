<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useProjectStore } from '../stores/projectStore'
import { SelectImageFile, ImportAsset, GetImageBase64, SetProjectPath } from '../../wailsjs/go/main/App'

const store = useProjectStore()
const cache = ref<Record<string,string>>({})
const ops = [
  { value: 'gte', label: '≥' },
  { value: 'lte', label: '≤' },
  { value: 'gt', label: '>' },
  { value: 'lt', label: '<' },
  { value: 'eq', label: '=' },
]
const statDefs = computed(() => store.statsSystem?.stats || [])

function norm(p: string) {
  return (p || '').replace(/\\/g, '/').trim()
}

function getThumb(p: string) {
  if (!p) return ''
  const n = norm(p)
  return cache.value[p] || cache.value[n] || cache.value[p.trim()] || ''
}

const shortName = (p: string) => norm(p).replace('images/av_','').replace('images/','').slice(0,20)

async function loadThumbs(force=false){
  if(!store.projectPath) return
  await SetProjectPath(store.projectPath)
  for(const raw of store.avatarAssets){
    const key = norm(raw)
    if(!key) continue
    if(!force && (cache.value[key] || cache.value[raw])) continue
    try{
      const b64 = await GetImageBase64(key)
      cache.value[key]=b64
      cache.value[raw]=b64
      cache.value[raw.trim()]=b64
    }catch(e){ console.warn('[AV] thumb',key,e) }
  }
}

function close(){
  store.saveProject()
  store.loadAssets()
  store.ui.showAvatarEditor=false
}

async function importAvatar(){
  const f=await SelectImageFile()
  if(!f||!store.projectPath) return
  await SetProjectPath(store.projectPath)

  let newRel=""
  try{
    newRel = await ImportAsset(f, 'av') as string
  }catch(e){
    console.warn('[AV] import fail',e)
    return
  }

  const n = norm(newRel)

  // RETRY - to naprawia czarny kwadrat po imporcie
  for(let i=0;i<6;i++){
    try{
      const b64 = await GetImageBase64(n)
      if(b64 && b64.length>100){
        cache.value[n]=b64
        cache.value[newRel]=b64
        break
      }
    }catch{
      await new Promise(r=>setTimeout(r,150))
    }
  }

  await new Promise(r=>setTimeout(r,200))
  await store.loadAssets()
  await loadThumbs(true)
}

function getIfObj(r:any){return r?.if||{}}
function getFirstStat(r:any){return Object.keys(getIfObj(r))[0]||statDefs.value[0]?.id||'cebula'}
function getFirstOp(r:any){const c=getIfObj(r)[getFirstStat(r)]; return c&&typeof c==='object'?Object.keys(c)[0]:'gte'}
function getFirstVal(r:any){return getIfObj(r)[getFirstStat(r)]?.[getFirstOp(r)]??50}
function updateWhen(r:any,s:string,o:string,v:number){r.if={[s]:{[o]:Number(v)||0}}}
onMounted(()=>loadThumbs(true))
watch(()=>store.avatarAssets.length,()=>loadThumbs(true))
watch(()=>store.ui.showAvatarEditor,v=>{if(v)loadThumbs(true)})
</script>

<template>
  <div v-if="store.ui.showAvatarEditor" class="overlay" @click.self="close">
    <div class="modal">
      <header><h2>🧠 DANE GLOBALNE</h2><button @click="close" class="btn-close">✕</button></header>

      <div class="section highlight stats-edit">
        <div class="head">
          <label>1. STATYSTYKI - CO MA GRACZ</label>
          <button @click="store.addStat()" class="btn-add">+ Nowa statystyka</button>
        </div>
        <div class="hint">NAZWA to co widzi gracz (np. STRES). START to wartość na początku gry.</div>
        <div v-for="st in statDefs" :key="st.id" class="stat-row">
          <input :value="st.name" @input="st.name = ($event.target as HTMLInputElement).value.toUpperCase()" class="stat-name" placeholder="NAZWA" />
          <div class="init-wrap">START:<input type="number" v-model.number="st.initial" class="stat-init" /></div>
          <button @click="store.deleteStat(st.id)" class="btn-del">✕</button>
        </div>
      </div>

      <div class="section highlight">
        <label>2. DOMYŚLNY AVATAR</label>
        <div class="picker-row">
          <div class="preview-box">
            <img v-if="getThumb(store.avatarSystem.default)" :src="getThumb(store.avatarSystem.default)" />
            <span v-else>?</span>
          </div>
          <div class="avatar-window big">
            <div v-for="av in store.avatarAssets" :key="av" :class="['av-tile', { active: av === store.avatarSystem.default }]" @click="store.setDefaultAvatar(av)">
              <img v-if="getThumb(av)" :src="getThumb(av)" />
              <div v-else class="tile-ph"></div>
            </div>
          </div>
        </div>
        <button @click="importAvatar" class="btn-sm">+ IMPORT AVATARA</button>
      </div>

      <div class="section">
        <div class="head">
          <label>3. REGUŁY JANUSZA</label>
          <button @click="store.addAvatarRule()" class="btn-add">+ Reguła</button>
        </div>
        <div v-for="rule in store.sortedAvatarRules" :key="rule.id" class="rule-block">
          <div class="rule-top">
            <span class="if">JEŚLI</span>
            <select :value="getFirstStat(rule)" @change="updateWhen(rule, ($event.target as HTMLSelectElement).value, getFirstOp(rule), getFirstVal(rule))">
              <option v-for="s in statDefs" :key="s.id" :value="s.id">{{ s.name }}</option>
            </select>
            <select :value="getFirstOp(rule)" @change="updateWhen(rule, getFirstStat(rule), ($event.target as HTMLSelectElement).value, getFirstVal(rule))">
              <option v-for="o in ops" :key="o.value" :value="o.value">{{ o.label }}</option>
            </select>
            <input type="number" :value="getFirstVal(rule)" @input="updateWhen(rule, getFirstStat(rule), getFirstOp(rule), Number(($event.target as HTMLInputElement).value))" class="num" />
            <input type="number" v-model.number="rule.priority" class="prio" title="Priorytet" />
            <button @click="store.deleteAvatarRule(rule.id)" class="btn-del">✕</button>
          </div>
          <div class="avatar-window">
            <div v-for="av in store.avatarAssets" :key="av" :class="['av-tile small', { active: av === rule.use }]" @click="rule.use = av">
              <img v-if="getThumb(av)" :src="getThumb(av)" />
              <div v-else class="tile-ph"></div>
            </div>
          </div>
        </div>
      </div>

      <footer><button @click="close" class="btn-save">ZAMKNIJ I ZAPISZ</button></footer>
    </div>
  </div>
</template>

<style scoped>
.overlay{position:fixed;inset:0;background:rgba(0,0,0,.85);display:flex;align-items:center;justify-content:center;z-index:500;backdrop-filter:blur(4px)}
.modal{width:760px;max-width:94vw;max-height:92vh;overflow-y:auto;background:#0D1117;border:1px solid #30363D;border-radius:10px;padding:16px;display:flex;flex-direction:column;gap:16px;box-sizing:border-box}
header{display:flex;justify-content:space-between;align-items:center}header h2{margin:0;color:#00FF94;font-size:16px}
.btn-close{width:32px;height:32px;flex:0 0 32px;background:#dc2626;color:#fff;border:0;border-radius:50%;cursor:pointer;display:inline-flex;align-items:center;justify-content:center;font-size:14px;font-weight:800;line-height:1;padding:0}
.section{display:flex;flex-direction:column;gap:8px}.section.highlight{border:1px solid rgba(0,255,148,.25);background:rgba(0,255,148,.06);border-radius:8px;padding:10px}
.stats-edit{border-color:#94a3b8!important;background:rgba(148,163,184,0.06)!important}
.head{display:flex;justify-content:space-between;align-items:center;gap:10px}
.hint{font-size:11px;color:#7D8590;margin:0}
.stat-row{display:flex;gap:8px;align-items:center;background:#161B22;border:1px solid #21262D;border-radius:6px;padding:6px;box-sizing:border-box}
.stat-name{flex:1 1 0;min-width:0;height:32px;background:#0D1117;border:1px solid #30363D;color:#E6EDF3;border-radius:6px;padding:0 8px;font-weight:700;box-sizing:border-box}
.init-wrap{flex:0 0 112px;height:32px;display:flex;align-items:center;gap:6px;font-size:10px;color:#7D8590;background:#0D1117;border:1px solid #30363D;border-radius:6px;padding:0 8px;box-sizing:border-box}
.stat-init{width:56px;height:100%;background:transparent;border:0;border-left:1px solid #21262D;color:#E6EDF3;text-align:center;outline:none;padding:0;margin:0;box-sizing:border-box}
.picker-row{display:flex;gap:12px;align-items:flex-start}
.preview-box{width:72px;min-width:72px;height:72px;background:#000;border:1px solid #00FF94;border-radius:8px;overflow:hidden;display:flex;align-items:center;justify-content:center;box-sizing:border-box}
.preview-box img{width:100%;height:100%;object-fit:cover}
.avatar-window{display:flex;flex-wrap:wrap;gap:6px;max-height:110px;overflow-y:auto;padding:6px;background:#161B22;border:1px solid #21262D;border-radius:6px;flex:1;align-content:flex-start;box-sizing:border-box}
.avatar-window.big{max-height:140px}
.av-tile{position:relative;width:48px;height:48px;flex:0 0 48px;border-radius:6px;border:2px solid transparent;cursor:pointer;background:#0D1117;overflow:hidden;box-sizing:border-box}
.av-tile:hover{border-color:#30363D}.av-tile.active{border-color:#00FF94;background:rgba(0,255,148,.1)}
.av-tile img{width:100%;height:100%;object-fit:cover;display:block}
.tile-ph{width:100%;height:100%;background:#21262D}
.rule-block{background:#161B22;border:1px solid #21262D;border-radius:8px;padding:8px;display:flex;flex-direction:column;gap:8px;box-sizing:border-box}
.rule-top{display:flex;gap:8px;align-items:center;flex-wrap:nowrap;width:100%;box-sizing:border-box}
.if{font-size:10px;color:#7D8590;font-weight:700;flex:0 0 28px}
.rule-top select{flex:1 1 0;min-width:0;height:32px;background:#0D1117;border:1px solid #30363D;color:#E6EDF3;border-radius:6px;padding:0 6px;font-size:12px;box-sizing:border-box}
.rule-top select:nth-of-type(2){flex:0 0 64px}
.num,.prio{flex:0 0 64px;width:64px;height:32px;background:#0D1117;border:1px solid #30363D;color:#E6EDF3;border-radius:6px;text-align:center;font-size:12px;box-sizing:border-box;padding:0}
.btn-sm,.btn-add,.btn-save{background:#21262D;border:1px solid #30363D;color:#E6EDF3;border-radius:6px;padding:6px 10px;cursor:pointer;font-size:11px;box-sizing:border-box}
.btn-add,.btn-save{background:#00FF94;color:#0D1117;font-weight:700}
.btn-del{width:32px;height:32px;flex:0 0 32px;background:#2A1215;color:#FF7A90;border:1px solid #3A1A20;border-radius:6px;cursor:pointer;box-sizing:border-box;display:inline-flex;align-items:center;justify-content:center;font-size:13px;font-weight:800;line-height:1;padding:0}
footer{display:flex;justify-content:flex-end;border-top:1px solid #21262D;padding-top:12px}
</style>