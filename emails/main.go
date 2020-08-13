package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

func main() {
	urls := []string{
		"https://gist.githubusercontent.com/adamloving/4401361/raw/66688cf8ad890433b917f3230f44489aa90b03b7",
		"https://gist.githubusercontent.com/michenriksen/8710649/raw/d42c080d62279b793f211f0caaffb22f1c980912",
		"https://gist.githubusercontent.com/smeinecke/78b229031cc885a776c8b84c56e1c5ee/raw/0b2200109d68537c588066d05bc70b6bbe1d312d/generator_email_hosts",
		"https://raw.githubusercontent.com/wesbos/burner-email-providers/master/emails.txt",
		"https://raw.githubusercontent.com/andreis/disposable/master/blacklist.txt",
		"https://raw.githubusercontent.com/GeroldSetz/emailondeck.com-domains/master/emailondeck.com_domains_from_bdea.cc.txt",
		"https://raw.githubusercontent.com/andreis/disposable/master/whitelist.txt",
		"https://raw.githubusercontent.com/andreis/disposable-email-domains/master/domains.txt",
	}

	var wg sync.WaitGroup

	wg.Add(len(urls))

	for _, url := range urls {
		go func(url string) {
			defer wg.Done()
			tokens := strings.Split(url, "/")
			fileName := tokens[len(tokens)-1]
			fmt.Println("Downloading", url, "to", fileName)

			output, err := os.Create(fileName)
			if err != nil {
				log.Fatal("Error while creating", fileName, "-", err)
			}
			defer output.Close()

			res, err := http.Get(url)
			if err != nil {
				log.Fatal("http get error: ", err)
			} else {
				defer res.Body.Close()
				_, err = io.Copy(output, res.Body)
				if err != nil {
					log.Fatal("Error while downloading", url, "-", err)
				} else {
					fmt.Println("Downloaded", fileName)
				}
			}
		}(url)
	}
	wg.Wait()
	fmt.Println("Done")
}
