package domain

type Scraper interface {
	Scrape(task *FetchedTask) (*ScrapedTask, error)
}
