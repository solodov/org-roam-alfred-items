/*
Copyright © 2023 Peter Solodov <solodov@gmail.com>
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/solodov/org-roam-alfred-items/alfred"
	"github.com/spf13/cobra"
)

type fromToPair struct {
	from, to string
}

func translitConversionTable() []fromToPair {
	return []fromToPair{
		{"tvz", "ъ"},
		{"shh", "щ"},
		{"mjz", "ь"},
		{"Shh", "Щ"},
		{"zh", "ж"},
		{"yu", "ю"},
		{"yo", "ё"},
		{"ya", "я"},
		{"sh", "ш"},
		{"ju", "ю"},
		{"jo", "ё"},
		{"je", "э"},
		{"ja", "я"},
		{"ch", "ч"},
		{"Zh", "Ж"},
		{"Sh", "Ш"},
		{"Ju", "Ю"},
		{"Jo", "Ё"},
		{"Je", "Э"},
		{"Ja", "Я"},
		{"Ch", "Ч"},
		{"''", "Ь"},
		{"##", "Ъ"},
		{"z", "з"},
		{"y", "ы"},
		{"x", "х"},
		{"w", "щ"},
		{"v", "в"},
		{"u", "у"},
		{"t", "т"},
		{"s", "с"},
		{"r", "р"},
		{"q", "я"},
		{"p", "п"},
		{"o", "о"},
		{"n", "н"},
		{"m", "м"},
		{"l", "л"},
		{"k", "к"},
		{"j", "й"},
		{"i", "и"},
		{"h", "х"},
		{"g", "г"},
		{"f", "ф"},
		{"e", "е"},
		{"d", "д"},
		{"c", "ц"},
		{"b", "б"},
		{"a", "а"},
		{"Z", "З"},
		{"Y", "Ы"},
		{"V", "В"},
		{"U", "У"},
		{"T", "Т"},
		{"S", "С"},
		{"R", "Р"},
		{"P", "П"},
		{"O", "О"},
		{"N", "Н"},
		{"M", "М"},
		{"L", "Л"},
		{"K", "К"},
		{"J", "Й"},
		{"I", "И"},
		{"H", "Х"},
		{"G", "Г"},
		{"F", "Ф"},
		{"E", "Е"},
		{"D", "Д"},
		{"C", "Ц"},
		{"B", "Б"},
		{"A", "А"},
		{"'", "ь"},
		{"#", "ъ"},
	}
}

var translitCmd = &cobra.Command{
	Use:   "translit",
	Short: "Convert Latin-transliterated Russian into Cyrillic Russian",
	Run: func(cmd *cobra.Command, args []string) {
		var b strings.Builder
		table := translitConversionTable()
		for _, arg := range args {
			pos := 0
		found:
			for pos < len(arg) {
				for _, pair := range table {
					if strings.HasPrefix(arg[pos:], pair.from) {
						fmt.Fprint(&b, pair.to)
						pos += len(pair.from)
						continue found
					}
				}
				fmt.Fprint(&b, arg[pos:pos+1])
				pos += 1
			}
		}
		translation := b.String()
		printJson(alfred.Result{Items: []alfred.Item{alfred.Item{
			Title: translation,
			Text:  alfred.Text{Copy: translation, LargeType: translation},
			Arg:   translation,
		}}})
	},
}

func init() {
	rootCmd.AddCommand(translitCmd)
}
