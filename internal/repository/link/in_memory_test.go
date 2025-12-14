package link

import (
	"fmt"
	"testing"

	"github.com/Alexey-zaliznuak/shortener/internal/config"
	"github.com/Alexey-zaliznuak/shortener/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func BenchmarkCreate(b *testing.B) {
	repo := NewInMemoryLinksRepository(&config.AppConfig{})

	for i := 0; i < b.N; i++ {
		u, err := uuid.NewRandom()

		require.NoError(b, err)

		repo.Create(&model.CreateLinkDto{FullURL: fmt.Sprintf("http://example.com/%s", u.String()), Shortcut: u.String()}, u.String(), nil)
	}
}

func BenchmarkGetByShortcut(b *testing.B) {
	repo := NewInMemoryLinksRepository(&config.AppConfig{})

	// Подготовка данных
	shortcuts := make([]string, 1000)
	for i := range 1000 {
		u, _ := uuid.NewRandom()
		shortcut := u.String()
		shortcuts[i] = shortcut
		repo.Create(&model.CreateLinkDto{
			FullURL:  fmt.Sprintf("http://example.com/%s", shortcut),
			Shortcut: shortcut,
		}, u.String(), nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.GetByShortcut(shortcuts[i%1000])
	}
}

func BenchmarkGetByFullURL(b *testing.B) {
	repo := NewInMemoryLinksRepository(&config.AppConfig{})

	// Подготовка данных
	urls := make([]string, 1000)
	for i := range 1000 {
		u, _ := uuid.NewRandom()
		fullURL := fmt.Sprintf("http://example.com/%s", u.String())
		urls[i] = fullURL
		repo.Create(&model.CreateLinkDto{
			FullURL:  fullURL,
			Shortcut: u.String(),
		}, u.String(), nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.GetByFullURL(urls[i%1000])
	}
}

func BenchmarkGetByUserID(b *testing.B) {
	repo := NewInMemoryLinksRepository(&config.AppConfig{})

	// Подготовка данных
	userID := uuid.New().String()
	for range 100 {
		u, _ := uuid.NewRandom()
		repo.Create(&model.CreateLinkDto{
			FullURL:  fmt.Sprintf("http://example.com/%s", u.String()),
			Shortcut: u.String(),
		}, userID, nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.GetByUserID(userID)
	}
}

func BenchmarkCreateConcurrent(b *testing.B) {
	repo := NewInMemoryLinksRepository(&config.AppConfig{})

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			u, err := uuid.NewRandom()
			require.NoError(b, err)

			repo.Create(&model.CreateLinkDto{
				FullURL:  fmt.Sprintf("http://example.com/%s", u.String()),
				Shortcut: u.String(),
			}, u.String(), nil)
		}
	})
}

func BenchmarkGetByShortcutConcurrent(b *testing.B) {
	repo := NewInMemoryLinksRepository(&config.AppConfig{})

	// Подготовка данных
	shortcuts := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		u, _ := uuid.NewRandom()
		shortcut := u.String()
		shortcuts[i] = shortcut
		repo.Create(&model.CreateLinkDto{
			FullURL:  fmt.Sprintf("http://example.com/%s", shortcut),
			Shortcut: shortcut,
		}, u.String(), nil)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			repo.GetByShortcut(shortcuts[i%1000])
			i++
		}
	})
}
