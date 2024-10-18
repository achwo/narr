# TODO

## Use Case
Input: ein haufen flacs
-> to alac für raw_audio
-> to m4b für lib

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



## Entwurf 
narr m4b config <dir> => generates a prepared config file
narr m4b check chapters <config> => prints chapters with rules applied
narr m4b check metadata <config> => prints metadata with rules applied
narr m4b check filename <config> => prints filename with rules applied
narr m4b check --no-rules [thing] <config> => prints metadata without rules applied
narr m4b run <config> => make an m4b with all steps
narr m4b run --no-cache <config> => make an m4b with all steps, ignoring previously executed steps

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
