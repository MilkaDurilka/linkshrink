package service

import (
	"math/rand"
	"strconv"
	"sync"
	"time"
)

// IDGenerator - генератор уникальных идентификаторов.
type IDGenerator struct {
	randGen *rand.Rand
	mu      sync.Mutex
}

func NewIDGenerator() *IDGenerator {
	return &IDGenerator{
		randGen: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (g *IDGenerator) GenerateID() string {
	g.mu.Lock()         // Блокируем доступ к генератору
	defer g.mu.Unlock() // Освобождаем блокировку после выполнения

	return strconv.FormatInt(g.randGen.Int63(), 10)
}
