package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// LoadEnv membaca .env bila ada. Di server, nilai biasanya datang dari
// environment sistem, jadi berkas yang tidak ada bukan kesalahan.
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("berkas .env tidak ditemukan, memakai environment sistem")
	}
}

func GetEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func GetEnvInt(key string, fallback int) int {
	v, err := strconv.Atoi(GetEnv(key, ""))
	if err != nil {
		return fallback
	}
	return v
}
