package main

import (
	"fmt"
	"github.com/dhowden/tag"
	"github.com/Songmu/prompter"
	"github.com/bradfitz/slice"
	"os"
	"path/filepath"
	"regexp"
	"log"
	"encoding/csv"
	"path"
	"strings"
)

func logDescription() {
	fmt.Print(`
Music Library Compiler
======================

This tool creates a list of all the songs you have in a specific directory. The
list can be output as CSV file or even as HTML page.
We are currently only processing MP3 files. If you need other file types as well,
please let us know, we can then adjust the script.

Created by Rick Schubert / rickschubert@gmx.de / rickschubert.net

Please provide us with a bit of information before we can start:


`)
}

type Song struct {
	Artist string
	Title string
	Album string
}

func IsDirectory(path string) (bool, error) {
    fileInfo, err := os.Stat(path)
    if err != nil{
      return false, err
    }
    return fileInfo.IsDir(), err
}

func getSongData(filePath string) (Song, error) {
	dat, _ := os.Open(filePath)
	m, err := tag.ReadFrom(dat)

	var song Song
	if err != nil {
		return song, err
	} else {
		song = Song{
			Artist: m.Artist(),
			Title: m.Title(),
			Album: m.Album(),
		}
		return song, nil
	}
}

func collectSongsFromFileNames(fileNames []string) []Song {
	var songs []Song

    for _, file := range fileNames {
		isMP3, _ := regexp.MatchString(".mp3$", file)
		if (isMP3) {
			fmt.Println(file)
			song, err := getSongData(file)
			if (err != nil) {
				fmt.Println(fmt.Sprintf("ERROR reading tags from file: %s", file))
				continue
			}
			// Don't append empty songs
			if !(song.Album == "" && song.Artist == "" && song.Title == "") {
				songs = append(songs, song)
			}
		}
	}

	return songs
}

func getAllFileNamesInDirectoryRecursively(directoryPath string, ignoredDirectories []string) []string {
	var files []string
    err := filepath.Walk(directoryPath, func(path string, info os.FileInfo, err error) error {
		// Don't do anything for ignored directories
		var shouldBeIgnored bool
		for _, ignoredDirectory := range ignoredDirectories {
			if ignoredDirectory != "" {
				if strings.HasPrefix(strings.ToLower(path), strings.ToLower(ignoredDirectory)) {
					shouldBeIgnored = true
					break
				}
			}
		}
		if shouldBeIgnored {
			return nil
		}

		isDirectory, _ := IsDirectory(path)
		if !isDirectory {
			files = append(files, path)
		}
        return nil
    })
    if err != nil {
        panic(err)
	}
	return files
}

func promptForDirectory() string {
	directory := prompter.Prompt("Please provide us with the path of the directory where all your music lies. I.e. C:\\Music", "")
	if directory == "" {
		log.Fatal("You need to enter a valid path")
	}
	return directory
}

func promptForIgnoredDirectories() []string {
	directories := prompter.Prompt("Are there any directories which should be ignored? Please provide a comma separated list. If nothing should be ignored, just leave it empty.\nExample: C:\\Music\\ignored_one,C:\\Music\\ignored_two", "")
	directoriesSplit := strings.Split(directories, ",")
	return directoriesSplit
}

func promptForFormat() string {
	format := prompter.Prompt("What type of format should the list be generated in? Choose between 'html' and 'csv'.", "")
	if format != "csv" && format != "html" {
		log.Fatal("The format option you asked for is not supported.")
	}
	return format
}

func promptForOutputDirectory() string {
	directory := prompter.Prompt("To which directory should we write the output file?", "")
	if directory == "" {
		log.Fatal("You need to enter a valid path")
	}
	return directory
}

func finishMessage() {
	prompter.Prompt("\n\nThe script has finished. Press any key to close this window.", "")
}

// If byTitle is false, sorts songs by artist name
func sortSongs(songs []Song, byTitle bool) {
	slice.Sort(songs[:], func(i, j int) bool {
		if byTitle {
			return songs[i].Title < songs[j].Title
		} else {
			return songs[i].Artist < songs[j].Artist
		}
	})
}

func printAllSongs(songs []Song) {
	for _, song := range songs {
		fmt.Println(fmt.Sprintf("%s -- %s -- %s", song.Title, song.Artist, song.Album))
	}
}

func createCSV(songs []Song, outputDirectory string) {
	targetFile := path.Join(outputDirectory, "library.csv")
    file, err := os.Create(targetFile)
    if err != nil {
		panic(err)
	}
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    for _, song := range songs {
        err := writer.Write([]string{
			song.Title,
			song.Artist,
			song.Album,
		})
		if err != nil {
			log.Fatal("Unable to write to CSV file.")
		}
	}
	fmt.Println(fmt.Sprintf("\n\n=================\nSUCCESS = CSV file successfully created under %s", targetFile))
}

func createHTML(songs []Song, outputDirectory string) {
	targetFile := path.Join(outputDirectory, "library.html")
    file, err := os.Create(targetFile)
    if err != nil {
		panic(err)
	}
	defer file.Close()

    var songsAsHtmlTableRows strings.Builder
	for _, song := range songs {
		songsAsHtmlTableRows.WriteString(fmt.Sprintf(`<tr><td>%s</td><td>%s</td><td>%s</td></tr>`, song.Title, song.Artist, song.Album))
	}

	htmlStringToWrite := strings.Replace(htmlBase, "__HERE_GO_THE_TABLE_ROWS__", songsAsHtmlTableRows.String(), 1)
	file.WriteString(htmlStringToWrite)
	fmt.Println(fmt.Sprintf("\n\n=================\nSUCCESS = HTML file successfully created under %s", targetFile))
}

var htmlBase = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Great American Songbook</title>
    <link rel="stylesheet" type="text/css" href="https://cdn.datatables.net/1.10.22/css/jquery.dataTables.min.css">
    <script type="text/javascript" language="javascript" src="https://code.jquery.com/jquery-3.5.1.js"></script>
	<script type="text/javascript" language="javascript" src="https://cdn.datatables.net/1.10.22/js/jquery.dataTables.min.js"></script>
    <style>
#songLibrary {
 width:100%;
 display:none;

}

body {
    font-family: sans-serif;
}
    </style>
</head>
<body>
	<div id="loadingMessage">
		<img src="https://cdnjs.cloudflare.com/ajax/libs/galleriffic/2.0.1/css/loader.gif" />
		<p>Loading music library... Might take a while due to its size. Please bear with us.</p>
	</div>
	<table id="songLibrary" class="display">
        <thead>
            <tr>
                <th>Title</th>
                <th>Artist</th>
                <th>Album</th>
			</tr>
			</thead>
			<tbody>
			__HERE_GO_THE_TABLE_ROWS__

        </tbody>
    </table>
</body>
<script>$(document).ready(function() {
    $('#songLibrary').css("display", "table");
	// Put pagination on top
	$('#songLibrary').DataTable({"dom": '<"top"fp><t><"bottom"l>'});
	$('#loadingMessage').css("display", "none");
} );</script>
</html>


`

func main() {
	logDescription()

	// mainDirectory := "G:\\Musik"
	// mainDirectory := "G:\\Musik\\0 - Restmusik"
	mainDirectory := promptForDirectory()

	// sortSongsByTitle := true
	var sortSongsByTitle bool = prompter.YN("Should we sort the list by title? If you say no, we will sort by artist.", true)

	// ignoredDirectories := []string{"G:\\Musik\\0 - Restmusik\\2000 Punk"}
	ignoredDirectories := promptForIgnoredDirectories()

	// format := "html"
	format := promptForFormat()

	// outputDirectory := "C:\\Users\\turm\\Desktop\\Learning_Coding\\music-library-reader"
	outputDirectory := promptForOutputDirectory()

	files := getAllFileNamesInDirectoryRecursively(mainDirectory, ignoredDirectories)
	songs := collectSongsFromFileNames(files)
	sortSongs(songs, sortSongsByTitle)

	if format == "csv" {
		createCSV(songs, outputDirectory)
	}

	if format == "html" {
		createHTML(songs, outputDirectory)
	}

	finishMessage()
}
