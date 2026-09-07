namespace JanuszSimulator.Services;

public class AssetWhitelistService
{
    private readonly string[] _prefixes = { "av_", "bg_", "re_" };
    private readonly string[] _ext = { ".jpg", ".webp" };

    public AssetWhitelistService(string[] prefixes) => _prefixes = prefixes;

    public bool IsAllowed(string path)
    {
        var file = Path.GetFileName(path).ToLowerInvariant();
        return _prefixes.Any(p => file.StartsWith(p)) && _ext.Any(e => file.EndsWith(e));
    }

    // filtruje json z avatarami
    public AvatarSystem FilterAvatarSystem(AvatarSystem sys)
    {
        sys.Rules = sys.Rules.Where(r => IsAllowed(r.Use)).ToList();
        if (!IsAllowed(sys.Default)) sys.Default = "images/av_front.jpg";
        return sys;
    }
}

public class AvatarSystem
{
    public string Default { get; set; } = "images/av_front.jpg";
    public List<AvatarRule> Rules { get; set; } = new();
}
public class AvatarRule
{
    public string Id { get; set; } = "";
    public string Name { get; set; } = ""; // tego Ci brakowało
    public string Use { get; set; } = "";
    public Dictionary<string, Dictionary<string, int>> If { get; set; } = new();
    public int Priority { get; set; }
}