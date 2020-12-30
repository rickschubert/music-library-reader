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
list can be output as PDF document, CSV file or even as HTML page.
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
			songs = append(songs, song)
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
		songsAsHtmlTableRows.WriteString(fmt.Sprintf(`<tr><td class="table-title">%s</td><td class="table-artist">%s</td><td class="table-album">%s</td></tr>`, song.Title, song.Artist, song.Album))
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

	<style>

#songLibrary {
  border-collapse: collapse;
  width: 100%;
  border: 1px solid #ddd;
  border-left: 0px;
  border-right: 0px;
  font-size: 16px;
}

#songLibrary th, #songLibrary td {
  text-align: left;
  padding: 12px;
}

#songLibrary tr {

  border-bottom: 1px solid #ddd;
}

#songLibrary tr.header, #songLibrary tr:hover {
  background-color: #91c220;
}

#title-heading {
	width: 40%;
}

#artist-heading, #album-heading {
	width: 30%;
}

body {
	background-color: #FFFFCC;
	font-family: sans-serif;
}
	</style>
</head>
<body>
<table id="songLibrary">
	<tr class="header">
		<th id="title-heading">Title<br><input id="title-filter" type="text" oninput="filterBy('title')"></input></th>
		<th id="artist-heading">Artist<br><input id="artist-filter" type="text" oninput="filterBy('artist')"></input></th>
		<th id="album-heading">Album<br><input id="album-filter" type="text" oninput="filterBy('album')"></input></th>
	</tr>
		__HERE_GO_THE_TABLE_ROWS__
</table>
</body>
<script>
	function filterBy(type) {
	  const input = document.getElementById(type + '-filter');
	  const filter = input.value.toUpperCase().trim();
	  const table = document.getElementById("songLibrary");
	  const tr = table.getElementsByTagName("tr");

	  for (i = 0; i < tr.length; i++) {
		const td = tr[i].getElementsByClassName('table-' + type)[0];
		if (td) {
		  const txtValue = td.textContent || td.innerText;
		  if (txtValue.toUpperCase().trim().indexOf(filter) > -1) {
			tr[i].style.display = "";
		  } else {
			tr[i].style.display = "none";
		  }
		}
	  }
	}
</script>
</html>

`

func main() {
	logDescription()

	mainDirectory := "G:\\Musik"
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
