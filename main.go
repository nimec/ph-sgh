package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Page handlers
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/menu", handleMenu)
	http.HandleFunc("/about", handleAbout)
	http.HandleFunc("/contact", handleContact)

	// API handlers
	http.HandleFunc("/api/order", handleOrder)

	// Serve static assets (CSS, JS, images)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("static/assets"))))

	port := ":8080"
	fmt.Printf("🍕 Pizzeria-Server läuft auf http://localhost%s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}

// serveHTML reads and serves HTML file with proper content type
func serveHTML(w http.ResponseWriter, filename string) {
	content, err := os.ReadFile("static/" + filename)
	if err != nil {
		http.Error(w, "Seite nicht gefunden", http.StatusNotFound)
		log.Printf("Fehler beim Laden von %s: %v", filename, err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(content)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	serveHTML(w, "index.html")
}

func handleMenu(w http.ResponseWriter, r *http.Request) {
	serveHTML(w, "menu.html")
}

func handleAbout(w http.ResponseWriter, r *http.Request) {
	serveHTML(w, "about.html")
}

func handleContact(w http.ResponseWriter, r *http.Request) {
	serveHTML(w, "contact.html")
}

func handleOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"status": "Bestellung angenommen"}`)
}
