package repository

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"sync"
)

var (
	ErrURLNotFound     = errors.New("URL not found")
	ErrIDAlreadyExists = errors.New("ID already exists")
)

type URLData struct {
	UUID        string `json:"uuid"`
	OriginalURL string `json:"original_url"`
}

type IURLRepository interface {
	Save(id string, originalURL string) error
	Find(id string) (string, error)
	LoadFromFile() error
	SaveToFile() error
}

type URLRepository struct {
	store    map[string]string // Хранилище для хранения пар ID и оригинальных URL
	mu       *sync.Mutex       // Мьютекс для обеспечения потокобезопасности
	filePath string
}

// NewStore создает новый экземпляр URLRepository.
func NewStore(filePath string) *URLRepository {
	repo := &URLRepository{
		store:    make(map[string]string),
		mu:       &sync.Mutex{},
		filePath: filePath,
	}

	if err := repo.LoadFromFile(); err != nil {
		log.Printf("Ошибка при загрузке из файла: %v", err)
	}

	return repo
}

// LoadFromFile загружает данные из файла в репозиторий.
func (r *URLRepository) LoadFromFile() error {
	file, err := os.Open(r.filePath)
	if err != nil {
		return errors.New("не удалось открыть файл: " + err.Error())
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("Ошибка при закрытии файла: %v", closeErr)
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
		r.store[url.UUID] = url.OriginalURL
	}

	return nil
}

// SaveToFile сохраняет данные репозитория в файл.
func (r *URLRepository) SaveToFile() error {
	// r.mu.Lock()
	// defer r.mu.Unlock()

	const initialCapacity = 1000
	urls := make([]URLData, 0, initialCapacity)
	for id, originalURL := range r.store {
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

// Save сохраняет оригинальный URL по ID.
func (r *URLRepository) Save(id string, originalURL string) error {
	r.mu.Lock()         // Блокируем мьютекс
	defer r.mu.Unlock() // Разблокируем мьютекс после завершения работы

	_, ok := r.store[id]
	if !ok {
		r.store[id] = originalURL
		if err := r.SaveToFile(); err != nil {
			return err
		}
		return nil
	}

	return ErrIDAlreadyExists
}

// Find ищет оригинальный URL по ID.
func (r *URLRepository) Find(id string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	originalURL, ok := r.store[id] // Проверяем, существует ли ID в хранилище
	if !ok {
		return "", ErrURLNotFound
	}
	return originalURL, nil
}
