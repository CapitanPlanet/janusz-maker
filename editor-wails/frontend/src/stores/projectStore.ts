import { defineStore } from 'pinia'
import { CreateProject, ReadJSON, WriteJSON, ListFiles, ListAssets, ListAudioAssets, DeleteFile } from '../../wailsjs/go/main/App'

export interface StatDef { id: string; name: string; initial: number }
interface Choice {
  id?: string;
  Text: string;
  Next: string;
  NextDayId?: string;
  NextDay?: string;
  Stats?: Record<string, number>;
  ReactionText?: string;
  ReactionImage?: string;
  SoundFile?: string;
  FlagsSet?: string[];
  FlagsRequired?: string[];
  MinPortfel?: number | null;
  KosztPortfel?: number | null;
  FailText?: string;
  [key: string]: any
}

export interface EndDayTransfers {
  keep_flags: string[]
  keep_stats: string[]
  summary_text?: string
  nextDay?: string
}

interface Scene {
  Id: string;
  SceneTitle?: string;
  Background?: string;
  Text?: string;
  Choices?: Choice[];
  IsEndDay?: boolean;
  Type?: 'normal' | 'end_of_day';
  Day?: number;
  NextDay?: string;
  NextDayId?: string;
  Transfers?: EndDayTransfers | null;
  [key: string]: any
}

type ConditionOp = { gte?: number; lte?: number; gt?: number; lt?: number; eq?: number }
export interface AvatarRule { id: string; use: string; if: Record<string, ConditionOp>; priority: number; when?: any }
export interface AvatarSystem { default: string; rules: AvatarRule[] }
export interface StatsSystem { stats: StatDef[] }
interface ProjectMeta { gameName: string; author: string; version: string; engineVersion: string; startDay: string; startScene: string }

const DEFAULT_STATS: StatDef[] = [
  { id: 'CEBULA', name: 'CEBULA', initial: 0 },
  { id: 'WSTYD', name: 'WSTYD', initial: 0 },
  { id: 'PORTFEL', name: 'PORTFEL', initial: 0 },
  { id: 'REPUTACJA', name: 'REPUTACJA', initial: 0 },
]

const LEGACY_MAP: Record<string,string> = { A:'CEBULA', B:'WSTYD', C:'PORTFEL', D:'REPUTACJA', PORTFEL1:'PORTFEL' }

function toSafeName(raw: string): string {
  const up = raw.toString().trim().toUpperCase()
  return LEGACY_MAP[up] || up
}

function getBasename(path: string): string {
  return path.replace(/\\/g,'/').split('/').pop() || ''
}

function isImageOfType(path: string, prefix: 'bg_'|'re_'|'av_'): boolean {
  const base = getBasename(path).toLowerCase()
  return base.startsWith(prefix)
}

function fixRelPath(input: string): string {
  if (!input) return ''
  let p = input.replace(/\\/g,'/').trim()
  // Strip absolute Windows / Unix
  if (p.includes(':') || p.startsWith('/') || p.startsWith('C:')) {
    // keep only images/... or sounds/... or just filename
    const idxImg = p.toLowerCase().lastIndexOf('images/')
    const idxSnd = p.toLowerCase().lastIndexOf('sounds/')
    if (idxImg >= 0) p = p.substring(idxImg)
    else if (idxSnd >= 0) p = p.substring(idxSnd)
    else p = 'images/' + p.split('/').pop()!
  }
  return p.replace(/^\/+/, '')
}

export const useProjectStore = defineStore('project', {
  state: () => ({
    projectPath: null as string | null,
    meta: null as ProjectMeta | null,
    avatarSystem: { default: '', rules: [] } as AvatarSystem,
    statsSystem: { stats: [...DEFAULT_STATS] } as StatsSystem,
    days: {} as Record<string, Scene[]>,
    currentDay: 'day1',
    currentSceneId: null as string | null,
    assets: { all: [] as string[], images: [] as string[], sounds: [] as string[] },
    saveStatus: '',
    ui: { showAvatarEditor: false }
  }),
  getters: {
    isProjectLoaded: (s) => !!s.projectPath,
    currentDayScenes: (s) => s.days[s.currentDay] || [],
    currentScene: (s) => {
      const scenes = s.days[s.currentDay] || []
      return scenes.find(x => x.Id === s.currentSceneId) || null
    },
    sceneIdsInCurrentDay: (s) => (s.days[s.currentDay] || []).map(x => x.Id),
    dayFileList: (s) => Object.keys(s.days),
    // FIX: filtruj po basename, nie includes
    backgroundAssets: (s) => s.assets.images.filter(a => isImageOfType(a, 'bg_')),
    reactionAssets: (s) => s.assets.images.filter(a => isImageOfType(a, 're_')),
    avatarAssets: (s) => s.assets.images.filter(a => isImageOfType(a, 'av_')),
    availableBackgrounds: (s) => s.assets.images.filter(a => isImageOfType(a, 'bg_')),
    availableSounds: (s) => s.assets.sounds,
    sortedAvatarRules(s): AvatarRule[] { return [...s.avatarSystem.rules].sort((a,b) => b.priority - a.priority) },
    statList: (s) => s.statsSystem.stats,
    endDaySceneInCurrentDay: (s) => (s.days[s.currentDay] || []).find(x => x.IsEndDay || x.Type === 'end_of_day' || x.Id.startsWith('koniec_dnia_'))
  },
  actions: {
    selectScene(id: string) { this.currentSceneId = id },
    ensureChoiceIds(scene: Scene) {
      scene.Choices?.forEach(c => { if(!c.id) c.id = (crypto as any).randomUUID?.() || Math.random().toString(36).slice(2) })
    },
    createEmptyStats(): Record<string, number> {
      const obj: Record<string, number> = {}
      this.statsSystem.stats.forEach(s => { obj[s.name] = 0 })
      return obj
    },
    ensureStatsSystem() {
      if (!this.statsSystem || !Array.isArray(this.statsSystem.stats) || !this.statsSystem.stats.length) {
        this.statsSystem = { stats: [...DEFAULT_STATS] }
        return
      }
      const used = new Set<string>()
      const fixed: StatDef[] = []
      for(const s of this.statsSystem.stats){
        if(!s.name && !s.id) continue
        const finalName = toSafeName((s.name || s.id) as string)
        if(!finalName) continue
        if(used.has(finalName)) continue
        used.add(finalName)
        fixed.push({ id: finalName, name: finalName, initial: s.initial ?? 0 })
      }
      this.statsSystem.stats = fixed.length ? fixed : [...DEFAULT_STATS]
    },
    ensureAvatarSystem() {
      this.ensureStatsSystem()
      if (!this.avatarSystem) this.avatarSystem = { default: '', rules: [] }
      if (!this.avatarSystem.rules) this.avatarSystem.rules = []
      this.avatarSystem.default = fixRelPath(this.avatarSystem.default || '')
      if(!this.avatarSystem.default && this.avatarAssets.length > 0){
        this.avatarSystem.default = this.avatarAssets[0]
      }
      // Fix legacy stat keys in rules
      this.avatarSystem.rules.forEach(r=>{
        if(r.if){
          const newIf: Record<string, ConditionOp> = {}
          for(const [k,v] of Object.entries(r.if)){
            const final = toSafeName(k)
            if(this.statsSystem.stats.some(s=>s.name===final)){
              newIf[final] = v as ConditionOp
            }
          }
          r.if = newIf
        }
        if(r.use) r.use = fixRelPath(r.use)
      })
    },
    migrateChoiceStats(choice: any) {
      if (!choice.Stats) choice.Stats = {}
      const newStats: Record<string, number> = {}
      for (const [k,v] of Object.entries(choice.Stats as Record<string,any>)) {
        if(typeof v !== 'number') continue
        const finalKey = toSafeName(k)
        if(this.statsSystem.stats.some(s=>s.name===finalKey)){
          newStats[finalKey] = v as number
        }
      }
      // cleanup legacy props
      const toDelete = ['Cebula','Wstyd','Portfel','Reputacja','cebula','wstyd','portfel','reputacja','a','b','c','d','A','B','C','D']
      toDelete.forEach(k => delete choice[k])
      choice.Stats = newStats
      this.statsSystem.stats.forEach(s=>{ if(choice.Stats[s.name]==null) choice.Stats[s.name]=0 })
      // FIX: fix paths in choice
      if(choice.ReactionImage) choice.ReactionImage = fixRelPath(choice.ReactionImage)
      if(choice.SoundFile) choice.SoundFile = fixRelPath(choice.SoundFile)
    },
    normalizeScene(scene: Scene): Scene {
      // FIX: fix all asset paths in scene
      if(scene.Background) scene.Background = fixRelPath(scene.Background)

      const isEnd = scene.Id.startsWith('koniec_dnia_') || scene.Id.startsWith('koniec_dnia') || scene.IsEndDay || scene.Type === 'end_of_day'
      if (isEnd) {
        scene.IsEndDay = true
        scene.Type = 'end_of_day'
        if (!scene.NextDayId) {
          const dayNum = scene.Day || parseInt(this.currentDay.replace('day','')) || 1
          scene.NextDayId = `day${dayNum+1}`
          scene.NextDay = `day${dayNum+1}`
        }
        if (!scene.Transfers) {
          scene.Transfers = { keep_flags: [], keep_stats: this.statsSystem.stats.map(s=>s.name), summary_text: '' }
        } else {
          scene.Transfers.keep_stats = this.statsSystem.stats.map(s=>s.name)
          if(!scene.Transfers.keep_flags) scene.Transfers.keep_flags = []
        }
        if(!scene.Choices?.length){
          scene.Choices = [{
            Text: 'Śpij',
            Next: 'END_DAY',
            NextDayId: scene.NextDayId,
            Stats: this.createEmptyStats(),
            ReactionText: '',
            ReactionImage: '',
            SoundFile: ''
          }]
        }
      }
      if (scene.IsEndDay && !scene.Type) scene.Type = 'end_of_day'
      return scene
    },
    async loadAssets() {
      if (!this.projectPath) { this.assets.all=[]; this.assets.images=[]; this.assets.sounds=[]; return }
      try {
        const [images, sounds] = await Promise.all([
          ListAssets(this.projectPath).catch(()=>[] as string[]),
          ListAudioAssets(this.projectPath).catch(()=>[] as string[])
        ])
        // FIX: ListAssets returns images, ListAudioAssets returns sounds - merge correctly
        this.assets.images = (images || []).map(fixRelPath)
        this.assets.sounds = (sounds || []).map(fixRelPath)
        this.assets.all = [...this.assets.images, ...this.assets.sounds]
      } catch {
        this.assets.all=[]; this.assets.images=[]; this.assets.sounds=[]
      }
      this.ensureAvatarSystem()
    },
    async createProjectAtPath(p: string, name: string) {
      p = p.replace(/\\/g,'/').replace(/\/+/g,'/').replace(/\/$/,'')
      try { await CreateProject(p, name) } catch {}
      await this.loadProjectFromPath(p)
      return p
    },
    async deleteDay(dayId: string) {
      if (this.dayFileList.length <= 1) { alert('Musisz zostawić minimum 1 dzień!'); return }
      if (this.projectPath) { try { await DeleteFile(this.projectPath, `Data/${dayId}.json`) } catch {} }
      delete this.days[dayId]
      if (this.currentDay === dayId) { this.currentDay = Object.keys(this.days)[0]; this.currentSceneId = this.days[this.currentDay]?.[0]?.Id || null }
    },
    async loadProjectFromPath(path: string) {
      path = path.replace(/\\/g,'/').replace(/\/+/g,'/').replace(/\/$/,'')
      this.projectPath = path
      try {
        const metaRaw = await ReadJSON(`${path}/project.janproj`)
        const parsed = JSON.parse(metaRaw)
        this.meta = {
          gameName: parsed.gameName||'',
          author: parsed.author||'',
          version: parsed.version||'1.0.0',
          engineVersion: parsed.engineVersion||'2.0.0',
          startDay: parsed.startDay||'day1',
          startScene: parsed.startScene||'start'
        }
        let incomingStats: StatDef[] = parsed.statsSystem?.stats || [...DEFAULT_STATS]
        const migrated: StatDef[] = incomingStats.map(s=>{
          const finalName = toSafeName((s.name||s.id||'').toString())
          return { id: finalName, name: finalName, initial: s.initial??0 }
        }).filter(s=>s.name)
        this.statsSystem = { stats: migrated.length ? migrated : [...DEFAULT_STATS] }
        this.ensureStatsSystem()
        this.avatarSystem = {
          default: fixRelPath(parsed.avatarSystem?.default||''),
          rules: Array.isArray(parsed.avatarSystem?.rules)?parsed.avatarSystem.rules:[]
        }
        this.avatarSystem.rules.forEach((r:any)=>{
          if(!r.id) r.id = (crypto as any).randomUUID?.() || Math.random().toString(36).slice(2)
          if(r.priority==null) r.priority = 0
          if(r.use) r.use = fixRelPath(r.use)
        })
        let dayFiles: string[] = []
        try { dayFiles = await ListFiles(`${path}/Data`, '.json') || [] } catch { dayFiles = ['day1.json'] }
        dayFiles = dayFiles.filter(f => !f.startsWith('_') && !f.startsWith('.'))
        this.days = {}
        for (const file of dayFiles) {
          const dayName = file.replace('.json','')
          try {
            const raw = await ReadJSON(`${path}/Data/${file}`)
            const data = JSON.parse(raw) as Scene[]
            data.forEach(s => { this.normalizeScene(s); this.ensureChoiceIds(s); s.Choices?.forEach(c => this.migrateChoiceStats(c)) })
            this.days[dayName] = data
          } catch {}
        }
        this.currentDay = this.meta.startDay || Object.keys(this.days)[0] || 'day1'
        this.currentSceneId = this.meta.startScene || this.days[this.currentDay]?.[0]?.Id || null
        await this.loadAssets()
        this.ensureAvatarSystem()
      } catch (e) { console.error(e); this.projectPath = null; throw e }
    },
    async scanAssets() {
      // DEPRECATED - use loadAssets, kept for compatibility
      await this.loadAssets()
    },
    async saveProject() {
      if (!this.projectPath || !this.meta) return
      this.saveStatus = 'Zapisywanie...'
      try {
        this.ensureAvatarSystem(); this.ensureStatsSystem()
        Object.values(this.days).forEach(scenes => {
          scenes.forEach(scene => {
            this.normalizeScene(scene)
            scene.Choices?.forEach((c:any) => this.migrateChoiceStats(c))
          })
        })
        const toSave = {
          ...this.meta,
          avatarSystem: { default: fixRelPath(this.avatarSystem.default), rules: this.avatarSystem.rules.map(r=>({...r, use: fixRelPath(r.use)})) },
          statsSystem: { stats: this.statsSystem.stats }
        }
        await WriteJSON(`${this.projectPath}/project.janproj`, JSON.stringify(toSave, null, 2))
        for (const dayFile of Object.keys(this.days)) {
          const cleaned = this.days[dayFile].map(scene => ({
            ...scene,
            Background: fixRelPath(scene.Background || ''),
            Choices: scene.Choices?.map((c:any) => {
              const {id,...cleanChoice}=c
              return {
                ...cleanChoice,
                ReactionImage: fixRelPath(cleanChoice.ReactionImage||''),
                SoundFile: fixRelPath(cleanChoice.SoundFile||'')
              }
            })
          }))
          await WriteJSON(`${this.projectPath}/Data/${dayFile}.json`, JSON.stringify(cleaned, null, 2))
        }
        const manifest = {
          days: Object.keys(this.days),
          startDay: this.meta.startDay || Object.keys(this.days)[0] || 'day1',
          startScene: this.meta.startScene || 'start',
          version: new Date().toISOString()
        }
        await WriteJSON(`${this.projectPath}/Data/_manifest.json`, JSON.stringify(manifest, null, 2))
        this.saveStatus='Zapisano + manifest'; setTimeout(()=>{this.saveStatus=''},2500)
      } catch (e) { console.error(e); this.saveStatus='Błąd zapisu: '+e }
    },
    addStat() {
      const used = new Set(this.statsSystem.stats.map(s=>s.name))
      let newName = `STAT_${Date.now().toString(36).toUpperCase()}`
      let counter = 0
      while(used.has(newName) && counter < 10){ newName = `STAT_${Date.now().toString(36).toUpperCase()}_${counter++}` }
      if(!used.has(newName)){
        this.statsSystem.stats.push({ id: newName, name: newName, initial: 0 })
      }
      Object.values(this.days).forEach(scenes => scenes.forEach(scene => scene.Choices?.forEach(c => { if(!c.Stats) c.Stats={}; if(c.Stats[newName]==null) c.Stats[newName]=0 })))
    },
    deleteStat(id: string) {
      if (this.statsSystem.stats.length <= 1) { alert('Musi zostać przynajmniej 1 statystyka'); return }
      this.statsSystem.stats = this.statsSystem.stats.filter(s => s.id!==id && s.name!==id)
      Object.values(this.days).forEach(scenes => scenes.forEach(scene => scene.Choices?.forEach((c:any) => { if(c.Stats && c.Stats[id]!=null) delete c.Stats[id] })))
    },
    setDefaultAvatar(path: string) { this.avatarSystem.default = fixRelPath(path) },
    addAvatarRule(imagePath?: string) {
      const maxPrio = Math.max(0,...this.avatarSystem.rules.map(r => r.priority))
      const firstStat = this.statsSystem.stats[0]?.name || 'CEBULA'
      this.avatarSystem.rules.push({
        id: (crypto as any).randomUUID?.() || Math.random().toString(36).slice(2),
        use: fixRelPath(imagePath || this.avatarAssets[0] || this.avatarSystem.default || ''),
        if: { [firstStat]: { gte: 50 } },
        priority: maxPrio+10
      })
    },
    updateAvatarRule(ruleId: string, patch: Partial<AvatarRule>) { const rule = this.avatarSystem.rules.find(r => r.id===ruleId); if(rule) Object.assign(rule, patch) },
    deleteAvatarRule(ruleId: string) { this.avatarSystem.rules = this.avatarSystem.rules.filter(r => r.id!==ruleId) },
    closeProject() {
      this.projectPath=null; this.meta=null; this.avatarSystem={default:'',rules:[]};
      this.statsSystem={stats:[...DEFAULT_STATS]}; this.days={}; this.currentDay='day1'; this.currentSceneId=null;
      this.assets={all:[],images:[],sounds:[]}
    },
    addSceneToCurrentDay(sceneId?: string, preset: Partial<Scene> = {}) {
      const newId = sceneId || `scene_${Date.now()}`
      const currentDayNum = parseInt(this.currentDay.replace('day','')) || 1
      const base: Scene = {
        Id: newId,
        SceneTitle: preset.SceneTitle || newId,
        Background: fixRelPath(preset.Background || ''),
        Text: preset.Text || '',
        Choices: preset.Choices || [],
        IsEndDay: preset.IsEndDay || false,
        Type: preset.Type || (preset.IsEndDay? 'end_of_day' : 'normal'),
        Day: preset.Day || currentDayNum,
        NextDayId: preset.NextDayId || `day${currentDayNum+1}`,
        NextDay: preset.NextDay || `day${currentDayNum+1}`,
        Transfers: preset.Transfers || { keep_flags: [], keep_stats: this.statsSystem.stats.map(s=>s.name), summary_text: '' }
      }
      const newScene: Scene = {...base,...preset, Id: newId }
      if (newScene.IsEndDay) {
        newScene.Type = 'end_of_day'
        if (!preset.NextDayId) {
          newScene.NextDay = `day${currentDayNum+1}`
          newScene.NextDayId = `day${currentDayNum+1}`
        }
        newScene.Transfers = { keep_flags: [], keep_stats: this.statsSystem.stats.map(s=>s.name), summary_text: '' }
        if (!preset.Choices?.length) {
          newScene.Choices = [{
            Text: 'Śpij',
            Next: 'END_DAY',
            NextDayId: newScene.NextDayId,
            Stats: this.createEmptyStats(),
            ReactionText: '',
            ReactionImage: '',
            SoundFile: ''
          }]
        }
      } else {
        if (!newScene.Choices?.length) newScene.Choices = []
      }
      if(!this.days[this.currentDay]) this.days[this.currentDay]=[]
      this.days[this.currentDay].push(newScene)
      this.currentSceneId=newId
      this.ensureChoiceIds(newScene)
    },
    duplicateScene(sceneId: string) {
      const scenes = this.days[this.currentDay]; if(!scenes) return;
      const toCopy = scenes.find(s => s.Id===sceneId); if(!toCopy) return;
      const copy = JSON.parse(JSON.stringify(toCopy));
      copy.Id=`${toCopy.Id}_copy_${Date.now()}`;
      copy.SceneTitle=`${toCopy.SceneTitle||sceneId} - Kopia`;
      this.ensureChoiceIds(copy);
      scenes.splice(scenes.findIndex(s=>s.Id===sceneId)+1,0,copy);
      this.currentSceneId=copy.Id
    },
    deleteScene(sceneId: string) {
      const scenes=this.days[this.currentDay];
      if(!scenes||scenes.length<=1){alert('Nie możesz usunąć ostatniej sceny');return}
      const target = scenes.find(s => s.Id===sceneId)
      if (target?.IsEndDay) {
        if (!confirm(`Usuwasz KONIEC DNIA (${sceneId}). Silnik nie będzie miał jak przejść do następnego dnia. Na pewno?`)) return
      }
      const idx=scenes.findIndex(s=>s.Id===sceneId);
      if(idx>-1){scenes.splice(idx,1); if(this.currentSceneId===sceneId) this.currentSceneId=scenes[0]?.Id||null}
    },
    updateCurrentScene(field: keyof Scene, value: any) {
      if(!this.currentScene) return;
      if(field==='Background' || field==='NextDayId' || field==='NextDay'){
        value = fixRelPath(value as string)
      }
      (this.currentScene as any)[field]=value
      if (this.currentScene.IsEndDay && (field === 'NextDayId' || field === 'NextDay')) {
        const clean = (value as string).replace('.json','').trim()
        this.currentScene.NextDayId = clean
        this.currentScene.NextDay = clean
        if (this.currentScene.Choices?.[0]?.Next === 'END_DAY' && !this.currentScene.Choices[0].NextDayId) {
          this.currentScene.Choices[0].NextDayId = clean
        }
      }
    },
    addDay(dayId: string) {
      if(this.days[dayId]){alert('Dzień już istnieje');return}
      const dayNum = parseInt(dayId.replace('day','')) || Object.keys(this.days).length+1
      const endId = `koniec_dnia_${dayNum}`
      this.days[dayId]=[
        {
          Id:`start_d${dayNum}`,
          SceneTitle:`Dzień ${dayNum} Start`,
          Background:'images/bg_tutorial.webp',
          Text:`Początek dnia ${dayNum}`,
          Choices:[{ id: (crypto as any).randomUUID?.() || Math.random().toString(36).slice(2), Text:'Dalej', Next:endId, Stats: this.createEmptyStats() }],
          Type:'normal' as const,
          Day: dayNum
        },
        {
          Id:endId,
          SceneTitle:`KONIEC DNIA ${dayNum}`,
          Background:'images/bg_tutorial.webp',
          Text:`Koniec dnia ${dayNum}. Idziesz spać.`,
          IsEndDay: true,
          Type:'end_of_day' as const,
          Day: dayNum,
          NextDayId: `day${dayNum+1}`,
          NextDay: `day${dayNum+1}`,
          Transfers: { keep_flags: [], keep_stats: this.statsSystem.stats.map(s=>s.name), summary_text: '' },
          Choices:[{ id: (crypto as any).randomUUID?.() || Math.random().toString(36).slice(2), Text:'Śpij', Next:'END_DAY', NextDayId: `day${dayNum+1}`, Stats: this.createEmptyStats(), ReactionText:'', ReactionImage:'', SoundFile:'' }]
        }
      ];
      this.currentDay=dayId; this.currentSceneId=`start_d${dayNum}`
    },
    addChoiceToCurrentScene() {
      if(!this.currentScene) return;
      if(!this.currentScene.Choices) this.currentScene.Choices=[];
      const isEnd = this.currentScene.IsEndDay
      this.currentScene.Choices.push({
        id: (crypto as any).randomUUID?.() || Math.random().toString(36).slice(2),
        Text: isEnd? 'Idź do...' : 'Nowy wybór',
        Next: isEnd? 'END_DAY' : this.currentScene.Id,
        NextDayId: isEnd? this.currentScene.NextDayId : undefined,
        Stats: this.createEmptyStats(),
        ReactionText:'',
        ReactionImage:'',
        SoundFile:''
      })
    }
  }
})
