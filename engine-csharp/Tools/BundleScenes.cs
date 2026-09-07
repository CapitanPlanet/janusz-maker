using System.Text;
using System.Text.Json;
using System.Text.Json.Nodes;

namespace JanuszTools;

public class BundleScenes
{
    public static int Main(string[] args)
    {
        try
        {
            var projectDir = args.Length > 0? args[0] : Directory.GetCurrentDirectory();
            Console.WriteLine($"[BUNDLER] ProjectDir: {projectDir}");

            var assetsDir = Path.Combine(projectDir, "Assets", "Data");
            var outDir = Path.Combine(projectDir, "wwwroot", "data");

            if (!Directory.Exists(assetsDir))
            {
                Console.WriteLine($"[BUNDLER ERROR] Brak folderu: {assetsDir}");
                return 1;
            }

            Directory.CreateDirectory(outDir);

            // 1. KOPIUJ _statsSystem.json 1:1 osobno
            var statsSrc = Path.Combine(assetsDir, "_statsSystem.json");
            var statsDst = Path.Combine(outDir, "_statsSystem.json");
            if (File.Exists(statsSrc))
            {
                File.Copy(statsSrc, statsDst, true);
                Console.WriteLine($"[BUNDLER] _statsSystem.json copied");
            }

            // 2. SCENY - bez żadnej transformacji, kopiuj 1:1
            var all = new JsonArray();
            var files = Directory.GetFiles(assetsDir, "day*.json").OrderBy(f => f).ToArray();

            if (files.Length == 0)
            {
                Console.WriteLine($"[BUNDLER ERROR] Brak plikow day*.json w {assetsDir}");
                return 1;
            }

            foreach (var file in files)
            {
                var filename = Path.GetFileName(file);
                var json = File.ReadAllText(file, Encoding.UTF8);
                var node = JsonNode.Parse(json);
                if (node == null) continue;

                if (node is JsonObject obj && obj["scenes"] is JsonArray scenes)
                {
                    foreach (var s in scenes) if (s!= null) all.Add(s.DeepClone());
                    Console.WriteLine($"[BUNDLER] {filename}: {scenes.Count} scen");
                }
                else if (node is JsonArray arr)
                {
                    foreach (var s in arr) if (s!= null) all.Add(s.DeepClone());
                    Console.WriteLine($"[BUNDLER] {filename}: {arr.Count} scen");
                }
            }

            var result = new JsonObject { ["scenes"] = all };
            var resultJson = result.ToJsonString(new JsonSerializerOptions
            {
                WriteIndented = true,
                Encoder = System.Text.Encodings.Web.JavaScriptEncoder.UnsafeRelaxedJsonEscaping
            });

            var outPath = Path.Combine(outDir, "scenes.json");
            File.WriteAllText(outPath, resultJson, new UTF8Encoding(false));
            Console.WriteLine($"[BUNDLER OK] Zapisano {all.Count} scen z {files.Length} plikow -> {outPath}");
            return 0;
        }
        catch (Exception ex)
        {
            Console.WriteLine($"[BUNDLER CRASH] {ex.GetType().Name}: {ex.Message}");
            Console.WriteLine(ex.StackTrace);
            return 1;
        }
    }
}