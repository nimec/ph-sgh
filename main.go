package main

import (
	"fmt"
	"html/template"
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
	// English comment: Populating data from the provided images
	data := MenuData{
		Pizzas: []Product{
			{ID: 1, Name: "Pizza mit Tomaten und Käse", Description: "Klassisch mit Tomatensauce und Käse", Price: 7.50},
			{ID: 2, Name: "Pizza mit Paprika", Description: "Frische Paprika, Tomaten, Käse", Price: 8.00},
			{ID: 4, Name: "Pizza mit Salami", Description: "Würzige Salami, Tomaten, Käse", Price: 8.50},
			{ID: 9, Name: "Pizza mit Thunfisch", Description: "Thunfisch, Zwiebeln, Tomaten, Käse", Price: 9.00},
			{ID: 11, Name: "Pizza mit Schinken und Ananas", Description: "Hawaiian Style mit Schinken и Ananas", Price: 9.50},
		},
		Pastas: []Product{
			{ID: 25, Name: "Spaghetti mit Tomatensauce", Description: "Klassische italienische Tomatensauce", Price: 6.50},
			{ID: 26, Name: "Spaghetti Bolognese", Description: "Mit herzhafter Fleischsauce", Price: 7.00},
		},
		Salats: []Product{
			{ID: 22, Name: "Griechischer Bauernsalat", Description: "Eisbergsalat, Tomaten, Gurken, Paprika, Schafskäse, Oliven", Price: 7.00},
		},
	}

	tmpl, err := template.ParseFiles("templates/menu.html")
	if err != nil {
		log.Printf("Error: %v", err)
		http.Error(w, "Internal Error", 500)
		return
	}
	tmpl.Execute(w, data)
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

// Product represents a menu item
type Product struct {
	ID          int
	Name        string
	Description string
	Price       float64
}

// MenuData holds all categories for the template
type MenuData struct {
	Pizzas []Product
	Pastas []Product
	Salats []Product
}
