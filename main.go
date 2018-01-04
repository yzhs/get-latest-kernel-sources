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

func run(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

// Check which version is in ~/src/linux
func getLocalVersion() semver.Version {
	if err := os.Chdir(os.ExpandEnv(SOURCE_DIR)); err != nil {
		panic(err)
	}

	cmd := exec.Command("make", "kernelversion")
	bytes, err := cmd.Output()
	if err != nil {
		version, _ := semver.Make("4.0.0")
		return version
	}

	localVersion, err := semver.Make(strings.TrimSpace(string(bytes)))
	if err != nil {
		panic(err)
	}

	return localVersion
}

// Look for the latest version on kernel.org and get the link for downloading
// the source archive for that version.
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

func backupConfig() {
	run("mv", ".config", os.ExpandEnv("$HOME/prj/linux_config/config"))
}

// Download and extract the new sources
func extractNewVersion(url string) {
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	curl := exec.Command("curl", "-s", url)
	tar := exec.Command("tar", "xhJ")

	var err error
	tar.Stdin, err = curl.StdoutPipe()
	if err != nil {
		panic(err)
	}
	if err = tar.Start(); err != nil {
		panic(err)
	}
	if err = curl.Run(); err != nil {
		panic(err)
	}
	if err = tar.Wait(); err != nil {
		panic(err)
	}
}

func replaceSourceDirectory(newVersion semver.Version) {
	name := "linux-" + newVersion.String()

	run("mv", "linux/.git", "linux/.config", name)
	run("trash-put", "-r", "linux")
	run("mv", name, "linux")
	if err := os.Chdir("linux"); err != nil {
		panic(err)
	}
}

func updateGitRepo(newVersion semver.Version) {
	run("git", "commit", "-am", "Add kernel "+newVersion.String())
}

func main() {
	localVersion := getLocalVersion()
	latestVersion, url := getLatestVersion()

	if localVersion.LT(latestVersion) {
		log.Println("The current local version is", localVersion, "while the latest version is", latestVersion)
		log.Println("Updating...")

		log.Println("Backing up config to ~/prj/linux_config")
		backupConfig()

		log.Println("Downloading the latest version")
		extractNewVersion(url)

		log.Println("Replacing the source directory with the latest version")
		replaceSourceDirectory(latestVersion)

		log.Println("Adding files to version control")
		updateGitRepo(latestVersion)
	} else {
		log.Printf("The local version (%v) is up to date\n", localVersion)
	}
}
