
# janusz-maker - zrób własną grę dla znajomych

Non-profitowy projekt. Każdy Janusz może zrobić własną grę.

## Struktura monorepo
- `/engine-csharp` - silnik gry C# (Twój JanuszSimulator)
- `/editor-wails` - edytor w Go/Wails (Vue)
- `/games` - przykładowe projekty gier (każdy folder = osobna gra)
- `/runtime` - eksport EXE / paczka web

## Jak odpalić
1. Skopiuj swój JanuszSimulator do /engine-csharp
2. Skopiuj janusz-editor-wails do /editor-wails
3. Uruchom edytor: `cd editor-wails && wails dev`

## Roadmap
- [ ] ETAP 1: Monorepo + rename komponentów Vue
- [ ] ETAP 2: Gry jako projekty w /games
- [ ] ETAP 3: Przycisk BUILD -> EXE / WEB (GitHub Pages)

## GitHub Pages
/hosting dla wersji web - sprawdzamy po wrzuceniu.
