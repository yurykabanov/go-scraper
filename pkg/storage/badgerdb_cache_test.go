package storage

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/dgraph-io/badger"
	"github.com/stretchr/testify/assert"
	"github.com/yurykabanov/scraper/pkg/domain"
)

func mustOpenTempBadgerDB() (*badger.DB, func()) {
	tempDir, _ := ioutil.TempDir(os.TempDir(), "scraper_badger_db_cache_test_")

	opts := badger.DefaultOptions
	opts.Dir = tempDir
	opts.ValueDir = tempDir

	log.SetOutput(ioutil.Discard)

	db, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}

	return db, func() {
		_ = db.Close()
		_ = os.RemoveAll(tempDir)
	}
}

func TestBadgerDBCache(t *testing.T) {
	db, cleanup := mustOpenTempBadgerDB()
	defer cleanup()

	cache := NewBadgerDBCache(db)

	v, err := cache.Get("whatever")

	assert.Nil(t, v)
	assert.Nil(t, err)

	src := &domain.Result{
		Url: "some_url",
		Body: "some_body",
		StatusCode: 123,
		Headers: map[string][]string{
			"X-Something": {"value"},
		},
	}
	err = cache.Put("whatever", src)

	assert.Nil(t, err)

	dst, err := cache.Get("whatever")

	assert.Equal(t, src, dst)
	assert.Nil(t, err)
}
