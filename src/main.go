package main

import (
    "html/template"
    "io/fs"
    "log"
    "net/http"
    "path/filepath"
    "sync"
)

type PageData struct {
    Videos     []VideoInfo
    VisitCount int
}

type VideoInfo struct {
    Path      string
    Title     string
    PlayCount int // Add a field for play count
}

var (
    visitCount int
    playCount  map[string]int
    mu         sync.Mutex
)

func main() {
    // Serving static files from the 'static' directory
    staticFs := http.FileServer(http.Dir("./static"))
    http.Handle("/static/", http.StripPrefix("/static/", staticFs))

    // Serving video files from the 'videos' directory
    videoFs := http.FileServer(http.Dir("./videos"))
    http.Handle("/videos/", http.StripPrefix("/videos/", videoFs))

    // Initialize visit count and play count
    visitCount = 0
    playCount = make(map[string]int)

    // Define a handler function for the root ("/") route
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        tmpl, err := template.ParseFiles("templates/index.html")
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        var videos []VideoInfo
        filepath.WalkDir("videos", func(path string, d fs.DirEntry, err error) error {
            if err != nil {
                return err
            }
            if !d.IsDir() {
                relPath, _ := filepath.Rel("videos", path)
                playCountValue := playCount[relPath]
                videos = append(videos, VideoInfo{
                    Path:      relPath,
                    Title:     relPath,
                    PlayCount: playCountValue,
                })
            }
            return nil
        })

        data := PageData{Videos: videos, VisitCount: visitCount}
        tmpl.Execute(w, data)
    })

    // Create a route to increment the visit count
    http.HandleFunc("/increment", func(w http.ResponseWriter, r *http.Request) {
        visitCount++
        http.Redirect(w, r, "/", http.StatusSeeOther) // Redirect back to the home page
    })

    // Create a route to increment the play count
    http.HandleFunc("/play", func(w http.ResponseWriter, r *http.Request) {
        title := r.URL.Query().Get("title")

        mu.Lock()
        playCount[title]++
        visitCount++
        mu.Unlock()

        http.Redirect(w, r, "/", http.StatusSeeOther) // Redirect back to the home page
    })

    log.Println("Listening on :8080...")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal(err)
    }
}
