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

func addFilesToGlobalFilesVariable(directory string, files *[]string) {
    err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		// Ignore going over the same directory over and over again
		if (path != directory) {
			isDirectory, _ := IsDirectory(path)
			if isDirectory {
				addFilesToGlobalFilesVariable(path, files)
			} else {
				*files = append(*files, path)
			}
		}
        return nil
    })
    if err != nil {
        panic(err)
	}
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
		fmt.Println(file)
		isMP3, _ := regexp.MatchString(".mp3$", file)
		if (isMP3) {
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

func getAllFileNamesInDirectoryRecursively(directoryPath string) []string {
	var files []string
	addFilesToGlobalFilesVariable(directoryPath, &files)
	return files
}

func promptForDirectory() string {
	directory := prompter.Prompt("Please provide us with the path of the directory where all your music lies. I.e. C:\\Music", "")
	if directory == "" {
		log.Fatal("You need to enter a valid path")
	}
	return directory
}

func promptForFormat() string {
	format := prompter.Prompt("What type of format should the list be generated in? Choose between 'html', 'csv' and 'pdf'.", "")
	if format != "csv" && format != "pdf" && format != "html" {
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
	fmt.Println("SUCCESS = CSV file successfully created under %s", targetFile)
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
	fmt.Println("SUCCESS = HTML file successfully created under %s", targetFile)
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
</head>
<body>
	<div id="loadingMessage">Loading music library... Might take a while due to its size and cause display issues. A pretty style will soon kick in, don't worry.</div>
    <table id="songLibrary" class="display" style="width:100%" style="display:none">
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
	$('#songLibrary').DataTable();
	$('#loadingMessage').css("display", "none");
} );</script>
</html>


`

func main() {
	logDescription()

	// mainDirectory := "G:\\Musik"
	mainDirectory := "G:\\Musik\\0 - Restmusik"
	// TODO: Comment back in if all is done
	// mainDirectory := promptForDirectory()

	sortSongsByTitle := true
	// TODO: Comment back in if all is done
	// var sortSongsByTitle bool = prompter.YN("Should we sort the list by title? If you say no, we will sort by artist.", true)

	format := "html"
	outputDirectory := "C:\\Users\\turm\\Desktop\\Learning_Coding\\music-library-reader"
	// TODO: Comment back in if all is done
	// format := promptForFormat()
	// outputDirectory := promptForOutputDirectory()
	fmt.Println(format, outputDirectory)

	files := getAllFileNamesInDirectoryRecursively(mainDirectory)
	songs := collectSongsFromFileNames(files)
	sortSongs(songs, sortSongsByTitle)

	printAllSongs(songs)

	if format == "csv" {
		createCSV(songs, outputDirectory)
	}

	if format == "html" {
		createHTML(songs, outputDirectory)
	}
}
