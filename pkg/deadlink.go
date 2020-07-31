package pocket

import (
	"os"
	"sync"

	"github.com/pkg/errors"
	log "github.com/whitekid/go-utils/logging"
	"github.com/whitekid/go-utils/request"
)

// CheckDeadLink ...
func CheckDeadLink() error {
	api := NewGetPocketAPI(os.Getenv("CONSUMER_KEY"), os.Getenv("ACCESS_TOKEN"))
	items, err := api.Articles.Get(GetOpts{Favorite: Favorited})
	if err != nil {
		return errors.Wrap(err, "articles.Get(Favorite)")
	}
	log.Debug("items: %d", len(items))

	ch := make(chan Article)

	go func() {
		notFoundItems := []string{"274841724", "758026316", "392120428", "494194220"}

		for _, v := range items {
			// skip if link was not found
			for _, link := range notFoundItems {
				if v.ResolvedURL == link {
					continue
				}
			}

			ch <- v
		}
		close(ch)
	}()

	// start 4 worker
	var itemsToDelete []string
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for article := range ch {
				log.Infof("checking %s %s", article.ItemID, article.ResolvedURL)
				resp, err := request.Get(article.ResolvedURL).
					Header("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.89 Safari/537.36").
					Do()
				if err != nil {
					log.Errorf("check link failed: itemID: %s,   link: %s, err: %s", article.ItemID, article.ResolvedURL, err)
					itemsToDelete = append(itemsToDelete, article.ItemID)
					continue
				}

				if !resp.Success() {
					log.Errorf("failed with %d, itemID: %s, link: %s", resp.StatusCode, article.ItemID, article.ResolvedURL)
					itemsToDelete = append(itemsToDelete, article.ItemID)
				}
			}
		}()
	}

	wg.Wait()

	log.Infof("deleting: %v", itemsToDelete)

	if err := api.Articles.Delete(itemsToDelete...); err != nil {
		return errors.Wrapf(err, "articles.Delete(%s)", itemsToDelete)
	}

	return nil
}
