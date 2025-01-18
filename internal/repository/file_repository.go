package repository

import (
	"encoding/json"
	"errors"
	"io"
	"linkshrink/internal/utils"
	"linkshrink/internal/utils/logger"
	"os"
	"sync"

	"go.uber.org/zap"
)

type FileStore struct {
	memory      MemoryStore // Встраивание MemoryStore
	mu          *sync.Mutex // Мьютекс для обеспечения потокобезопасности
	idGenerator *utils.IDGenerator
	logger      logger.Logger
	filePath    string
}

func NewFileStore(filePath string, log logger.Logger) (*FileStore, error) {
	memory, _ := NewMemoryStore(log)
	repo := &FileStore{
		memory:      *memory,
		filePath:    filePath,
		mu:          &sync.Mutex{},
		logger:      log,
		idGenerator: utils.NewIDGenerator(),
	}

	if err := repo.LoadFromFile(); err != nil {
		return nil, err
	}

	return repo, nil
}

// LoadFromFile загружает данные из файла в репозиторий.
func (r *FileStore) LoadFromFile() error {
	file, err := os.Open(r.filePath)
	if err != nil {
		file, err = os.Create(r.filePath)
		if err != nil {
			return errors.New("не удалось прочитать файл: " + err.Error())
		}
		const filePermission = 0o600
		data, err := json.Marshal([]string{})
		if err != nil {
			return errors.New("не удалось прочитать файл: " + err.Error())
		}
		err = os.WriteFile(r.filePath, data, filePermission)
		if err != nil {
			r.logger.Error("Ошибка при редактировании файла", zap.Error(err))
		}
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
func (r *FileStore) Save(originalURL string) (string, error) {
	r.mu.Lock() // Блокируем мьютекс
	defer r.mu.Unlock()
	id, err := r.memory.Save(originalURL)

	if err != nil {
		return "", errors.New("не удалось сохранить в файл: " + err.Error())
	}

	if err := r.SaveToFile(); err != nil { // Сохраняем в файл после сохранения в память
		return "", errors.New("не удалось сохранить в файл: " + err.Error())
	}
	return id, nil
}

func (r *FileStore) Find(id string) (string, error) {
	originalURL, err := r.memory.Find(id)
	if err != nil {
		return "", errors.New("URL not found")
	}
	return originalURL, nil
}
