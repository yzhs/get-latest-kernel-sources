package main

import (
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/blang/semver"
)

const (
	SOURCE_DIR = "$HOME/src/linux"
	URL        = "https://kernel.org/"
)

func getLocalVersion() semver.Version {
	if err := os.Chdir(os.ExpandEnv(SOURCE_DIR)); err != nil {
		panic(err)
	}

	cmd := exec.Command("make", "kernelversion")
	bytes, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	localVersion, err := semver.Make(strings.TrimSpace(string(bytes)))
	if err != nil {
		panic(err)
	}

	return localVersion
}

func getLatestVersion() (semver.Version, string) {
	doc, err := goquery.NewDocument(URL)
	if err != nil {
		panic(err)
	}

	var latestVersion semver.Version
	var url string
	doc.Find("td#latest_link > a").Each(func(i int, s *goquery.Selection) {
		var exists bool
		url, exists = s.Attr("href")
		if !exists {
			panic("Could not find attribute href!")
		}

		latestVersion, err = semver.Make(s.Text())
		if err != nil {
			panic(err)
		}
	})
	return latestVersion, url
}

func cleanRepository() {
}

func extractNewVersion() {
}

func main() {
	localVersion := getLocalVersion()
	latestVersion, url := getLatestVersion()

	log.Println("The current local version is", localVersion)
	log.Println("The latest version is", latestVersion)

	if localVersion.LT(latestVersion) {
		log.Println("Updating to the newest version")

		log.Println("Cleaning the source directory")
		cleanRepository()

		log.Println("Downloading the latest version")
		extractNewVersion()
	}
}
