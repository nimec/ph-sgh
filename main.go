package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/glebarez/sqlite" // English comment: Pure Go SQLite driver (no CGO needed)
	"gorm.io/gorm"
)

func main() {
	initDB() // English comment: Setup database before starting the server

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	// Page handlers
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/menu", handleMenu)
	http.HandleFunc("/about", handleAbout)
	http.HandleFunc("/contact", handleContact)
	http.HandleFunc("/admin/delete", handleDelete)
	http.HandleFunc("/admin/save", handleSave)
	http.HandleFunc("/admin", handleAdmin)

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
	var pizzas, pastas []Product

	// English comment: Fetch items from DB filtered by category
	db.Where("category = ?", "Pizza").Find(&pizzas)
	db.Where("category = ?", "Pasta").Find(&pastas)

	data := struct {
		Pizzas []Product
		Pastas []Product
	}{
		Pizzas: pizzas,
		Pastas: pastas,
	}

	tmpl, err := template.ParseFiles("templates/menu.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
func handleAdmin(w http.ResponseWriter, r *http.Request) {
	var products []Product
	// English comment: Fetch all products from DB to display in the admin table
	db.Find(&products)

	tmpl, err := template.ParseFiles("templates/admin.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// English comment: Pass the list of products to the template
	tmpl.Execute(w, products)
}
func handleDelete(w http.ResponseWriter, r *http.Request) {
	// English comment: Get ID from URL query, e.g., /admin/delete?id=10
	id := r.URL.Query().Get("id")

	if id != "" {
		// English comment: GORM perform a soft delete (or hard delete if no DeletedAt field)
		db.Delete(&Product{}, id)
	}

	// English comment: Redirect back to admin panel
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// handleSave processes the form submission to create a new product
func handleSave(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests for data submission
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	// Parse form values from the request body
	// Equivalent to @RequestParam or request.getParameter() in Java
	name := r.FormValue("name")
	description := r.FormValue("description")
	category := r.FormValue("category")
	menu_number := r.FormValue("menu_number")

	// Convert the price string to a float64
	// In Go, we must explicitly handle the conversion and potential error
	price, err := strconv.ParseFloat(r.FormValue("price"), 64)
	if err != nil {
		log.Printf("Invalid price input: %v", err)
		price = 0.0
	}

	// Create a new Product instance with the form data
	newProduct := Product{
		Name:        name,
		MenuNumber:  menu_number,
		Description: description,
		Price:       price,
		Category:    category,
	}

	// Persist the new product to the database using GORM
	// Similar to repository.save(entity) in Spring Data JPA
	result := db.Create(&newProduct)
	if result.Error != nil {
		log.Printf("Database error: %v", result.Error)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Redirect back to the admin dashboard after successful save
	// This follows the Post/Redirect/Get pattern
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

var db *gorm.DB

func initDB() {
	var err error
	// English comment: Open connection to SQLite database file
	db, err = gorm.Open(sqlite.Open("pizza.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// English comment: Automatically create the table based on Product struct
	db.AutoMigrate(&Product{})

	// English comment: Seed data if table is empty
	var count int64
	db.Model(&Product{}).Count(&count)
	if count == 0 {
		initialProducts := []Product{
			{Name: "Pizza mit Tomaten und Käse", Description: "Grundlage für alle Pizzen", Price: 7.50, Category: "Pizza"},
			{Name: "Pizza mit Paprika", Description: "Frische Paprika", Price: 8.00, Category: "Pizza"},
			{Name: "Spaghetti Bolognese", Description: "Hausgemachte Fleischsauce", Price: 7.00, Category: "Pasta"},
		}
		db.Create(&initialProducts)
		log.Println("Database seeded with initial items.")
	}
}

// Product represents a menu item
type Product struct {
	gorm.Model
	ID          int
	Name        string
	MenuNumber  string // English comment: Store numbers like "01" or "02a"
	Description string
	Price       float64
	Category    string // English comment: To distinguish between Pizza, Pasta, etc.
}

// MenuData holds all categories for the template
type MenuData struct {
	Pizzas []Product
	Pastas []Product
	Salats []Product
}
