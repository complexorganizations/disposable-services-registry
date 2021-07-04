package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime/debug"
	"sort"
	"strings"
	"sync"

	"github.com/openrdap/rdap"
	"golang.org/x/net/publicsuffix"
)

var (
	// Location of the configuration in the local system path
	disposableDomains          = "assets/disposable-domains"
	exclusionsDomains          = "assets/exclusions-domains"
	disposableTelephoneNumbers = "assets/disposable-telephone-numbers"
	exclusionsTelephoneNumbers = "assets/exclusions-telephone-numbers"
	// Memorandum with a domain list.
	disposableDomainsArray          []string
	exclusionsDomainsArray          []string
	disposableTelephoneNumbersArray []string
	exclusionsTelephoneNumbersArray []string
	// Go routines using waitgrops.
	scrapeWaitGroup     sync.WaitGroup
	validationWaitGroup sync.WaitGroup
	// The user expresses his or her opinion on what should be done.
	update   bool
	// err stands for error.
	err error
)

func init() {
	// If any user input flags are provided, use them.
	if len(os.Args) > 1 {
		tempUpdate := flag.Bool("update", false, "Make any necessary changes to the listings.")
		flag.Parse()
		update = *tempUpdate
	} else {
		os.Exit(0)
	}
}

func main() {
	// Lists should be updated.
	if update {
		// Clear your memories as much as possible
		os.RemoveAll(os.TempDir())
		os.Mkdir(os.TempDir(), 0777)
		debug.FreeOSMemory()
		// Max ammount of go routines
		debug.SetMaxThreads(10000)
		// Remove the old files from your system if they are found.
		err = os.Remove(disposableDomains)
		if err != nil {
			log.Println(err)
		}
		err = os.Remove(disposableTelephoneNumbers)
		if err != nil {
			log.Println(err)
		}
		// Read through all of the exclusion domains before appending them.
		if fileExists(exclusionsDomains) {
			exclusionsDomainsArray = readAndAppend(exclusionsDomains, exclusionsDomainsArray)
		}
		if fileExists(exclusionsTelephoneNumbers) {
			exclusionsTelephoneNumbersArray = readAndAppend(exclusionsTelephoneNumbers, exclusionsTelephoneNumbersArray)
		}
		// Scrape all of the domains and save them afterwards.
		startScraping()
		// We'll make everything distinctive once everything is finished.
		makeEverythingUnique(disposableDomains)
		makeEverythingUnique(disposableTelephoneNumbers)
	}
}

// Replace the URLs in this section to create your own list or add new lists.
func startScraping() {
	// Disposable Domains
	domainsLists := []string{
		"https://raw.githubusercontent.com/ivolo/disposable-email-domains/master/index.json",
		"https://raw.githubusercontent.com/ivolo/disposable-email-domains/master/wildcard.json",
		"https://raw.githubusercontent.com/martenson/disposable-email-domains/master/disposable_email_blocklist.conf",
		"https://raw.githubusercontent.com/packetstream/disposable-email-domains/master/emails.txt",
		"https://raw.githubusercontent.com/andreis/disposable-email-domains/master/domains.txt",
		"https://raw.githubusercontent.com/di/disposable-email-domains/master/source_data/disposable_email_blocklist.conf",
		"https://raw.githubusercontent.com/wesbos/burner-email-providers/master/emails.txt",
		"https://raw.githubusercontent.com/groundcat/disposable-email-domain-list/master/domains.txt",
		"https://raw.githubusercontent.com/abimaelmartell/goverify/master/list.txt",
		"https://raw.githubusercontent.com/maxmalysh/disposable-emails/master/disposable_emails/data/domains.txt",
		"https://raw.githubusercontent.com/Xyborg/disposable-burner-email-providers/master/disposable-domains.txt",
		"https://raw.githubusercontent.com/pidario/disposable/master/list/index.json",
		"https://raw.githubusercontent.com/ivolo/disposable-email-domains/master/index.json",
		"https://raw.githubusercontent.com/amieiro/disposable-email-domains/master/denyDomains.txt",
	}
	// Phone Numbers
	phoneNumberList := []string{
		"https://raw.githubusercontent.com/amieiro/disposable-email-domains/master/denyDomains.txt",
	}
	// Let's start by making everything one-of-a-kind so we don't scrape the same thing twice.
	uniqueDomainsLists := makeUnique(domainsLists)
	domainsLists = nil
	uniquePhoneNumberList := makeUnique(phoneNumberList)
	domainsLists = nil
	// Disposable Domains
	for _, content := range uniqueDomainsLists {
		if validURL(content) {
			scrapeWaitGroup.Add(1)
			go scrapeContent(content, disposableDomains, disposableDomainsArray)
		}
	}
	// Phone Numbers
	for _, content := range uniquePhoneNumberList {
		if validURL(content) {
			scrapeWaitGroup.Add(1)
			go scrapeContent(content, disposableTelephoneNumbers, disposableTelephoneNumbersArray)
		}
	}
	// Clear the memory via force.
	debug.FreeOSMemory()
	// We'll just wait for it to finish as a group.
	scrapeWaitGroup.Wait()
}

func scrapeContent(url string, saveLocation string, returnContent []string) {
	// Send a request to acquire all the information you need.
	response, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
	// read all the content of the body.
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
	}
	// Examine the page's response code.
	if response.StatusCode == 404 {
		log.Println("Sorry, but we were unable to scrape the page you requested due to a 404 error.", url)
	}
	// Scraped data is read and appended to an array.
	scanner := bufio.NewScanner(bytes.NewReader(body))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		returnContent = append(returnContent, scanner.Text())
	}
	// When you're finished, close the body.
	response.Body.Close()
	for _, content := range returnContent {
		// Make sure the domain is at least 3 characters long
		if len(content) > 1 {
			// This is a list of all the domains discovered using the regex.
			foundDomains := regexp.MustCompile(`(?:[a-z0-9_](?:[a-z0-9_-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]`).Find([]byte(content))
			// all the emails from rejex
			foundDomain := string(foundDomains)
			if len(foundDomain) > 3 {
				// Validate the entire list of domains.
				if len(foundDomain) < 255 && checkIPAddress(foundDomain) && !strings.Contains(foundDomain, " ") && strings.Contains(foundDomain, ".") && !strings.Contains(foundDomain, "#") && !strings.Contains(foundDomain, "*") && !strings.Contains(foundDomain, "!") {
					// icann.org confirms it's a public suffix domain
					eTLD, icann := publicsuffix.PublicSuffix(foundDomain)
					// Start the other tests if the domain has a valid suffix.
					if icann || strings.IndexByte(eTLD, '.') >= 0 {
						validationWaitGroup.Add(1)
						// Go ahead and verify it in the background.
						go validateTheDomains(foundDomain, saveLocation)
					}
				}
			}
		}
	}
	debug.FreeOSMemory()
	scrapeWaitGroup.Done()
	// While the validation is being performed, we wait.
	validationWaitGroup.Wait()
}

func validateTheDomains(uniqueDomain string, locatioToSave string) {
	// Validate each and every found domain.
	if validateDomainViaLookupNS(uniqueDomain) || validateDomainViaLookupAddr(uniqueDomain) || validateDomainViaLookupIP(uniqueDomain) || validateDomainViaLookupCNAME(uniqueDomain) || validateDomainViaLookupMX(uniqueDomain) || validateDomainViaLookupTXT(uniqueDomain) || validateDomainViaLookupHost(uniqueDomain) || domainRegistration(uniqueDomain) || validateDomainViaHTTP(uniqueDomain) || validateDomainViaHTTPS(uniqueDomain) || validateApplicationViaHTTP(uniqueDomain) || validateApplicationViaHTTPS(uniqueDomain) {
		// Maintain a list of all authorized domains.
		writeToFile(locatioToSave, uniqueDomain)
	}
	// When it's finished, we'll be able to inform waitgroup that it's finished.
	validationWaitGroup.Done()
}

// Take a list of domains and make them one-of-a-kind
func makeUnique(randomStrings []string) []string {
	flag := make(map[string]bool)
	var uniqueString []string
	for _, content := range randomStrings {
		if !flag[content] {
			flag[content] = true
			uniqueString = append(uniqueString, content)
		}
	}
	return uniqueString
}

// Using name servers, verify the domain.
func validateDomainViaLookupNS(domain string) bool {
	valid, _ := net.LookupNS(domain)
	return len(valid) >= 1
}

// Using ip address, verify the domain.
func validateDomainViaLookupIP(domain string) bool {
	valid, _ := net.LookupIP(domain)
	return len(valid) >= 1
}

// Using a lookup address, verify the domain.
func validateDomainViaLookupAddr(domain string) bool {
	valid, _ := net.LookupAddr(domain)
	return len(valid) >= 1
}

// Using cname, verify the domain.
func validateDomainViaLookupCNAME(domain string) bool {
	valid, _ := net.LookupCNAME(domain)
	return len(valid) >= 1
}

// mx records are used to validate the domain.
func validateDomainViaLookupMX(domain string) bool {
	valid, _ := net.LookupMX(domain)
	return len(valid) >= 1
}

// Using txt records, validate the domain.
func validateDomainViaLookupTXT(domain string) bool {
	valid, _ := net.LookupTXT(domain)
	return len(valid) >= 1
}

// Ping the server using http to see if anything is there.
func validateDomainViaHTTP(domain string) bool {
	pingThis := fmt.Sprint(domain + ":" + "80")
	_, err := net.Dial("tcp", pingThis)
	return err == nil
}

// Using https, ping the server and see if anything is there.
func validateDomainViaHTTPS(domain string) bool {
	pingThis := fmt.Sprint(domain + ":" + "443")
	_, err := net.Dial("tcp", pingThis)
	return err == nil
}

// To check if the website is up and functioning, send an HTTP request to it.
func validateApplicationViaHTTP(domain string) bool {
	httpValue := fmt.Sprint("http://" + domain)
	_, err := http.Get(httpValue)
	return err == nil
}

// Send a request to see if the program is running.
func validateApplicationViaHTTPS(domain string) bool {
	httpValue := fmt.Sprint("https://" + domain)
	_, err := http.Get(httpValue)
	return err == nil
}

// Using host, see if the domain is legitimate.
func validateDomainViaLookupHost(domain string) bool {
	valid, _ := net.LookupHost(domain)
	return len(valid) >= 1
}

// Validate the domain by checking the domain registration.
func domainRegistration(domain string) bool {
	client := &rdap.Client{}
	_, ok := client.QueryDomain(domain)
	return ok == nil
}

// Make sure it's not an IP address.
func checkIPAddress(ip string) bool {
	return net.ParseIP(ip) == nil
}

// Verify the URI.
func validURL(uri string) bool {
	_, err = url.ParseRequestURI(uri)
	return err == nil
}

// Check to see if a file already exists.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Remove a string from a slice
func removeStringFromSlice(originalSlice []string, removeString string) []string {
	// go though the array
	for i, content := range originalSlice {
		// if the array matches with the string, you remove it from the array
		if content == removeString {
			return append(originalSlice[:i], originalSlice[i+1:]...)
		}
	}
	return originalSlice
}

// Save the information to a file.
func writeToFile(pathInSystem string, content string) {
	// open the file and if its not there create one.
	filePath, err := os.OpenFile(pathInSystem, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	// write the content to the file
	_, err = filePath.WriteString(content + "\n")
	if err != nil {
		log.Println(err)
	}
	// close the file
	defer filePath.Close()
}

// Read and append to array
func readAndAppend(fileLocation string, arrayName []string) []string {
	file, err := os.Open(fileLocation)
	if err != nil {
		log.Println(err)
	}
	// scan the file, and read the file
	scanner := bufio.NewScanner(file)
	// split each line
	scanner.Split(bufio.ScanLines)
	// append each line to array
	for scanner.Scan() {
		arrayName = append(arrayName, scanner.Text())
	}
	// close the file before func ends
	defer file.Close()
	return arrayName
}

// Read the completed file, then delete any duplicates before saving it.
func makeEverythingUnique(contentLocation string) {
	var finalDomainList []string
	finalDomainList = readAndAppend(contentLocation, finalDomainList)
	// Make each domain one-of-a-kind.
	uniqueDomains := makeUnique(finalDomainList)
	// It is recommended that the array be deleted from memory.
	finalDomainList = nil
	// Sort the entire string.
	sort.Strings(uniqueDomains)
	// Remove all the exclusions domains from the list.
	for _, content := range exclusionsDomainsArray {
		uniqueDomains = removeStringFromSlice(uniqueDomains, content)
	}
	// Delete the original file and rewrite it.
	err = os.Remove(contentLocation)
	if err != nil {
		log.Println(err)
	}
	// Begin composing the document
	for _, content := range uniqueDomains {
		writeToFile(contentLocation, content)
	}
	// remove it from memory
	uniqueDomains = nil
	debug.FreeOSMemory()
}
