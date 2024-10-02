package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
)

type Song struct {
	GroupName   string
	SongName    string
	ReleaseDate string
	Text        template.HTML
	Link        string
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	action := r.FormValue("action")
	songTitle := r.FormValue("nazvanie")

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbSslMode := os.Getenv("DB_SSLMODE")

	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", dbUser, dbPassword, dbName, dbSslMode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Println("Ошибка подключения к базе данных:", err)
		http.Error(w, "Ошибка подключения к базе данных", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var songs []Song

	log.Printf("Получено действие: %s для песни: %s", action, songTitle)

	switch action {
	case "GetText":
		log.Println("Запрос текста песни")
		rows, err := db.Query("SELECT group_name, song_title, release_date, text, link FROM songs WHERE song_title = $1", songTitle)
		if err != nil {
			log.Println("Ошибка запроса на получение текста песни:", err)
			http.Error(w, "Ошибка получения текста песни", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var song Song
			if err := rows.Scan(&song.GroupName, &song.SongName, &song.ReleaseDate, &song.Text, &song.Link); err != nil {
				log.Println("Ошибка сканирования строки:", err)
				http.Error(w, "Ошибка обработки данных песни", http.StatusInternalServerError)
				return
			}
			a := song.Text
			song.Text = template.HTML(strings.ReplaceAll(string(a), "\n", "<br>"))
			songs = append(songs, song)
		}
	case "GetData":
		log.Println("Запрос всех песен")
		rows, err := db.Query("SELECT group_name, song_title, release_date, text, link FROM songs")
		if err != nil {
			log.Println("Ошибка запроса на получение всех песен:", err)
			http.Error(w, "Ошибка получения данных о песнях", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var song Song
			if err := rows.Scan(&song.GroupName, &song.SongName, &song.ReleaseDate, &song.Text, &song.Link); err != nil {
				log.Println("Ошибка сканирования строки:", err)
				http.Error(w, "Ошибка обработки данных о песнях", http.StatusInternalServerError)
				return
			}
			a := song.Text
			song.Text = template.HTML(strings.ReplaceAll(string(a), "\n", "<br>"))
			songs = append(songs, song)
		}
	case "DeleteSong":
		log.Printf("Запрос на удаление песни: %s", songTitle)
		if _, err = db.Exec("DELETE FROM songs WHERE song_title = $1", songTitle); err != nil {
			log.Println("Ошибка удаления песни:", err)
			http.Error(w, "Ошибка при удалении песни", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Песня '%s' успешно удалена.", songTitle)
	    return 
	case "AddSong":
		groupName := r.FormValue("group_name")
		releaseDate := r.FormValue("release_date")
		text := r.FormValue("text")
		link := r.FormValue("link")

		text = strings.ReplaceAll(text, "\n", "<br>")

        query := `INSERT INTO songs (group_name, song_title, release_date, text, link) VALUES ($1, $2, $3, $4, $5)`
		
        log.Printf("Добавление новой песни: %s от группы: %s", songTitle, groupName)

        if _, err := db.Exec(query, groupName, songTitle, releaseDate, text, link); err != nil {
            log.Println("Ошибка вставки новой песни:", err)
            http.Error(w, "Ошибка при добавлении новой песни", http.StatusInternalServerError)
            return
        }

        fmt.Fprintf(w, "Песня '%s' успешно добавлена.", songTitle)
        return 
	case "UpdateSong":
	    newGroupName := r.FormValue("group_name")
	    newReleaseDate := r.FormValue("release_date")
	    newText := r.FormValue("text")
	    newLink := r.FormValue("link")

	    log.Printf("Обновление данных о песне: %s", songTitle)

	    query := `UPDATE songs SET group_name = $1, release_date = $2, text = $3, link = $4 WHERE song_title = $5`
	    if _, err := db.Exec(query, newGroupName, newReleaseDate, newText, newLink, songTitle); err != nil {
	    	log.Println("Ошибка обновления данных о песне:", err)
	    	http.Error(w, "Ошибка при обновлении данных о песне", http.StatusInternalServerError)
	    	return
	    }

	    fmt.Fprintf(w, "Данные о песне '%s' успешно обновлены.", songTitle)
	    return 
	default:
	    log.Println("Неизвестное действие:", action)
	    fmt.Fprintf(w, "Неизвестное действие")
	    return
    }

	tmpl := template.Must(template.ParseFiles("index.html"))
	if err := tmpl.Execute(w, songs); err != nil {
        log.Println("Ошибка выполнения шаблона:", err)
    }
}

func Show(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html")) 
    _ = tmpl.Execute(w, []Song{})
}

func main() {
    if err := godotenv.Load(); err != nil {
        log.Fatal("Ошибка загрузки .env файла:", err)
    }

	http.HandleFunc("/", Show)   
	http.HandleFunc("/handle", handleRequest) 
    
	port := os.Getenv("PORT")

	log.Printf("Сервер работает на порту :%s", port)
    if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
        log.Fatal("Ошибка запуска сервера:", err)
    }
}
