# TODO

## Use Case
Ein Hörbuch von CD zu m4b machen

1. Dateien sammeln
    - [ ] Wenn die Dateien nicht sortiert sind (z.B. unpadded Disk numbers), dann für Sortierung sorgen.
    - [ ] Wenn die Dateien in Unterordnern sind, müssen die Ordner in der richtigen Reihenfolge verarbeitet werden
    - [ ] Soll auch einen Ordner von mehreren Hörbüchern akzeptieren
        - dann müssen alle regeln mit Regex arbeiten
        - wenn eine Regel nicht greift, muss dies am Ende geprinted werden
2. Metadaten und Kapitel anpassen
    - [ ] Metadaten sollen gecached werden
    - [ ] entweder Config-Datei für Metadaten-Änderungen, oder interaktiver flow
        - [ ] egal welcher Weg, ich möchte die Information danach im Zielordner haben, damit ich nachchecken kann
    - [ ] Der Prozess der metadata und Kapitelanpassungen muss schnell und einfach gehen
        - schneller als der bisherige
    - [ ] Disk count und disk nummer löschen
    - [ ] Chapter soll auch replace logik haben (z.B. Book I: "Bla", Chapter 1_ => I-01_)
3. Cover anpassen
3. Dateien konvertieren
    - [ ] parallelisierbar mit param
    - [ ] mit progress bar
4. File rename
    - [ ] Ordner struktur: Author/Title/Title.m4b
        - [ ] interactive?
5. Tool config file adden
    - [ ] aus den gegebenen Einstellungen eine Konfig erstellen und neben die m4b legen



## Entwurf 2
narr m4b config <dir> => generates a prepared config file
narr m4b check chapters <config> => prints chapters with rules applied
narr m4b check metadata <config> => prints metadata with rules applied
narr m4b check filename <config> => prints filename with rules applied
narr m4b check --no-rules [thing] <config> => prints metadata without rules applied
narr m4b run <config> => make an m4b with all steps
narr m4b run --no-cache <config> => make an m4b with all steps, ignoring previously executed steps
narr m4b fix chapters <project-dir> => (needs an existing project structure), applies new rules to original chapters, doing all necessary steps
narr m4b fix filename <project-dir> => (needs an existing project structure), applies new rules to original filename, doing all necessary steps
narr m4b fix metadata <project-dir> => (needs an existing project structure), applies new rules to original filename, doing all necessary steps
narr m4b fix cover <project-dir>


der vorteil von fix filename etc. auf project-dir ist, dass es wiederholbar ist


## Entwurf 1
- Wenn Config da ist und alle Required files (z. B. Cover, audio files), dann können 1-6 übersprungen werden
- Wenn nachträglich auffällt, dass es nicht gut war:
    "narr m4b fix --step metadata --step chapters <dir>"
    - [ ] Zeigt dann nur die nötigen Schritte an und mit Vorauswahl


narr m4b generate <dir>

1. Alle Dateien werden in Order gepadded aufgelistet
"Sieht das gut aus?"
n => cancel (need to code more probably)

2. Metadaten werden ausgelesen und angezeigt
"Gib mir tag, der angepasst werden soll"
"Was willst du damit tun?" Löschen, Ändern (regex), Ändern (händisch)
"Gib mir regex"
"Gib mir replace format"
"Besser so?" n => don't save

3. Kapitel werden ausgelesen und angezeigt
"Gib mir regex"
"Gib mir replace"
"Besser so?" n => don't save

4. Ist das Cover ok? y => keep it
"Leg ein besseres in dieses Verzeichnis"

5. Wie soll die Datei heißen? Fix, Regex

6. Config und Cover speichern

7. Dateien konvertieren
8. Dateien concatten
9. Metadaten anhängen
10. Kapitel anhängen
11. Datei umbenennen
12. Temp files löschen


## Config

- audio input file dir: string
- cover input file: string
- metadata rules: tagrule[]
    - regex replace (tag, match, format)
    - delete (tag)
    - set (tag, value)
- chapters: boolean
- chapter rules: rule[]
    - regex replace (match, format)
- output file name & folder structure rules: rule[]
    - regex replace (match, format)
    - set (value)

- multiproject : boolean (if true, every subdir in root is a single project)
- keep temp: boolean
- parallel: boolean
- executed-steps: step[]
