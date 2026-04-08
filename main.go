package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	// Обслуживание статических файлов (HTML, CSS, JS, изображения)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	// Дополнительный обработчик для главной страницы
	http.HandleFunc("/api/menu", handleMenu)
	http.HandleFunc("/api/order", handleOrder)

	port := ":8080"
	fmt.Printf("🍕 Сервер пиццерии запущен на http://localhost%s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}

func handleMenu(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"menu": "Меню пиццерии"}`)
}

func handleOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"status": "Заказ принят"}`)
}
