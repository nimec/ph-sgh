package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/glebarez/sqlite" // Pure Go SQLite driver (no CGO needed)
	"gorm.io/gorm"
)

func main() {
	initDB() // Setup the database before starting the server

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Page handlers
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/menu", handleMenu)
	http.HandleFunc("/about", handleAbout)

	http.HandleFunc("/admin/delete", requireAdmin(handleDelete))
	http.HandleFunc("/admin/save", requireAdmin(handleSave))
	http.HandleFunc("/admin/update-image", requireAdmin(handleUpdateImage)) // <-- ADD THIS LINE
	http.HandleFunc("/admin", requireAdmin(handleAdmin))

	// Serve static assets (CSS, JS, images)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("static/assets"))))

	// Listen on all network interfaces for cloud deployment compatibility (e.g., Fly.io)
	port := "0.0.0.0:8080"
	fmt.Printf("🍕 Pizzeria-Server is running on http://localhost%s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}

var (
	adminUser = os.Getenv("ADMIN_USER")
	adminPass = os.Getenv("ADMIN_PASS")
)

// requireAdmin is a middleware that enforces Basic Authentication for admin routes
func requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != adminUser || pass != adminPass {
			w.Header().Set("WWW-Authenticate", `Basic realm="Admin area"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// serveHTML reads and serves an HTML file with the proper content type
func serveHTML(w http.ResponseWriter, filename string) {
	content, err := os.ReadFile("static/" + filename)
	if err != nil {
		http.Error(w, "Page not found", http.StatusNotFound)
		log.Printf("Error loading %s: %v", filename, err)
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
	var pizzas, pastas, extras, drinks, salats, sonstige []Product // Add sonstige variable

	// Fetch items from the database filtered by category
	db.Where("category = ?", "Pizza").Find(&pizzas)
	db.Where("category = ?", "Pasta").Find(&pastas)
	db.Where("category = ?", "Extra").Find(&extras)
	db.Where("category = ?", "Getränke").Find(&drinks)   // Fetch drinks
	db.Where("category = ?", "Salat").Find(&salats)      // Fetch salats
	db.Where("category = ?", "Sonstige").Find(&sonstige) // Fetch sonstige

	// Add Extras to the data structure that gets sent to the HTML
	data := struct {
		Pizzas   []Product
		Pastas   []Product
		Extras   []Product
		Drinks   []Product
		Salats   []Product
		Sonstige []Product // Add sonstige field
	}{
		Pizzas:   pizzas,
		Pastas:   pastas,
		Extras:   extras,
		Drinks:   drinks,
		Salats:   salats,
		Sonstige: sonstige, // Pass the fetched sonstige
	}

	tmpl, err := template.ParseFiles("templates/menu.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

// handleUpdateImage processes an image upload for an existing product
func handleUpdateImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	// Limit upload size to 10 MB
	r.ParseMultipartForm(10 << 20)

	id := r.FormValue("id")
	if id == "" {
		http.Redirect(w, r, "/admin", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		log.Printf("Error retrieving file: %v", err)
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	defer file.Close()

	// Save the file
	os.MkdirAll("static/images", 0755)
	filename := filepath.Base(header.Filename)
	dstPath := filepath.Join("static/images", filename)

	dst, err := os.Create(dstPath)
	if err != nil {
		log.Printf("Error saving file: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	io.Copy(dst, file)

	// Find the product in the DB and update its Image column
	var product Product
	if err := db.First(&product, id).Error; err == nil {
		db.Model(&product).Update("image", "images/"+filename)
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
func handleAbout(w http.ResponseWriter, r *http.Request) {
	serveHTML(w, "about.html")
}

func handleAdmin(w http.ResponseWriter, r *http.Request) {
	var products []Product

	// Fetch all products from the database to display in the admin table
	db.Find(&products)

	tmpl, err := template.ParseFiles("templates/admin.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Pass the list of products to the template
	tmpl.Execute(w, products)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	// Get the ID from the URL query (e.g., /admin/delete?id=10)
	id := r.URL.Query().Get("id")

	if id != "" {
		// Perform a soft delete (or hard delete if there is no DeletedAt field)
		db.Delete(&Product{}, id)
	}

	// Redirect back to the admin panel
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
	name := r.FormValue("name")
	description := r.FormValue("description")
	category := r.FormValue("category")
	menu_number := r.FormValue("menu_number")

	// Convert the price string to a float64
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

	// Parse multipart form to handle file uploads (limit to 10 MB)
	r.ParseMultipartForm(10 << 20)

	// Handle the image file upload
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		os.MkdirAll("static/images", 0755)
		filename := filepath.Base(header.Filename)
		dstPath := filepath.Join("static/images", filename)
		dst, _ := os.Create(dstPath)
		defer dst.Close()
		io.Copy(dst, file)
		newProduct.Image = "images/" + filename
	}

	// Persist the new product to the database
	result := db.Create(&newProduct)
	if result.Error != nil {
		log.Printf("Database error: %v", result.Error)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Redirect back to the admin dashboard after a successful save
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

var db *gorm.DB

func initDB() {
	var err error

	// Create the /data directory (required for Fly.io persistent volumes)
	os.MkdirAll("/data", 0755)

	// Open connection to the SQLite database file inside the persistent volume
	db, err = gorm.Open(sqlite.Open("/data/pizza.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Automatically create the table based on the Product struct
	db.Migrator().DropTable(&Product{})
	db.AutoMigrate(&Product{})

	// Seed data if the table is empty
	var count int64
	db.Model(&Product{}).Count(&count)
	if count == 0 {
		loadMenuFromJSON("menu.json")
		log.Println("Database seeded from menu.json.")
		loadMenuFromJSON("extras.json")
		log.Println("Database seeded from extras.json.")
	}
}

// loadMenuFromJSON reads the menu.json file and inserts all items into the database
func loadMenuFromJSON(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("Error reading %s: %v", filename, err)
		return
	}

	var products []Product
	err = json.Unmarshal(data, &products)
	if err != nil {
		log.Printf("Error parsing JSON: %v", err)
		return
	}

	// Save each product to the database (will insert or update)
	for _, product := range products {
		result := db.Save(&product)
		if result.Error != nil {
			log.Printf("Error saving product %s from %s: %v", product.Name, filename, result.Error)
			// Decide if you want to continue or return on error
		}
	}
	log.Printf("Successfully loaded %d products from %s", len(products), filename)
}

// Product represents a menu item
type Product struct {
	gorm.Model
	ID          int
	Name        string
	MenuNumber  string `json:"menu_number"` // Store numbers like "01" or "02a"
	Description string
	Price       float64
	Category    string // Used to distinguish between Pizza, Pasta, etc.
	Image       string
}

// MenuData holds all categories for the template
type MenuData struct {
	Pizzas []Product
	Pastas []Product
	Salats []Product
}
