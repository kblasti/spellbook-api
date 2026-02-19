package main

import (
  "net/http"
  "os"
  "github.com/joho/godotenv"
  "github.com/kblasti/spellbook/internal/database"
  "github.com/kblasti/spellbook/internal/api"
  "database/sql"
  "log"
)

func main() {
  godotenv.Load()
  dbURL := os.Getenv("POSTGRES_DBURL")
  db, err := sql.Open("postgres", dbURL)
  if err != nil {
      log.Fatal(err)
  }
  dbQueries := database.New(db)
  cfg := &api.APIConfig{
    DB:         dbQueries,
    Platform:   os.Getenv("PLATFORM"),
    Secret:     os.Getenv("SECRET"),
  }
  port := os.Getenv("PORT")
  filepathRoot:= "/app/"
  mux := http.NewServeMux()
  mux.Handle(filepathRoot, http.StripPrefix("/app", http.FileServer(http.Dir("."))))

  mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/plain; charset=utf-8")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK\n"))
    })
  mux.HandleFunc("GET /api/index.html", func(w http.ResponseWriter, r *http.Request) {
      w.Header().Set("Content-Type", "text/plain; charset=utf-8")
      w.WriteHeader(http.StatusOK)
      w.Write([]byte("Welcome to the api homepage!\nThis work includes material from the System Reference Document 5.2.1 (“SRD 5.2.1”) by Wizards of the\nCoast LLC, available at https://www.dndbeyond.com/srd. The SRD 5.2.1 is licensed under the Creative\nCommons Attribution 4.0 International License, available at https://creativecommons.org/licenses/by/4.0/\nlegalcode.\n"))
  })
  mux.Handle( 
    "POST /api/spells/update/{index}", 
    cfg.AuthMiddleware( 
      cfg.AdminOnly( 
        http.HandlerFunc(cfg.HandlerUpdateSpell), 
      ), 
    ), 
  )
  mux.Handle(
    "POST /api/admin/users",
    cfg.AuthMiddleware(
      cfg.AdminOnly(
        http.HandlerFunc(cfg.HandlerCreateAdminUser),
      ),
    ),
  )
  mux.HandleFunc("GET /api/spells", cfg.HandlerGetAllSpells)
  mux.HandleFunc("GET /api/spells/{index}", cfg.HandlerGetSpell)
  mux.HandleFunc("GET /api/classes/{class}", cfg.HandlerGetSpellsClass)
  mux.HandleFunc("GET /api/subclasses/{subclass}", cfg.HandlerGetSpellsSubclass)
  mux.HandleFunc("GET /api/spells/levels/{level}", cfg.HandlerGetSpellsLevel)
  mux.HandleFunc("GET /api/spells/concentration", cfg.HandlerGetSpellsConcentration)
  mux.HandleFunc("GET /api/spells/ritual", cfg.HandlerGetSpellsRitual)
  mux.HandleFunc("POST /api/users", cfg.HandlerCreateUser)
  mux.HandleFunc("POST /api/login", cfg.HandlerLogin)
  mux.HandleFunc("POST /api/refresh", cfg.HandlerRefresh)
  mux.HandleFunc("POST /api/revoke", cfg.HandlerRevoke)
  mux.HandleFunc("PUT /api/users", cfg.HandlerUpdateUser)
  mux.HandleFunc("POST /api/users/delete", cfg.HandlerDeleteUser)
  mux.HandleFunc("POST /api/characters/delete", cfg.HandlerDeleteCharacter)
  mux.HandleFunc("POST /api/characters", cfg.HandlerCreateCharacter)
  mux.HandleFunc("PUT /api/characters", cfg.HandlerUpdateCharacter)
  mux.HandleFunc("POST /api/characters/slots", cfg.HandlerGetSpellSlots)
  mux.HandleFunc("GET /api/characters", cfg.HandlerGetUserCharacters)
  mux.HandleFunc("POST /api/characters/spells", cfg.HandlerCharacterSpells)
  mux.HandleFunc("POST /api/characters/spells/list", cfg.HandlerGetCharacterSpells)
  mux.HandleFunc("POST /api/characters/spells/delete", cfg.HandlerRemoveCharacterSpell)

  srv := &http.Server{
        Addr:    ":" + port,
        Handler: api.EnableCORS(mux),
    }  

  log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
  log.Fatal(srv.ListenAndServe())
}