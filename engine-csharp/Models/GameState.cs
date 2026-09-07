namespace JanuszSimulator.Models;

public class StatDef
{
    public string id { get; set; } = "";
    public string name { get; set; } = "";
    public int initial { get; set; }
}

public class StatsSystem
{
    public List<StatDef> stats { get; set; } = new();
}

public class Transfers
{
    public List<string> keep_flags { get; set; } = new();
    public List<string> keep_stats { get; set; } = new();
    public string? summary_text { get; set; }
    public string? nextDay { get; set; }
}

public class DynamicStats
{
    public Dictionary<string, int> Values { get; set; } = new();

    public int Get(string id)
    {
        if (string.IsNullOrWhiteSpace(id)) return 0;
        // normalizacja legacy
        var up = id.Trim().ToUpperInvariant();
        var key = up switch { "A" => "CEBULA", "B" => "WSTYD", "C" => "PORTFEL", "D" => "REPUTACJA", "E" => "REPUTACJA", _ => up };
        return Values.TryGetValue(key, out var v)? v : 0;
    }
    public void Set(string id, int value)
    {
        if (string.IsNullOrWhiteSpace(id)) return;
        var up = id.Trim().ToUpperInvariant();
        var key = up switch { "A" => "CEBULA", "B" => "WSTYD", "C" => "PORTFEL", "D" => "REPUTACJA", "E" => "REPUTACJA", _ => up };
        Values[key] = value;
    }

    // BACKWARD COMPAT - teraz mapuje na nowe klucze
    public int Cebula { get => Get("CEBULA"); set => Set("CEBULA", value); }
    public int Wstyd { get => Get("WSTYD"); set => Set("WSTYD", value); }
    public int Portfel { get => Get("PORTFEL"); set => Set("PORTFEL", value); }
    public int Reputacja { get => Get("REPUTACJA"); set => Set("REPUTACJA", value); }
}

public class PlayerStats : DynamicStats {}
public class Stats : DynamicStats {}

public class Scene
{
    public string Id { get; set; } = "";
    public string Background { get; set; } = "";
    public string Text { get; set; } = "";
    public List<Choice> Choices { get; set; } = new();
    public bool IsIntermission { get; set; } = false;
    public string IntermissionNext { get; set; } = "";
    public string ReactionText { get; set; } = "";
    public string ReactionImage { get; set; } = "";
    public string SceneTitle { get; set; } = "";

    // NOWE - koniec dnia
    public bool IsEndDay { get; set; } = false;
    public string? Type { get; set; }
    public int Day { get; set; } = 1;
    public string? NextDayId { get; set; }
    public string? NextDay { get; set; }
    public Transfers? Transfers { get; set; }
}

public class Choice
{
    public string Text { get; set; } = "";
    public string Next { get; set; } = "";
    public Dictionary<string, int> Stats { get; set; } = new();
    public int Cebula { get; set; }
    public int Wstyd { get; set; }
    public int Portfel { get; set; }
    public int Reputacja { get; set; }
    public string ReactionText { get; set; } = "";
    public string ReactionImage { get; set; } = "";
    public string SoundFile { get; set; } = "";
    public List<string> FlagsSet { get; set; } = new();
    public List<string> FlagsRequired { get; set; } = new();
    public Dictionary<string, int> MinStats { get; set; } = new();
    public Dictionary<string, int> MaxStats { get; set; } = new();
    public int? MinCebula { get; set; }
    public int? MaxCebula { get; set; }
    public int? MinWstyd { get; set; }
    public int? MaxWstyd { get; set; }
    public int? MinPortfel { get; set; }
    public int? MaxPortfel { get; set; }
    public int? MinReputacja { get; set; }
    public int? MaxReputacja { get; set; }
    public int? KosztPortfel { get; set; }
    public string FailText { get; set; } = "Janusz, nie stac cie na to.";
}

public class GameState
{
    public DynamicStats Stats { get; set; } = new();
    public List<StatDef> StatDefs { get; set; } = new();
    public HashSet<string> Flags { get; set; } = new();
    public List<Scene> AllScenes { get; set; } = new();
    public Scene? CurrentScene { get; set; }
    public string BackgroundImage { get; set; } = "";
    public string SceneTitle { get; set; } = "";
    public string NarrationText { get; set; } = "";
    public List<Choice> CurrentChoices { get; set; } = new();
    public string ReactionText { get; set; } = "";
    public string ReactionImage { get; set; } = "";
    public string JanuszImage { get; set; } = "";
    public bool IsFinished { get; set; }
}