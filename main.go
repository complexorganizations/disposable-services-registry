package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"
)

var (
	client    = http.DefaultClient
	exclusion []string
)

const (
	downloadWorkers   = 2500
	processWorkers    = 5000
	fileOutputName    = "assets/disposable-domains.txt"
	exclusionsDomains = "assets/exclusions-domains.txt"
)

func init() {
	file, err := os.Open(exclusionsDomains)
	if err != nil {
		log.Fatal("failed to open", exclusionsDomains)
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		exclusion = append(exclusion, scanner.Text())
	}
	file.Close()
	sort.Slice(exclusion, func(i, j int) bool {
		return exclusion[i] <= exclusion[j]
	})
}

func main() {
	client.Timeout = 30 * time.Second
	emails := readEmails()
	_ = os.Remove(fileOutputName)
	dm := newDownloaderManager(downloadWorkers)
	pm := newProcessManager(processWorkers)
	fm := newfileWriterManager()
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

func readEmails() []string {
	out := make([]string, 0)
	file, err := os.OpenFile(fileOutputName, os.O_CREATE|os.O_RDONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return out
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		out = append(out, scanner.Text())
	}
	_ = file.Close()
	return out
}

type typeURL struct {
	URL  string
	Type string
}

type downloadManager struct {
	workers     int
	emailOutput chan string
}

func newDownloaderManager(workers int) *downloadManager {
	return &downloadManager{
		workers:     workers,
		emailOutput: make(chan string, 50),
	}
}

func (dm *downloadManager) Run(urls []typeURL) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	urlsInput := make(chan typeURL, 10)
	go func() {
		for _, typeURL := range urls {
			urlsInput <- typeURL
			wg.Add(1)
		}
		wg.Done()
	}()
	for i := 0; i < dm.workers; i++ {
		go func() {
			for typeURL := range urlsInput {
				var emails []string
				switch typeURL.Type {
				case "txt":
					emails = downloadTextEmails(typeURL.URL)
				case "json":
					emails = downloadJSONEmails(typeURL.URL)
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

func (dm *downloadManager) Output() chan string {
	return dm.emailOutput
}

func downloadTextEmails(url string) []string {
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

func downloadJSONEmails(url string) []string {
	resp, err := client.Get(url)
	if err != nil {
		return make([]string, 0)
	}
	out := make([]string, 0)
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return out
}

type processManager struct {
	workers int
	output  chan string
}

func newProcessManager(workers int) *processManager {
	return &processManager{
		workers: workers,
		output:  make(chan string, 50),
	}
}

func (pm *processManager) Run(input chan string) {
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
				if validateDomain(email) {
					pm.output <- email
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func (pm *processManager) Output() chan string {
	return pm.output
}

func validateDomain(domain string) bool {
	mx, _ := net.LookupMX(domain)
	if len(mx) == 0 {
		return true
	}
	ns, _ := net.LookupNS(domain)
	return len(ns) != 0
}

type fileWriterManager struct{}

func newfileWriterManager() *fileWriterManager {
	return &fileWriterManager{}
}

func (pm *fileWriterManager) Run(input chan string) {
	file, err := os.OpenFile(fileOutputName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		log.Fatalln(err.Error())
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for email := range input {
			if !found(email, exclusion) {
				_, err := file.WriteString(email + "\n")
				if err != nil {
					log.Println(email, err)
				}
			} else {
				log.Println("Found ", email)
			}
		}
		wg.Done()
	}()
	wg.Wait()
	_ = file.Close()
}

func found(x string, a []string) bool {
	i := sort.Search(len(a), func(i int) bool { return x <= a[i] })
	if i < len(a) && a[i] == x {
		return true
	}
	return false
}

var urls = []typeURL{
	{
		URL:  "https://raw.githubusercontent.com/ivolo/disposable-email-domains/master/index.json",
		Type: "json",
	},
	{
		URL:  "https://raw.githubusercontent.com/ivolo/disposable-email-domains/master/wildcard.json",
		Type: "json",
	},
	{
		URL:  "https://raw.githubusercontent.com/martenson/disposable-email-domains/master/disposable_email_blocklist.conf",
		Type: "txt",
	},
	{
		URL:  "https://raw.githubusercontent.com/packetstream/disposable-email-domains/master/emails.txt",
		Type: "txt",
	},
	{
		URL:  "https://raw.githubusercontent.com/andreis/disposable-email-domains/master/domains.txt",
		Type: "txt",
	},
	{
		URL:  "https://raw.githubusercontent.com/di/disposable-email-domains/master/source_data/disposable_email_blocklist.conf",
		Type: "txt",
	},
	{
		URL:  "https://raw.githubusercontent.com/wesbos/burner-email-providers/master/emails.txt",
		Type: "txt",
	},
	{
		URL:  "https://raw.githubusercontent.com/groundcat/disposable-email-domain-list/master/domains.txt",
		Type: "txt",
	},
	{
		URL:  "https://raw.githubusercontent.com/abimaelmartell/goverify/master/list.txt",
		Type: "txt",
	},
	{
		URL:  "https://raw.githubusercontent.com/maxmalysh/disposable-emails/master/disposable_emails/data/domains.txt",
		Type: "txt",
	},
	{
		URL:  "https://raw.githubusercontent.com/Xyborg/disposable-burner-email-providers/master/disposable-domains.txt",
		Type: "txt",
	},
	{
		URL:  "https://raw.githubusercontent.com/pidario/disposable/master/list/index.json",
		Type: "json",
	},
	{
		URL:  "https://raw.githubusercontent.com/ivolo/disposable-email-domains/master/index.json",
		Type: "json",
	},
	{
		URL:  "https://raw.githubusercontent.com/amieiro/disposable-email-domains/master/denyDomains.txt",
		Type: "txt",
	},
}
