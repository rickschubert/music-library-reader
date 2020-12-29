package main

import (
	"fmt"
	"github.com/dhowden/tag"
	"os"
	"path/filepath"
	"regexp"
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

func main() {
	logDescription()

	mainDirectory := "G:\\Musik\\0 - Restmusik"

	files := getAllFileNamesInDirectoryRecursively(mainDirectory)
	songs := collectSongsFromFileNames(files)

	fmt.Println(songs)
}
