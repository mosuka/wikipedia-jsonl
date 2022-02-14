package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/dustin/go-wikiparse"
	"github.com/mosuka/wikipedia-jsonl/version"
	"github.com/semantosoph/gowiki"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

var (
	wikipediaJsonlCmd = &cobra.Command{
		Use:   "wikipedia-jsonl",
		Short: "Make JSONL file from Wikipedia dump",
		Long:  "Make JSONL file fromWikipedia dump",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check version flag.
			versionFlag, err := cmd.Flags().GetBool("version")
			if err != nil {
				return err
			}
			if versionFlag {
				fmt.Printf("wikipedia-jsonl version: %s\n", version.Version)
				return nil
			}

			// Check database flag.
			databasePath, err := cmd.Flags().GetString("database-path")
			if err != nil {
				return err
			}
			db, err := sql.Open("sqlite", databasePath)
			if err != nil {
				return err
			}
			defer db.Close()

			// Check abstruct flag.
			abstruct, err := cmd.Flags().GetBool("abstruct")
			if err != nil {
				return err
			}

			// Check ns flag.
			ns, err := cmd.Flags().GetBool("ns")
			if err != nil {
				return err
			}

			// Check redirect flag.
			redirect, err := cmd.Flags().GetBool("redirect")
			if err != nil {
				return err
			}

			// Check text-links flag.
			textLinks, err := cmd.Flags().GetBool("text-links")
			if err != nil {
				return err
			}

			// Check media flag.
			media, err := cmd.Flags().GetBool("media")
			if err != nil {
				return err
			}

			// Check links flag.
			links, err := cmd.Flags().GetBool("links")
			if err != nil {
				return err
			}

			// Check external-links flag.
			externalLinks, err := cmd.Flags().GetBool("external-links")
			if err != nil {
				return err
			}

			// Check categories flag.
			categories, err := cmd.Flags().GetBool("categories")
			if err != nil {
				return err
			}

			parser, err := wikiparse.NewParser(os.Stdin)
			if err != nil {
				return err
			}

			re := regexp.MustCompile(`\n{2,}`)

			for err == nil {
				var page *wikiparse.Page
				page, err = parser.Next()
				if err != nil {
					break
				}

				data := make(map[string]interface{})

				data["id"] = page.ID
				data["title"] = page.Title
				data["timestamp"] = page.Revisions[0].Timestamp

				article, err := GetArticle(page.Title, page.Revisions[0].Text)
				if err != nil {
					return err
				}

				text := ""
				if abstruct {
					text = article.GetAbstract()
				} else {
					text = article.GetText()
				}
				text = strings.Trim(text, "\n")
				text = re.ReplaceAllString(text, "\n")
				data["text"] = text

				if ns {
					data["ns"] = page.Ns
				}

				if redirect {
					data["redirect"] = page.Redir.Title
				}

				if textLinks {
					data["text_links"] = article.GetTextLinks()
				}

				if media {
					data["media"] = article.GetMedia()
				}

				if links {
					data["links"] = article.GetLinks()
				}

				if externalLinks {
					data["external_links"] = article.GetExternalLinks()
				}

				if categories {
					if db == nil {
						return fmt.Errorf("database is required")
					}

					categories, err := GetCategories(db, page.ID)
					if err != nil {
						return err
					}

					data["categories"] = categories
				}

				jsonBytes, err := json.Marshal(&data)
				if err != nil {
					return err
				}

				fmt.Println(string(jsonBytes))
			}

			return nil
		},
	}
)

func GetArticle(title string, text string) (*gowiki.Article, error) {
	return gowiki.ParseArticle(title, text, &gowiki.DummyPageGetter{})
}

func GetCategories(db *sql.DB, id uint64) ([]string, error) {
	query := fmt.Sprintf(`SELECT cl_to AS category FROM categorylinks WHERE cl_from = %d`, id)

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	categories := make([]string, 0)
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	return categories, nil
}

func Execute() error {
	if err := wikipediaJsonlCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	return nil
}

func init() {
	wikipediaJsonlCmd.Flags().BoolP("version", "v", false, "show version")
	wikipediaJsonlCmd.Flags().StringP("database-path", "d", "", "Wikipedia database path.")
	wikipediaJsonlCmd.Flags().BoolP("abstruct", "a", false, "use the abstruct as text")
	wikipediaJsonlCmd.Flags().BoolP("ns", "n", false, "include ns")
	wikipediaJsonlCmd.Flags().BoolP("redirect", "r", false, "include redirect")
	wikipediaJsonlCmd.Flags().BoolP("text-links", "t", false, "include links with text information")
	wikipediaJsonlCmd.Flags().BoolP("media", "m", false, "include media")
	wikipediaJsonlCmd.Flags().BoolP("links", "l", false, "include links")
	wikipediaJsonlCmd.Flags().BoolP("external-links", "e", false, "include external links")
	wikipediaJsonlCmd.Flags().BoolP("categories", "c", false, "include categories")
}
