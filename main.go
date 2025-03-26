package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Task struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	Priority  string    `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
}

var DB *gorm.DB

func initDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("taskhub.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("⚠️ Не удалось подключиться к базе данных:", err)
	}
	DB.AutoMigrate(&Task{})
}

func createTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}
	task.ID = uuid.New().String()
	task.CreatedAt = time.Now()
	if task.Status == "" {
		task.Status = "new"
	}
	if err := DB.Create(&task).Error; err != nil {
		http.Error(w, "⚠️ Ошибка создания задачи", http.StatusInternalServerError)
		return
	}
	log.Println("✅ Создана новая задача:", task.Title)
	json.NewEncoder(w).Encode(task)
}

func getTasks(w http.ResponseWriter, r *http.Request) {
	var tasks []Task
	DB.Find(&tasks)
	json.NewEncoder(w).Encode(tasks)
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "⚠️ ID не указан", http.StatusBadRequest)
		return
	}
	var task Task
	if err := DB.First(&task, "id = ?", id).Error; err != nil {
		http.Error(w, "⚠️ Задача не найдена", http.StatusNotFound)
		return
	}
	var updatedTask Task
	json.NewDecoder(r.Body).Decode(&updatedTask)
	if updatedTask.Status != "" {
		task.Status = updatedTask.Status
	}
	if updatedTask.Priority != "" {
		task.Priority = updatedTask.Priority
	}
	DB.Save(&task)
	log.Println("✏️ Обновлена задача:", id)
	json.NewEncoder(w).Encode(task)
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "⚠️ ID не указан", http.StatusBadRequest)
		return
	}
	DB.Delete(&Task{}, "id = ?", id)
	log.Println("🗑️ Удалена задача:", id)
	w.WriteHeader(http.StatusNoContent)
}

func logRequest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("📥 %s %s %s", r.Method, r.URL.Path, time.Since(start))
		next(w, r)
	}
}

func main() {
	initDB()
	mux := http.NewServeMux()

	mux.HandleFunc("/tasks", logRequest(getTasks))
	mux.HandleFunc("/task", logRequest(createTask))
	mux.HandleFunc("/task/update", logRequest(updateTask))
	mux.HandleFunc("/task/delete", logRequest(deleteTask))

	port := 8080
	log.Printf("🚀 Сервер запущен на порту %d", port)
	http.ListenAndServe(":"+strconv.Itoa(port), mux)
}




