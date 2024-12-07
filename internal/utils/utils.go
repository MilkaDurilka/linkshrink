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
