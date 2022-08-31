# Music Catalogue Creator

This project has been created for the website https://www.greatamericansongbook.info/ .

The tool can be downloaded [under the releases page](https://github.com/rickschubert/music-library-reader/releases). Simply download the tool and then execute it - it will then walk you through all the required steps. (Please note that this build is for Windows as this is the operating system used by greatamericansongbook.info. Executables for other OS's can be built as well though, no problem.)

**Should you wish to use the tool on a project other than Great American Songbook, you probably want a different background color and pagination size. You could either update the program yourself or contact me to release a new version for you.**

# Great American Songbook specific implementations

Due to requirements by the Great American Songbook page, the background color is beige and the default pagination size is 8. You can control the pagination with the following line in the main.go file:

```js
	$('#songLibrary').DataTable({
        "dom": '<"top"lfp><t>',
        // Change the page length and length options here
        pageLength: 8,
        lengthMenu: [8, 10, 25, 50, 100]
    });
```

The background color can be changed with this control in the main.go file:

```js
	background-color:
```

# How to build
GOOS=windows go build -o music-library-reader.exe
