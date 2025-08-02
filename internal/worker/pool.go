package worker

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/SerhiiZaderaIntellias/golangbootcamp/pkg/rss"
	"database/sql"
)

type Pool struct {
	db     *sql.DB
	jobs   chan string
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewPool(db *sql.DB, workers int) *Pool {
	ctx, cancel := context.WithCancel(context.Background())

	p := &Pool{
		db:     db,
		jobs:   make(chan string, 100),
		ctx:    ctx,
		cancel: cancel,
	}

	for i := 0; i < workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	return p
}

func (p *Pool) Submit(url string) {
	select {
	case p.jobs <- url:
		log.Printf("Job submitted: %s", url)
	default:
		log.Printf("Queue full, skipping: %s", url)
	}
}

func (p *Pool) worker(id int) {
	defer p.wg.Done()

	log.Printf("Worker #%d started", id)

	for {
		select {
		case <-p.ctx.Done():
			log.Printf("Worker #%d shutting down...", id)
			return
		case url := <-p.jobs:
			log.Printf("Worker #%d processing: %s", id, url)

			rssData, err := rss.FetchAndParse(url)
			if err != nil {
				log.Printf("Error fetching %s: %v", url, err)
				continue
			}

			err = rss.StoreItems(p.db, rssData.Channel[0].Items)
			if err != nil {
				log.Printf("Error storing items: %v", err)
				continue
			}

			log.Printf("Worker #%d finished: %s", id, url)
			time.Sleep(500 * time.Millisecond) // simulate load
		}
	}
}

func (p *Pool) Shutdown() {
	log.Println("Initiating graceful shutdown...")
	p.cancel()
	p.wg.Wait()
	log.Println("All workers shut down.")
}