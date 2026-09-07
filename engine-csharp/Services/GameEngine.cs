using Microsoft.JSInterop;
using JanuszSimulator.Models;
using System.Net.Http.Json;

namespace JanuszSimulator.Services;

public class GameEngine
{
    private readonly HttpClient _http;
    private readonly IJSRuntime _js;
    public GameState GameState { get; set; } = new();
    private CancellationTokenSource? _reactionCts;
    private bool _musicStarted = false;

    private AvatarSystem _avatarSystem = new();
    private readonly string[] _allowedPrefixes = { "av_", "bg_", "re_" };
    private string _currentDayId = "day1";

    public event Action? StateChanged;

    public GameEngine(HttpClient http, IJSRuntime js)
    {
        _http = http;
        _js = js;
    }

    public class GameDataWrapper
    {
        public List<Scene> scenes { get; set; } = new();
    }

    public class AvatarSystem
    {
        public string Default { get; set; } = "images/av_front.jpg";
        public string? FallbackDefault { get; set; }
        public List<AvatarRule> Rules { get; set; } = new();
    }
    public class AvatarRule
    {
        public string Id { get; set; } = "";
        public string Name { get; set; } = "";
        public string Use { get; set; } = "";
        public string? Fallback { get; set; }
        public Dictionary<string, Dictionary<string, int>>? If { get; set; }
        public int Priority { get; set; } = 0;
    }

    private string NormalizeStatId(string raw)
    {
        if (string.IsNullOrWhiteSpace(raw)) return raw;
        var up = raw.Trim().ToUpperInvariant();
        return up switch
        {
            "A" => "CEBULA",
            "B" => "WSTYD",
            "C" => "PORTFEL",
            "D" => "REPUTACJA",
            "E" => "REPUTACJA",
            "CEBULA" => "CEBULA",
            "WSTYD" => "WSTYD",
            "PORTFEL" => "PORTFEL",
            "REPUTACJA" => "REPUTACJA",
            _ => up
        };
    }

    private bool IsAllowedAsset(string path)
    {
        if (string.IsNullOrWhiteSpace(path)) return false;
        var file = Path.GetFileName(path).ToLowerInvariant();
        bool hasPrefix = _allowedPrefixes.Any(p => file.StartsWith(p));
        bool isAllowedExt = file.EndsWith(".jpg") || file.EndsWith(".jpeg") || file.EndsWith(".png") || file.EndsWith(".webp");
        return hasPrefix && isAllowedExt;
    }

    private string NormalizeBg(string? raw)
    {
        if (string.IsNullOrWhiteSpace(raw)) return "images/bg_dom.jpg";
        raw = raw.Trim().Replace("\\", "/");
        if (!raw.StartsWith("images/")) raw = $"images/{Path.GetFileName(raw)}";
        if (!IsAllowedAsset(raw))
        {
            Console.WriteLine($"[BG] Blokuję {raw} -> bg_dom.jpg");
            return "images/bg_dom.jpg";
        }
        return raw;
    }

    public async Task LoadGame()
    {
        // 1. Ładuj dni - nowy system day1, day2
        GameState.AllScenes = new List<Scene>();
        _currentDayId = "day1";
        try
        {
            var day1 = await _http.GetFromJsonAsync<List<Scene>>($"data/{_currentDayId}.json");
            if (day1!= null) GameState.AllScenes.AddRange(day1);
        }
        catch
        {
            // fallback stary scenes.json
            try
            {
                var data = await _http.GetFromJsonAsync<GameDataWrapper>("data/scenes.json")?? new();
                GameState.AllScenes = data.scenes;
            }
            catch { }
        }

        try
        {
            var statSystem = await _http.GetFromJsonAsync<StatsSystem>("data/_statsSystem.json");
            if (statSystem!= null && statSystem.stats.Any())
                GameState.StatDefs = statSystem.stats.Select(s => new StatDef { id = NormalizeStatId(s.id), name = NormalizeStatId(s.name), initial = s.initial }).ToList();
        }
        catch { }

        try
        {
            // FIX: edytor zapisuje _avatarSystem.json a nie avatarSystem.json
            AvatarSystem? avatarSys = null;
            try { avatarSys = await _http.GetFromJsonAsync<AvatarSystem>("data/_avatarSystem.json"); } catch { }
            if (avatarSys == null) avatarSys = await _http.GetFromJsonAsync<AvatarSystem>("data/avatarSystem.json");

            if (avatarSys!= null)
            {
                var filteredRules = avatarSys.Rules
                 .Where(r => IsAllowedAsset(r.Use))
                 .Select(r => {
                      if (r.If!= null)
                      {
                          var newIf = new Dictionary<string, Dictionary<string, int>>();
                          foreach (var kv in r.If) newIf[NormalizeStatId(kv.Key)] = kv.Value;
                          r.If = newIf;
                      }
                      return r;
                  })
                 .OrderByDescending(r => r.Priority)
                 .ToList();

                var defaultPath = IsAllowedAsset(avatarSys.Default)? avatarSys.Default : "images/av_front.jpg";
                _avatarSystem = new AvatarSystem { Default = defaultPath, FallbackDefault = avatarSys.FallbackDefault, Rules = filteredRules };
                Console.WriteLine($"[Avatar] {filteredRules.Count} reguł");
            }
        }
        catch (Exception ex)
        {
            Console.WriteLine($"[Avatar] fallback: {ex.Message}");
            _avatarSystem = new AvatarSystem { Default = "images/av_front.jpg", Rules = new() };
        }

        GameState.Stats = new DynamicStats();
        GameState.Flags = new();
        GameState.ReactionImage = "";
        GameState.ReactionText = "";

        if (GameState.StatDefs.Any())
        {
            foreach (var def in GameState.StatDefs)
            {
                var nid = NormalizeStatId(def.id);
                GameState.Stats.Values[nid] = def.initial;
            }
        }
        else
        {
            GameState.StatDefs = new List<StatDef>
            {
                new() { id = "CEBULA", name = "CEBULA", initial = 10 },
                new() { id = "WSTYD", name = "WSTYD", initial = 10 },
                new() { id = "PORTFEL", name = "PORTFEL", initial = 100 },
                new() { id = "REPUTACJA", name = "REPUTACJA", initial = 50 },
            };
            foreach (var d in GameState.StatDefs) GameState.Stats.Values[d.id] = d.initial;
        }

        var targetId = "start";
        if (!GameState.AllScenes.Any(s => s.Id == targetId))
        {
            targetId = GameState.AllScenes.FirstOrDefault(s => s.Id.StartsWith("start"))?.Id?? GameState.AllScenes.FirstOrDefault()?.Id?? "start";
        }
        Console.WriteLine($"[Game] Start sceny: {targetId} z {GameState.AllScenes.Count} scen");
        LoadScene(targetId);
    }

    public async Task LoadDay(string dayId)
    {
        if (string.IsNullOrWhiteSpace(dayId)) dayId = "day1";
        dayId = dayId.Replace(".json", "").Trim();
        Console.WriteLine($"[Game] Ładuję dzień: {dayId}");
        try
        {
            var scenes = await _http.GetFromJsonAsync<List<Scene>>($"data/{dayId}.json");
            if (scenes!= null && scenes.Any())
            {
                _currentDayId = dayId;
                GameState.AllScenes = scenes;
                var start = scenes.FirstOrDefault(s => s.Id.StartsWith("start"))?.Id?? scenes.First().Id;
                LoadScene(start);
            }
        }
        catch (Exception ex)
        {
            Console.WriteLine($"[Game] Nie udało się załadować {dayId}: {ex.Message}");
        }
    }

    public void LoadScene(string sceneId)
    {
        if (sceneId == "END_DAY")
        {
            Console.WriteLine("[Game] END_DAY jako LoadScene - ignoruję, użyj LoadDay");
            return;
        }
        GameState.CurrentScene = GameState.AllScenes.FirstOrDefault(s => s.Id == sceneId);
        if (GameState.CurrentScene == null)
        {
            Console.WriteLine($"[Game] Nie znaleziono sceny {sceneId}");
            return;
        }

        GameState.BackgroundImage = NormalizeBg(GameState.CurrentScene.Background);
        GameState.SceneTitle = GameState.CurrentScene.SceneTitle?? "";
        GameState.NarrationText = GameState.CurrentScene.Text?? "";
        GameState.CurrentChoices = GameState.CurrentScene.Choices?? new List<Choice>();
        GameState.ReactionText = "";
        GameState.ReactionImage = "";
        GameState.IsFinished = GameState.CurrentChoices.Count == 0;

        UpdateAvatar();
        NotifyStateChanged();
    }

    public bool CanSelectChoice(Choice choice)
    {
        if (choice.FlagsRequired.Any() &&!choice.FlagsRequired.All(f => GameState.Flags.Contains(f))) return false;
        foreach (var kv in choice.MinStats)
            if (GameState.Stats.Get(NormalizeStatId(kv.Key)) < kv.Value) return false;
        foreach (var kv in choice.MaxStats)
            if (GameState.Stats.Get(NormalizeStatId(kv.Key)) > kv.Value) return false;
        if (choice.KosztPortfel.HasValue && GameState.Stats.Get("PORTFEL") < choice.KosztPortfel.Value) return false;
        return true;
    }

    public IEnumerable<Choice> GetAvailableChoices()
        => GameState.CurrentChoices?.Where(c => CanSelectChoice(c))?? Enumerable.Empty<Choice>();

    public async Task MakeChoice(Choice choice)
    {
        if (!CanSelectChoice(choice))
        {
            GameState.ReactionText = string.IsNullOrWhiteSpace(choice.FailText)? "Nie możesz tego zrobić, Janusz." : choice.FailText;
            GameState.ReactionImage = "images/av_back.jpg";
            NotifyStateChanged();
            await Task.Delay(5000);
            GameState.ReactionText = "";
            GameState.ReactionImage = "";
            NotifyStateChanged();
            return;
        }

        if (!_musicStarted)
        {
            _musicStarted = true;
            try { await _js.InvokeVoidAsync("JanuszAudio.startMusic", "sounds/s1.mp3"); } catch {}
        }

        _reactionCts?.Cancel();
        ApplyChoiceEffects(choice);

        GameState.ReactionText = string.IsNullOrWhiteSpace(choice.ReactionText)? "Janusz coś kombinuje..." : choice.ReactionText;
        var reactionImgRaw = choice.ReactionImage;
        if (string.IsNullOrWhiteSpace(reactionImgRaw)) reactionImgRaw = "images/av_front.jpg";
        else if (!IsAllowedAsset(reactionImgRaw)) reactionImgRaw = "images/av_front.jpg";
        GameState.ReactionImage = reactionImgRaw;
        NotifyStateChanged();

        if (!string.IsNullOrWhiteSpace(choice.SoundFile))
            _ = _js.InvokeVoidAsync("JanuszAudio.playVoice", $"sounds/{choice.SoundFile}");

        if (!string.IsNullOrWhiteSpace(GameState.ReactionText) ||!string.IsNullOrWhiteSpace(GameState.ReactionImage))
        {
            _reactionCts = new CancellationTokenSource();
            try { await Task.Delay(5000, _reactionCts.Token); } catch (TaskCanceledException) { } finally { HideReaction(); }
        }

        // FIX KLUCZOWY: END_DAY
        if (!string.IsNullOrWhiteSpace(choice.Next))
        {
            if (choice.Next == "END_DAY")
            {
                var endScene = GameState.CurrentScene;
                var nextDay = endScene?.NextDayId?? endScene?.NextDay?? "day2";
                Console.WriteLine($"[Game] END_DAY -> ładuję {nextDay}");
                await LoadDay(nextDay);
                return;
            }
            LoadScene(choice.Next);
        }
    }

    private void ApplyChoiceEffects(Choice choice)
    {
        foreach (var kv in choice.Stats)
        {
            var key = NormalizeStatId(kv.Key);
            var cur = GameState.Stats.Get(key);
            GameState.Stats.Set(key, cur + kv.Value);
        }
        if (choice.Cebula!= 0) GameState.Stats.Set("CEBULA", GameState.Stats.Get("CEBULA") + choice.Cebula);
        if (choice.Wstyd!= 0) GameState.Stats.Set("WSTYD", GameState.Stats.Get("WSTYD") + choice.Wstyd);
        if (choice.Portfel!= 0) GameState.Stats.Set("PORTFEL", GameState.Stats.Get("PORTFEL") + choice.Portfel);
        if (choice.Reputacja!= 0) GameState.Stats.Set("REPUTACJA", GameState.Stats.Get("REPUTACJA") + choice.Reputacja);
        if (choice.KosztPortfel.HasValue)
            GameState.Stats.Set("PORTFEL", GameState.Stats.Get("PORTFEL") - choice.KosztPortfel.Value);

        foreach (var flag in choice.FlagsSet)
            GameState.Flags.Add(flag);

        foreach (var key in GameState.Stats.Values.Keys.ToList())
        {
            var v = GameState.Stats.Values[key];
            if (key == "CEBULA" || key == "WSTYD" || key == "REPUTACJA")
                GameState.Stats.Values[key] = Math.Clamp(v, 0, 100);
            else if (key == "PORTFEL")
                GameState.Stats.Values[key] = Math.Max(0, v);
        }
    }

    public void SkipReaction() => _reactionCts?.Cancel();
    public void HideReaction() { GameState.ReactionImage = ""; GameState.ReactionText = ""; NotifyStateChanged(); }

    private void UpdateAvatar()
    {
        foreach (var rule in _avatarSystem.Rules.OrderByDescending(r => r.Priority))
        {
            if (rule.If == null || rule.If.Count == 0) continue;
            bool matches = true;
            foreach (var kv in rule.If)
            {
                var currentValue = GameState.Stats.Get(NormalizeStatId(kv.Key));
                foreach (var cond in kv.Value)
                {
                    if (cond.Key == "gte" && currentValue < cond.Value) matches = false;
                    if (cond.Key == "lte" && currentValue > cond.Value) matches = false;
                    if (cond.Key == "gt" && currentValue <= cond.Value) matches = false;
                    if (cond.Key == "lt" && currentValue >= cond.Value) matches = false;
                    if (cond.Key == "eq" && currentValue!= cond.Value) matches = false;
                }
            }
            if (matches && IsAllowedAsset(rule.Use)) { GameState.JanuszImage = rule.Use; return; }
        }
        GameState.JanuszImage = _avatarSystem.Default;
    }

    private void NotifyStateChanged() => StateChanged?.Invoke();
}