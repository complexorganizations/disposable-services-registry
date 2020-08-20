package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

/* Tasks
- 1. Download all the list
- 2. Combine all the list into one.
- 3. Check and valid if the domain has a NS or a MX record;
- 4. If the domain has a NS or a MX record than its a valid domain and goes to the output.
- 5. Add checks to make sure its not added twice.
- 6. Remove all the original files; don't save anything other than a standard output.
*/

/*
"https://gist.githubusercontent.com/adamloving/4401361/raw/66688cf8ad890433b917f3230f44489aa90b03b7"
"https://gist.githubusercontent.com/michenriksen/8710649/raw/d42c080d62279b793f211f0caaffb22f1c980912"
"https://raw.githubusercontent.com/wesbos/burner-email-providers/master/emails.txt"
"https://raw.githubusercontent.com/andreis/disposable/master/blacklist.txt"
"https://raw.githubusercontent.com/GeroldSetz/emailondeck.com-domains/master/emailondeck.com_domains_from_bdea.cc.txt"
"https://raw.githubusercontent.com/andreis/disposable/master/whitelist.txt"
"https://raw.githubusercontent.com/andreis/disposable-email-domains/master/domains.txt"
"https://raw.githubusercontent.com/ivolo/disposable-email-domains/master/wildcard.json"
"https://raw.githubusercontent.com/ivolo/disposable-email-domains/master/index.json"
*/

const (
	DownloadWorkers = 8
	ProcessWorkers  = 4

	FileOutputName = "output.txt"
)

var (
	client = http.DefaultClient
)

func main() {
	client.Timeout = 30 * time.Second

	emails := ReadEmails()

	dm := NewDownloaderManager(DownloadWorkers)
	pm := NewProcessManager(ProcessWorkers)
	fm := NewFileWriterManager()

	var wg sync.WaitGroup
	wg.Add(4)
	chn := make(chan bool, 1)

	go func() {
		for _, email := range emails {
			dm.Output() <- email
		}
		chn <- true
		wg.Done()
	}()

	go func() {
		dm.Run(urls)
		_ = <-chn
		close(dm.Output())
		wg.Done()
	}()

	go func() {
		pm.Run(dm.Output())
		wg.Done()
		close(pm.Output())
	}()

	go func() {
		fm.Run(pm.Output())
		wg.Done()
	}()

	wg.Wait()
}

func ReadEmails() []string {
	out := make([]string, 0)

	file, err := os.OpenFile(FileOutputName, os.O_CREATE|os.O_RDONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return out
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		out = append(out, scanner.Text())
	}

	return out
}

type URLType struct {
	URL  string
	Type string
}

type DownloadManager struct {
	workers     int
	emailOutput chan string
}

func NewDownloaderManager(workers int) *DownloadManager {
	return &DownloadManager{
		workers:     workers,
		emailOutput: make(chan string, 50),
	}
}

func (dm *DownloadManager) Run(urls []URLType) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	urlsInput := make(chan URLType, 10)

	go func() {
		for _, urlType := range urls {
			urlsInput <- urlType
			wg.Add(1)
		}

		wg.Done()
	}()

	for i := 0; i < dm.workers; i++ {
		go func() {
			for urlType := range urlsInput {
				var emails []string

				switch urlType.Type {
				case "txt":
					emails = DownloadTextEmails(urlType.URL)
				case "json":
					emails = DownloadJsonEmails(urlType.URL)
				}

				for _, email := range emails {
					dm.emailOutput <- email
				}

				wg.Done()
			}
		}()
	}

	wg.Wait()
}

func (dm *DownloadManager) Output() chan string {
	return dm.emailOutput
}

func DownloadTextEmails(url string) []string {
	resp, err := client.Get(url)
	if err != nil {
		return make([]string, 0)
	}

	out := make([]string, 0)

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		out = append(out, scanner.Text())
	}

	_ = resp.Body.Close()

	return out
}

func DownloadJsonEmails(url string) []string {
	resp, err := client.Get(url)
	if err != nil {
		return make([]string, 0)
	}

	out := make([]string, 0)
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return out
}

type ProcessManager struct {
	workers int
	output  chan string
}

func NewProcessManager(workers int) *ProcessManager {
	return &ProcessManager{
		workers: workers,
		output:  make(chan string, 50),
	}
}

func (pm *ProcessManager) Run(input chan string) {
	wg := sync.WaitGroup{}
	wg.Add(pm.workers)
	var mu sync.Mutex

	visited := make(map[string]struct{})

	for i := 0; i < pm.workers; i++ {
		go func() {
			for email := range input {
				mu.Lock()
				_, ok := visited[email]
				if ok {
					mu.Unlock()
					continue
				}
				visited[email] = struct{}{}
				mu.Unlock()

				if ValidateDomain(email) {
					pm.output <- email
				}
			}

			wg.Done()
		}()
	}

	wg.Wait()
}

func (pm *ProcessManager) Output() chan string {
	return pm.output
}

func ValidateDomain(domain string) bool {
	mx, _ := net.LookupMX(domain)
	if len(mx) >= 0 {
		return true
	}

	ns, _ := net.LookupNS(domain)
	return len(ns) != 0
}

type FileWriterManager struct{}

func NewFileWriterManager() *FileWriterManager {
	return &FileWriterManager{}
}

func (pm *FileWriterManager) Run(input chan string) {
	file, err := os.OpenFile(FileOutputName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		log.Fatalln(err.Error())
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for email := range input {
			_, err := file.WriteString(email + "\n")
			if err != nil {
				log.Println(email, err)
			}
		}

		wg.Done()
	}()

	wg.Wait()
	_ = file.Close()
}

var urls = []URLType{
	{
		URL:  "https://gist.githubusercontent.com/adamloving/4401361/raw/66688cf8ad890433b917f3230f44489aa90b03b7",
		Type: "txt",
	},
	{
		URL:  "https://gist.githubusercontent.com/michenriksen/8710649/raw/d42c080d62279b793f211f0caaffb22f1c980912",
		Type: "txt",
	},
	{
		URL:  "https://raw.githubusercontent.com/wesbos/burner-email-providers/master/emails.txt",
		Type: "txt",
	},
	{
		URL:  "https://raw.githubusercontent.com/andreis/disposable/master/blacklist.txt",
		Type: "txt",
	},
	{
		URL:  "https://raw.githubusercontent.com/GeroldSetz/emailondeck.com-domains/master/emailondeck.com_domains_from_bdea.cc.txt",
		Type: "txt",
	},
	{
		URL:  "https://raw.githubusercontent.com/andreis/disposable/master/whitelist.txt",
		Type: "txt",
	},
	{
		URL:  "https://raw.githubusercontent.com/andreis/disposable-email-domains/master/domains.txt",
		Type: "txt",
	},
	{
		URL:  "https://raw.githubusercontent.com/ivolo/disposable-email-domains/master/wildcard.json",
		Type: "json",
	},
	{
		URL:  "https://raw.githubusercontent.com/ivolo/disposable-email-domains/master/index.json",
		Type: "json",
	},
}
