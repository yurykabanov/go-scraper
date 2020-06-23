package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/sirupsen/logrus"
	"github.com/yurykabanov/scraper/pkg/domain"
	"github.com/yurykabanov/scraper/pkg/domain/fetcher"
	"github.com/yurykabanov/scraper/pkg/domain/scraper"
	"github.com/yurykabanov/scraper/pkg/pipeline"
	"github.com/yurykabanov/scraper/pkg/storage"
	"go.uber.org/ratelimit"
	"gopkg.in/yaml.v2"
)

/*
TODO feature list:
1. refactor shit & add tests
2. timer should use optimistic locks
3. http pool should use optimistic locks

6. add `goquery` selectors
7. add TextAction.CaptureType enum(html, text) to capture html (text is current behavior)

10. max attempts per task
11. supervisor should not emit repeated tasks

13. fix dirty hack in supervisor (chan with extremely large buffer) using external task storage

15. implement graph cancellation
*/

func mustOpenBadgerDB(path string) *badger.DB {
	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path

	// TODO: WTF?
	// ok, it seems new version of badger db always uses default logger :(
	//badger.SetLogger(logrus.StandardLogger())

	db, err := badger.Open(opts)
	if err != nil {
		logrus.WithError(err).Fatal("unable to open badger db")
	}

	return db
}

type projectFile struct {
	Tasks []*domain.Task          `yaml:"tasks"`
	Rules []domain.TaskDefinition `yaml:"rules"`
}

func mustReadProject(path string) ([]*domain.Task, map[string]domain.TaskDefinition) {
	f, err := os.Open(path)
	if err != nil {
		logrus.WithError(err).Fatal("unable to open project file")
	}

	var project projectFile

	dec := yaml.NewDecoder(bufio.NewReader(f))

	err = dec.Decode(&project)
	if err != nil {
		logrus.WithError(err).Fatal("unable to decode project file")
	}

	var rules = make(map[string]domain.TaskDefinition, len(project.Rules))

	for _, rule := range project.Rules {
		rules[rule.Name] = rule
	}

	return project.Tasks, rules
}

func main() {
	var err error

	// TODO: debug level from flag
	logrus.SetLevel(logrus.DebugLevel)

	// TODO: better flags
	if len(os.Args) < 2 {
		fmt.Println("Usage: scraper <project.yml>")
		os.Exit(1)
	}

	projectFilePath := os.Args[1]
	projectPath := filepath.Dir(projectFilePath)
	cachePath := filepath.Join(projectPath, "cache/")
	outputFilePath := filepath.Join(projectPath, "results.json")

	// TODO: validations
	tasks, definitions := mustReadProject(projectFilePath)

	db := mustOpenBadgerDB(cachePath)
	defer func() {
		err := db.Close()
		if err != nil {
			logrus.WithError(err).Fatal("unable to close badger db")
		}
	}()

	cacheStorage := storage.NewBadgerDBCache(db)

	//           supervisor <~              <~ fetcher
	// source ~> supervisor ~> cache_loader ~> fetcher ~> cache_saver ~> scraper ~> broadcast ~> exporter
	//                         cache_loader ~>                        ~> scraper
	//           supervisor <~                                                   <~ broadcast

	sourceStage := pipeline.SourceFromSlice(tasks)

	supervisorStage := pipeline.Supervisor(sourceStage.Output())

	cacheLoaderStage := pipeline.CacheLoader(supervisorStage.Output(),
		cacheStorage, pipeline.WithLoaderWorkers(16)) // TODO: flag

	fetcherStage := pipeline.Fetcher(
		cacheLoaderStage.Misses(),
		fetcher.New(
			&http.Client{
				Timeout: 30 * time.Second, // TODO: flag
			},
			fetcher.WithRateLimiter(ratelimit.New(20)),                   // TODO: flag
			fetcher.WithResponseValidators(fetcher.AcceptHttpCodes(200)), // TODO: flag
		),
		pipeline.WithFetcherWorkers(16), // TODO: flag
	)

	cacheSaverStage := pipeline.CacheSaver(fetcherStage.Results(), cacheStorage)

	scraperStage := pipeline.Scraper(
		pipeline.FetchedTaskMerger(cacheLoaderStage.Hits(), cacheSaverStage.Output()),
		scraper.New(domain.NewMapTaskRepository(definitions)),
		pipeline.WithScraperWorkers(16), // TODO: flag
	)

	broadcast := pipeline.BroadcastScrapedTask(scraperStage.Output(),
		pipeline.WithBroadcastBuffer(256)) // TODO: flag

	supervisorStage.Scraped = broadcast.Output()
	supervisorStage.Failed = fetcherStage.Errors()

	fout, err := os.Create(outputFilePath)
	if err != nil {
		logrus.WithError(err).Fatal("Unable to create file to export results")
	}
	defer func() {
		if err := fout.Close(); err != nil {
			logrus.WithError(err).Fatal("Unable to close result file")
		}
	}()

	bw := bufio.NewWriter(fout)
	defer func() {
		if err := bw.Flush(); err != nil {
			logrus.WithError(err).Fatal("Unable to flush buffer to result file")
		}
	}()

	exporterStage := pipeline.Exporter(broadcast.Output(), bw)

	graph := pipeline.NewGraph(
		// Source
		sourceStage, supervisorStage,
		// Cache related
		cacheLoaderStage, cacheSaverStage,
		// Fetcher & Scraper
		fetcherStage, scraperStage,
		// Final stages
		broadcast, exporterStage,
	)
	graph.RunStage()
}
