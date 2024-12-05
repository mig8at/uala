package seeder

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"
	"user_service/internal/domain/models"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Seeder struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewSeeder(db *gorm.DB, redis *redis.Client) *Seeder {
	return &Seeder{db: db, redis: redis}
}

func (s *Seeder) Seed() {
	s.Clean()
	// Definir usuarios de prueba
	users := []models.User{
		{
			ID:       "2a42c7ae-7f78-4e36-8358-902342fe23f1",
			Name:     "Juan Pérez",
			Email:    "juan.perez@example.com",
			Nickname: "@juanito",
			Bio:      "Desarrollador de software",
			Avatar:   "https://picsum.photos/200/200?random=1",
		},
		{
			ID:       "83836283-0760-4879-a7df-af4769a2d1a4",
			Name:     "María Gómez",
			Email:    "maria.gomez@example.com",
			Nickname: "@mary",
			Bio:      "Diseñadora gráfica",
			Avatar:   "https://picsum.photos/200/200?random=2",
		},
		{
			ID:       "2327a87b-3fe7-4bc9-a275-75d33358f1bc",
			Name:     "Carlos Ramírez",
			Email:    "carlos.ramirez@example.com",
			Nickname: "@carlitos",
			Bio:      "Ingeniero mecánico",
			Avatar:   "https://picsum.photos/200/200?random=3",
		},
		{
			ID:       "26474281-d97a-474d-a593-68aa1c1f48ef",
			Name:     "Ana Fernández",
			Email:    "ana.fernandez@example.com",
			Nickname: "@anita",
			Bio:      "Médico pediatra",
			Avatar:   "https://picsum.photos/200/200?random=4",
		},
		{
			ID:       "2e3b1b92-62ba-4308-8872-6a3d964f3a80",
			Name:     "Pedro López",
			Email:    "pedro.lopez@example.com",
			Nickname: "@pedrito",
			Bio:      "Arquitecto de soluciones",
			Avatar:   "https://picsum.photos/200/200?random=5",
		},
		{
			ID:       "444f79b1-c805-4998-a4db-24086790b031",
			Name:     "Sofía Martínez",
			Email:    "sofia.martinez@example.com",
			Nickname: "@sofi",
			Bio:      "Chef profesional",
			Avatar:   "https://picsum.photos/200/200?random=6",
		},
		{
			ID:       "4774f12c-8c0c-4bfc-9692-f2b220553023",
			Name:     "Luis Hernández",
			Email:    "luis.hernandez@example.com",
			Nickname: "@luisito",
			Bio:      "Abogado corporativo",
			Avatar:   "https://picsum.photos/200/200?random=7",
		},
		{
			ID:       "4eef46a3-fe6d-4a60-af4c-fd354f987cc8",
			Name:     "Laura Castro",
			Email:    "laura.castro@example.com",
			Nickname: "@lau",
			Bio:      "Psicóloga clínica",
			Avatar:   "https://picsum.photos/200/200?random=8",
		},
		{
			ID:       "5f5d1800-e9a2-496f-8cae-52ea1f587acc",
			Name:     "Miguel Ángel",
			Email:    "miguel.angel@example.com",
			Nickname: "@mike",
			Bio:      "Artista plástico",
			Avatar:   "https://picsum.photos/200/200?random=9",
		},
		{
			ID:       "65c44c69-8663-4e00-ad7c-90fd48e95102",
			Name:     "Carmen Díaz",
			Email:    "carmen.diaz@example.com",
			Nickname: "@carmi",
			Bio:      "Escritora independiente",
			Avatar:   "https://picsum.photos/200/200?random=10",
		},
		{
			ID:       "6cebd913-085d-4144-a946-3d8fdfadae36",
			Name:     "José Torres",
			Email:    "jose.torres@example.com",
			Nickname: "@joseto",
			Bio:      "Ingeniero civil",
			Avatar:   "https://picsum.photos/200/200?random=11",
		},
		{
			ID:       "7f35a8d8-7af5-4e4d-a26c-e0afe5245bca",
			Name:     "Isabel Sánchez",
			Email:    "isabel.sanchez@example.com",
			Nickname: "@isa",
			Bio:      "Enfermera",
			Avatar:   "https://picsum.photos/200/200?random=12",
		},
		{
			ID:       "812f2527-cb7d-4b8a-9899-8035c301cf29",
			Name:     "Diego Ruiz",
			Email:    "diego.ruiz@example.com",
			Nickname: "@dieguito",
			Bio:      "Profesor de historia",
			Avatar:   "https://picsum.photos/200/200?random=13",
		},
		{
			ID:       "8372ea9d-2424-4650-b10b-9f9a15a50a6e",
			Name:     "Valeria Morales",
			Email:    "valeria.morales@example.com",
			Nickname: "@vale",
			Bio:      "Fotógrafa profesional",
			Avatar:   "https://picsum.photos/200/200?random=14",
		},
		{
			ID:       "c2dae4b6-d8ba-4b4c-89ef-39e0d92df28e",
			Name:     "Sebastián Navarro",
			Email:    "sebastian.navarro@example.com",
			Nickname: "@sebas",
			Bio:      "Analista financiero",
			Avatar:   "https://picsum.photos/200/200?random=15",
		},
		{
			ID:       "d6d6a128-ad31-45bd-8758-7bddd9871a05",
			Name:     "Gabriela Vargas",
			Email:    "gabriela.vargas@example.com",
			Nickname: "@gabi",
			Bio:      "Marketing digital",
			Avatar:   "https://picsum.photos/200/200?random=16",
		},
		{
			ID:       "e1dfb9c5-5fb1-4157-809f-4c78c9b5d355",
			Name:     "Manuel Ortiz",
			Email:    "manuel.ortiz@example.com",
			Nickname: "@manu",
			Bio:      "Ingeniero de software",
			Avatar:   "https://picsum.photos/200/200?random=17",
		},
		{
			ID:       "e574036a-4027-4dca-b698-6f9be442b03f",
			Name:     "Camila Pérez",
			Email:    "camila.perez@example.com",
			Nickname: "@cami",
			Bio:      "Consultora de negocios",
			Avatar:   "https://picsum.photos/200/200?random=18",
		},
		{
			ID:       "f3335e2a-d681-4e2a-9024-1164c85c5f87",
			Name:     "Rodrigo Rojas",
			Email:    "rodrigo.rojas@example.com",
			Nickname: "@rodrigo",
			Bio:      "Empresario",
			Avatar:   "https://picsum.photos/200/200?random=19",
		},
		{
			ID:       "fbe5a0a7-8ecb-4868-a9a8-54a280cd8edc",
			Name:     "Natalia Vega",
			Email:    "natalia.vega@example.com",
			Nickname: "@naty",
			Bio:      "Investigadora científica",
			Avatar:   "https://picsum.photos/200/200?random=20",
		},
	}

	rand.Seed(time.Now().UnixNano())

	// Crear mapa para almacenar relaciones de seguimiento y evitar duplicados
	followMap := make(map[string]map[string]bool)

	// Inicializar el mapa
	for _, user := range users {
		followMap[user.ID] = make(map[string]bool)
	}

	// Asignar seguidores a cada usuario
	for _, user := range users {
		// Generar un número aleatorio de seguidores entre 5 y 10
		numFollowers := rand.Intn(6) + 5

		// insertar usuario en la base de datos
		if err := s.db.Create(&user).Error; err != nil {
			log.Fatalf("Error al insertar usuario %s: %v", user.Email, err)
		}

		// insertar usuario en Redis
		userData, err := redisUser(&user)
		if err != nil {
			log.Fatalf("Error al serializar usuario %s: %v", user.Email, err)
		}

		key := fmt.Sprintf("users:%s", user.ID)
		if err := s.redis.Set(context.Background(), key, userData, 0).Err(); err != nil {
			log.Fatalf("Error al guardar usuario %s en Redis: %v", user.Email, err)
		}

		// Obtener una lista de otros usuarios (excluyendo al usuario actual)
		var otherUsers []models.User
		for _, u := range users {
			if u.ID != user.ID {
				otherUsers = append(otherUsers, u)
			}
		}

		// Mezclar la lista de otros usuarios
		rand.Shuffle(len(otherUsers), func(i, j int) {
			otherUsers[i], otherUsers[j] = otherUsers[j], otherUsers[i]
		})

		// Ajustar si el número de seguidores excede el número de usuarios disponibles
		if numFollowers > len(otherUsers) {
			numFollowers = len(otherUsers)
		}

		// Seleccionar los primeros 'numFollowers' usuarios como seguidores
		followers := otherUsers[:numFollowers]

		// Asignar seguidores al usuario actual
		for _, follower := range followers {
			// Verificar si ya existe la relación para evitar duplicados
			if !followMap[user.ID][follower.ID] {
				// Crear la relación de seguimiento en la base de datos
				follow := models.Follower{
					UserID:     user.ID,
					FollowerID: follower.ID,
				}
				if err := s.db.Create(&follow).Error; err != nil {
					log.Printf("Error al insertar seguimiento de %s a %s: %v", follower.Email, user.Email, err)
				}

				// Almacenar la relación en Redis
				followingKey := fmt.Sprintf("following:%s", follower.ID)
				followersKey := fmt.Sprintf("followers:%s", user.ID)

				s.redis.SAdd(context.Background(), followingKey, user.ID)
				s.redis.SAdd(context.Background(), followersKey, follower.ID)

				// Marcar la relación como existente
				followMap[user.ID][follower.ID] = true
			}
		}
	}
}

func (s *Seeder) Clean() {

	for _, table := range []string{"followers", "users"} {
		err := s.db.Exec("DELETE FROM " + table).Error
		if err != nil {
			log.Fatalf("Error al borrar el contenido de la tabla %s: %v", table, err)
		}
		log.Printf("Contenido de la tabla %s eliminado exitosamente", table)
	}

	deleteKeysWithPrefix(context.Background(), s.redis, "users:")
	deleteKeysWithPrefix(context.Background(), s.redis, "followers:")
}

func redisUser(u *models.User) ([]byte, error) {
	jsonData, err := json.Marshal(struct {
		Name     string `json:"name"`
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
	}{
		Name:     u.Name,
		Nickname: u.Nickname,
		Avatar:   u.Avatar,
	})
	if err != nil {
		return nil, fmt.Errorf("error al serializar el Tweet a JSON: %w", err)
	}
	return jsonData, nil
}

func deleteKeysWithPrefix(ctx context.Context, rdb *redis.Client, prefix string) error {
	// Usamos SCAN para buscar claves que coincidan con el prefijo
	iter := rdb.Scan(ctx, 0, prefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		err := rdb.Del(ctx, key).Err()
		if err != nil {
			return fmt.Errorf("error al borrar la clave %s: %w", key, err)
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("error durante la iteración de SCAN: %w", err)
	}

	return nil
}
