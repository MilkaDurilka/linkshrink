package repository_test

import (
	"fmt"
	"sync"
	"testing"

	"linkshrink/internal/repository" // Убедитесь, что путь к вашему пакету правильный

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURLRepository_Save(t *testing.T) {
	repo := repository.NewStore()

	// Тестирование сохранения URL
	err := repo.Save("abc123", "http://original.url")
	require.NoError(t, err)

	// Проверяем, что URL сохранен
	originalURL, err := repo.Find("abc123")
	require.NoError(t, err)
	assert.Equal(t, "http://original.url", originalURL)
}

func TestURLRepository_Find(t *testing.T) {
	repo := repository.NewStore()

	// Сохраняем URL для дальнейшего поиска
	err := repo.Save("abc123", "http://original.url")
	require.NoError(t, err)

	// Тестирование поиска существующего URL
	originalURL, err := repo.Find("abc123")
	require.NoError(t, err)
	assert.Equal(t, "http://original.url", originalURL)

	// Тестирование поиска несуществующего URL
	_, err = repo.Find("nonexistent")
	assert.Error(t, err)
	assert.Equal(t, "URL not found", err.Error())
}

func TestURLRepository_ConcurrentAccess(t *testing.T) {
	repo := repository.NewStore()

	// Используем WaitGroup для ожидания завершения всех горутин
	var wg sync.WaitGroup

	// Запускаем несколько горутин для сохранения URL
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			err := repo.Save(fmt.Sprintf("id%d", id), fmt.Sprintf("http://url%d.com", id))
			require.NoError(t, err)
		}(i)
	}

	// Ждем завершения всех горутин
	wg.Wait()

	// Проверяем, что все URL были сохранены
	for i := 0; i < 100; i++ {
		originalURL, err := repo.Find(fmt.Sprintf("id%d", i))
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("http://url%d.com", i), originalURL)
	}
}
