package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/whitekid/go-utils/log"
	pocket "github.com/whitekid/pocket-pick/pkg"
	"github.com/whitekid/pocket-pick/pkg/config"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:          "delete item_id_or_url",
		Long:         "delete article",
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteArticle(args...)
		},
	})
}

func deleteArticle(itemIDs ...string) error {
	api := pocket.NewGetPocketAPI(config.ConsumerKey(), config.AccessToken())
	for _, arg := range itemIDs {
		// delete by url
		if strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://") {
			items, err := api.Articles.Get(pocket.GetOpts{
				Search: arg,
			})
			if err != nil {
				return errors.Wrapf(err, "articles.Get(%+v)", arg)
			}

			if len(items) != 1 {
				return fmt.Errorf("not found: %s", arg)
			}

			for _, v := range items {
				log.Infof("deleting %s, %s", v.ItemID, v.ResolvedURL)
				if err := api.Articles.Delete(v.ItemID); err != nil {
					errors.Wrapf(err, "articles.Delete(%s)", v.ItemID)
					return err
				}
			}
		} else {
			// delete by id
			if _, err := strconv.Atoi(arg); err != nil {
				return fmt.Errorf("%s is not valid id", arg)
			}

			log.Infof("deleting item %s", arg)
			if err := api.Articles.Delete(arg); err != nil {
				return errors.Wrapf(err, "articles.Delete(%s)", arg)
			}
		}
	}

	return nil
}
