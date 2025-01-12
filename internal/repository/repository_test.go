package repository_test

import (
	"fmt"
	"log"
	"os"
	"sync"
	"testing"

	"linkshrink/internal/config"
	"linkshrink/internal/repository"
	"linkshrink/internal/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

const testFilePath = "default_storage.json"

func setup() {
	// Удаляем файл перед каждым тестом, чтобы избежать конфликтов
	_ = os.Remove(testFilePath)
}

var tests = []struct {
	name string
	cfg  config.Config
}{
	{
		name: "Memo",
		cfg:  config.Config{},
	},
	{
		name: "File",
		cfg: config.Config{
			FileStoragePath: testFilePath,
		},
	},
}

func TestURLRepository_Save(t *testing.T) {
	setup()
	logger := zaptest.NewLogger(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := repository.NewStore(&tt.cfg, logger)
			require.NoError(t, err)
			// Тестирование сохранения URL
			err = repo.Save("abc123", "http://original.url")
			require.NoError(t, err)

			// Проверяем, что URL сохранен
			originalURL, err := repo.Find("abc123")
			require.NoError(t, err)
			assert.Equal(t, "http://original.url", originalURL)
		})
	}
}

func TestURLRepository_Find(t *testing.T) {
	setup()
	logger := zaptest.NewLogger(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := repository.NewStore(&tt.cfg, logger)
			require.NoError(t, err)
			// Сохраняем URL для дальнейшего поиска
			err = repo.Save("abc123", "http://original.url")
			require.NoError(t, err)

			// Тестирование поиска существующего URL
			originalURL, err := repo.Find("abc123")
			require.NoError(t, err)
			assert.Equal(t, "http://original.url", originalURL)

			// Тестирование поиска несуществующего URL
			_, err = repo.Find("nonexistent")
			assert.Error(t, err)
			assert.Equal(t, "URL not found", err.Error())
		})
	}
}

func TestURLRepository_ConcurrentAccess(t *testing.T) {
	setup()
	logger := zaptest.NewLogger(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := repository.NewStore(&tt.cfg, logger)
			require.NoError(t, err)

			// Используем WaitGroup для ожидания завершения всех горутин
			var wg sync.WaitGroup

			// Запускаем несколько горутин для сохранения URL
			for i := range utils.Intrange(0, 100) {
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
			for i := range utils.Intrange(0, 100) {
				originalURL, err := repo.Find(fmt.Sprintf("id%d", i))
				require.NoError(t, err)
				assert.Equal(t, fmt.Sprintf("http://url%d.com", i), originalURL)
			}
		})
	}
}

func TestURLRepository_LoadFromFile(t *testing.T) {
	// Создаем тестовый репозиторий и сохраняем несколько URL
	logger := zaptest.NewLogger(t)
	cfg := config.Config{
		FileStoragePath: testFilePath,
	}

	repo, err := repository.NewStore(&cfg, logger)
	require.NoError(t, err)
	// Сохраняем несколько URL
	_ = repo.Save("abc123", "http://original.url")
	_ = repo.Save("def456", "http://another.url")

	// Создаем новый репозиторий, который должен загрузить данные из файла
	repo2, err := repository.NewStore(&cfg, logger)
	require.NoError(t, err)

	// Проверяем, что данные были загружены корректно
	originalURL, err := repo2.Find("abc123")
	require.NoError(t, err)
	assert.Equal(t, "http://original.url", originalURL)

	originalURL, err = repo2.Find("def456")
	require.NoError(t, err)
	assert.Equal(t, "http://another.url", originalURL)

	// Удаляем тестовый файл после теста
	defer func() {
		if err := os.Remove(testFilePath); err != nil {
			log.Printf("Ошибка при удалении файла: %v", err)
		} else {
			log.Println("Файл успешно удален.")
		}
	}()
}
