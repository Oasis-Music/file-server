package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"oasis/store/utils"
	"path/filepath"

	"github.com/joho/godotenv"
)

func init() {
	env := flag.String("env", "dev", "specify .env filename for flag")
	flag.Parse()

	if err := godotenv.Load(".env." + *env); err != nil {
		log.Printf("No .env.%s file found, load default", *env)
		godotenv.Load(".env.dev")
	}
	log.Printf("loaded \".env.%s\"\n", *env)
}

func main() {

	mode := utils.GetEnv("ENV")
	fmt.Printf("in %q mode\n", mode)

	soundtracks := http.FileServer(http.Dir("./audio"))
	cover := http.FileServer(http.Dir("./cover"))
	http.Handle("/audio/", cors(http.StripPrefix("/audio/", soundtracks)))
	http.Handle("/cover/", http.StripPrefix("/cover/", cover))
	// createTrack

	http.HandleFunc("/createTrack", createTrack)

	fmt.Println("Server started at port 5000")
	http.ListenAndServe(":5000", nil)
}

func cors(fs http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// do your cors stuff
		// return if you do not want the FileServer handle a specific request
		w.Header().Set("Access-Control-Allow-Origin", "*")

		fs.ServeHTTP(w, r)
	}
}

type ResponseData struct {
	AudioPath      string  `json:"audioPath"`
	CoverImagePath *string `json:"coverPath"`
}

func createTrack(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.ParseMultipartForm(32>>20 + 512) // max 32.5MB

	mode := utils.GetEnv("ENV")

	var resp ResponseData

	audioFile, trackHeader, err := r.FormFile("soundtrack")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("Got audio: n:%s size:%d\n", trackHeader.Filename, trackHeader.Size)

	defer audioFile.Close()

	audioName := "test_*.mp3"
	if mode != "dev" {
		audioName = "*.mp3"
	}

	osAudioFile, err := ioutil.TempFile("./audio", audioName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer osAudioFile.Close()

	_, err = io.Copy(osAudioFile, audioFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp.AudioPath = filepath.Base(osAudioFile.Name())

	coverFile, coverHeader, err := r.FormFile("cover")
	if err != nil {

		err := json.NewEncoder(w).Encode(&resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		return
	}

	fmt.Printf("Got cover: n:%s size:%d\n", coverHeader.Filename, coverHeader.Size)

	defer coverFile.Close()

	coverExt := filepath.Ext(coverHeader.Filename)

	coverName := "test_*-" + coverExt
	if mode != "dev" {
		coverName = "*" + coverExt
	}

	osCoverFile, err := ioutil.TempFile("./cover", coverName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer osCoverFile.Close()

	_, err = io.Copy(osCoverFile, coverFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fp := filepath.Base(osCoverFile.Name())

	resp.CoverImagePath = &fp

	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

}
