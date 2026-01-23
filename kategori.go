package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Kategori struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

var Categories = []Kategori{
	{ID: 1, Name: "Food", Description: "Food is a category of food"},
	{ID: 2, Name: "Drink", Description: "Drink is a category of drink"},
	{ID: 3, Name: "Snack", Description: "Snack is a category of snack"},
}

func GetAllKategori(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Categories)
}

func CreateKategori(w http.ResponseWriter, r *http.Request) {
	var kategoriBaru Kategori
	err := json.NewDecoder(r.Body).Decode(&kategoriBaru)
	if err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	kategoriBaru.ID = len(Categories) + 1
	Categories = append(Categories, kategoriBaru)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(kategoriBaru)
}

func GetKategoriByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/kategori/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid kategori ID", http.StatusBadRequest)
		return
	}

	for _, k := range Categories {
		if k.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(k)
			return
		}
	}

	http.Error(w, "Kategori belum ada", http.StatusNotFound)
}

func UpdateKategori(w http.ResponseWriter, r *http.Request) {
	//get id dari request
	idStr := strings.TrimPrefix(r.URL.Path, "/api/kategori/")

	//ganti int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Kategori ID", http.StatusBadRequest)
		return
	}

	//get data dari request
	var updateKategori Kategori
	err = json.NewDecoder(r.Body).Decode(&updateKategori)
	if err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	//loop kategori, cari id, ganti sesuai data dari request
	for i := range Categories {
		if Categories[i].ID == id {
			updateKategori.ID = id
			Categories[i] = updateKategori

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(updateKategori)
			return
		}
	}

	//jika tidak ketemu
	http.Error(w, "Kategori belum ada", http.StatusNotFound)
}

func DeleteKategori(w http.ResponseWriter, r *http.Request) {
	//get id
	idStr := strings.TrimPrefix(r.URL.Path, "/api/kategori/")
	//ganti id int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Kategori ID", http.StatusBadRequest)
		return
	}

	//loop kategori cari ID, dapat index yang mau dihapus
	for i, k := range Categories {
		if k.ID == id {
			//bikin slice baru dengan data sebelum dan sesudah index
			Categories = append(Categories[:i], Categories[i+1:]...)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"message": "berhasil delete",
			})
			return
		}
	}

	http.Error(w, "Kategori belum ada", http.StatusNotFound)
}
