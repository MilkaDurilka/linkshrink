package service_test

import (
	"linkshrink/internal/service"
	"linkshrink/internal/utils"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIDGenerator_GenerateID(t *testing.T) {
	generator := service.NewIDGenerator()

	// Генерируем несколько ID
	id1 := generator.GenerateID()
	id2 := generator.GenerateID()

	// Проверяем, что ID не пустые
	require.NotEmpty(t, id1, "Generated ID should not be empty")
	require.NotEmpty(t, id2, "Generated ID should not be empty")

	// Проверяем, что ID разные
	assert.NotEqual(t, id1, id2, "Generated IDs should be different")
}

func TestIDGenerator_ConcurrentAccess(t *testing.T) {
	generator := service.NewIDGenerator()

	// Используем WaitGroup для ожидания завершения всех горутин
	var wg sync.WaitGroup
	idSet := make(map[string]struct{}) // Для хранения уникальных ID
	mu := sync.Mutex{}                 // Мьютекс для защиты доступа к idSet

	// Запускаем несколько горутин для генерации ID
	for range utils.Intrange(0, 100) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := generator.GenerateID()

			// Защищаем доступ к idSet
			mu.Lock()
			idSet[id] = struct{}{} // Добавляем ID в сет
			mu.Unlock()
		}()
	}

	// Ждем завершения всех горутин
	wg.Wait()

	// Проверяем, что все сгенерированные ID уникальны
	assert.Equal(t, 100, len(idSet), "Generated IDs should be unique")
}
