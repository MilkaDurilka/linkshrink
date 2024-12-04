package service

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// IDGenerator - генератор уникальных идентификаторов
type IDGenerator struct {
	mu      sync.Mutex
	randGen *rand.Rand
}

func NewIDGenerator() *IDGenerator {
	return &IDGenerator{
		randGen: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (g *IDGenerator) GenerateID() string {
	g.mu.Lock()         // Блокируем доступ к генератору
	defer g.mu.Unlock() // Освобождаем блокировку после выполнения

	return fmt.Sprintf("%d", g.randGen.Int63())
}
