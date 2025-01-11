package filestore

import (
	"encoding/json"
	"errors"
	"io"
	memorystore "linkshrink/internal/repository/memory_store"
	"linkshrink/internal/utils/logger"
	"os"
	"sync"

	"go.uber.org/zap"
)

type URLData struct {
	UUID        string `json:"uuid"`
	OriginalURL string `json:"original_url"`
}

type IFileStore interface {
	Save(id string, originalURL string) error
	Find(id string) (string, error)
	LoadFromFile() error
	SaveToFile() error
}

type FileStore struct {
	memory   memorystore.MemoryStore // Встраивание MemoryStore
	mu       *sync.Mutex             // Мьютекс для обеспечения потокобезопасности
	logger   logger.Logger
	filePath string
}

func NewFileStore(filePath string, log logger.Logger) *FileStore {
	componentLogger := log.With(zap.String("component", "FileStore"))
	repo := &FileStore{
		memory:   *memorystore.NewMemoryStore(log),
		filePath: filePath,
		mu:       &sync.Mutex{},
		logger:   componentLogger,
	}

	if err := repo.LoadFromFile(); err != nil {
		componentLogger.Error("Ошибка при загрузке из файла", zap.Error(err))
	}

	return repo
}

// LoadFromFile загружает данные из файла в репозиторий.
func (r *FileStore) LoadFromFile() error {
	file, err := os.Open(r.filePath)
	if err != nil {
		return errors.New("не удалось открыть файл: " + err.Error())
	}
	defer func() {
		if err := file.Close(); err != nil {
			r.logger.Error("Ошибка при закрытии файла", zap.Error(err))
		}
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		return errors.New("не удалось прочитать файл: " + err.Error())
	}

	var urls []URLData
	if err := json.Unmarshal(data, &urls); err != nil {
		return errors.New("не удалось декодировать файл: " + err.Error())
	}

	for _, url := range urls {
		r.memory.Store[url.UUID] = url.OriginalURL
	}

	return nil
}

// SaveToFile сохраняет данные репозитория в файл.
func (r *FileStore) SaveToFile() error {
	const initialCapacity = 1000
	urls := make([]URLData, 0, initialCapacity)
	for id, originalURL := range r.memory.Store {
		urls = append(urls, URLData{
			UUID:        id,
			OriginalURL: originalURL,
		})
	}

	data, err := json.Marshal(urls)
	if err != nil {
		return errors.New("не удалось сериализовать данные: " + err.Error())
	}

	const filePermission = 0o600 // Read and write for owner only
	if err := os.WriteFile(r.filePath, data, filePermission); err != nil {
		return errors.New("не удалось записать файл: " + err.Error())
	}
	return nil
}

// Save сохраняет оригинальный URL по ID и затем сохраняет в файл.
func (r *FileStore) Save(id string, originalURL string) error {
	r.mu.Lock() // Блокируем мьютекс
	defer r.mu.Unlock()

	if err := r.memory.Save(id, originalURL); err != nil {
		return errors.New("не удалось сохранить в файл: " + err.Error())
	}
	return r.SaveToFile() // Сохраняем в файл после сохранения в память
}

func (r *FileStore) Find(id string) (string, error) {
	originalURL, err := r.memory.Find(id)
	if err != nil {
		return "", errors.New("URL not found")
	}
	return originalURL, nil
}
