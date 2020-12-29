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
}

func main() {
	logDescription()

	mainDirectory := "G:\\Musik\\0 - Restmusik"
	// TODO: Comment back in if all is done
	// mainDirectory := promptForDirectory()

	sortSongsByTitle := true
	// TODO: Comment back in if all is done
	// var sortSongsByTitle bool = prompter.YN("Should we sort the list by title? If you say no, we will sort by artist.", true)

	format := "csv"
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
}
