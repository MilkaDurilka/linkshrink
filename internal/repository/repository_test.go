package repository_test

import (
	"fmt"
	"sync"
	"testing"

	"linkshrink/internal/repository"
	"linkshrink/internal/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURLRepository_Save(t *testing.T) {
	repo := repository.NewStore()

	// Тестирование сохранения URL
	id, err := repo.Save("http://original.url")
	require.NoError(t, err)

	// Проверяем, что URL сохранен
	originalURL, err := repo.Find(id)
	require.NoError(t, err)
	assert.Equal(t, "http://original.url", originalURL)
}

func TestURLRepository_Find(t *testing.T) {
	repo := repository.NewStore()

	// Сохраняем URL для дальнейшего поиска
	id, err := repo.Save("http://original.url")
	require.NoError(t, err)

	// Тестирование поиска существующего URL
	originalURL, err := repo.Find(id)
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
	var ids [100]string

	// Запускаем несколько горутин для сохранения URL
	for i := range utils.Intrange(0, 100) {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			id, err := repo.Save(fmt.Sprintf("http://url%d.com", index))
			ids[index] = id
			require.NoError(t, err)
		}(i)
	}

	// Ждем завершения всех горутин
	wg.Wait()

	// Проверяем, что все URL были сохранены
	for i, id := range ids {
		originalURL, err := repo.Find(id)
		require.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("http://url%d.com", i), originalURL)
	}
}
