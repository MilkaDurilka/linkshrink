package utils

func Intrange(start, end int) <-chan int {
	ch := make(chan int)
	go func() {
		for i := start; i < end; i++ {
			ch <- i
		}
		close(ch)
	}()
	return ch
}

type BatchShortenParam struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchShortenReturnParam struct {
	CorrelationID string
	ShortURL      string
}