package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dustin/go-wikiparse"
	"github.com/mosuka/wikipedia-jsonl/util"
	"github.com/mosuka/wikipedia-jsonl/version"
	"github.com/spf13/cobra"
)

var (
	wikipediaJsonlCmd = &cobra.Command{
		Use:   "wikipedia-jsonl",
		Short: "Make JSONL file from Wikipedia dump",
		Long:  "Make JSONL file fromWikipedia dump",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check version flag.
			version_flag, err := cmd.Flags().GetBool("version")
			if err != nil {
				return err
			}
			if version_flag {
				fmt.Printf("wikipedia-jsonl version: %s\n", version.Version)
				return nil
			}

			parser, err := wikiparse.NewParser(os.Stdin)
			if err != nil {
				return err
			}

			for err == nil {
				var page *wikiparse.Page
				page, err = parser.Next()
				if err == nil {
					id := page.ID
					title := page.Title
					text := util.ParseArticle(title, page.Revisions[0].Text)
					timestamp := page.Revisions[0].Timestamp
					ns := page.Ns
					redirect := page.Redir.Title

					data := struct {
						Id        uint64 `json:"id"`
						Title     string `json:"title"`
						Text      string `json:"text"`
						Timestamp string `json:"timestamp"`
						Ns        uint64 `json:"ns"`
						Redirect  string `json:"redirect"`
					}{
						Id:        id,
						Title:     title,
						Text:      text,
						Timestamp: timestamp,
						Ns:        ns,
						Redirect:  redirect,
					}

					jsonBytes, err := json.Marshal(&data)
					if err != nil {
						return err
					}

					fmt.Println(string(jsonBytes))
				}
			}

			return nil
		},
	}
)

func Execute() error {
	if err := wikipediaJsonlCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	return nil
}

func init() {
	wikipediaJsonlCmd.Flags().BoolP("version", "v", false, "show version")
}
