+# Pizzeria Sangerhausen - Website in Go

Eine Pizzeria-Website, geschrieben in Go mit einem eingebauten HTTP-Server.

## Erste Schritte

### Anforderungen
- Go 1.21 oder höher

### Installation und Ausführung

```bash
# Repository klonen
git clone https://github.com/nimec/ph-sgh.git
cd ph-sgh

# Anwendung starten
go run main.go
```

Der Server ist verfügbar unter: `http://localhost:8080`

## Projektstruktur

```
ph-sgh/
├── main.go              # Hauptdatei der Anwendung
├── go.mod              # Go-Modul
├── README.md           # Diese Datei
└── static/             # Statische Dateien (HTML, CSS, JS, Bilder)
    ├── index.html
    ├── menu.html
    ├── css/
    └── js/
```

## Funktionen

- HTML-Seiten anzeigen
- API für Menü
- API für Bestellungen

## Entwicklung

Kopieren Sie Ihre HTML-Dateien aus Google Docs in den `static/` Ordner.
