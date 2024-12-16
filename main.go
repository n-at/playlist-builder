package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

///////////////////////////////////////////////////////////////////////////////

type MusicFile struct {
	SrcPath  string
	DestPath string
	ExtInf   string
}

///////////////////////////////////////////////////////////////////////////////

func main() {
	srcFile := flag.String("src", "", "Source m3u / m3u8 playlist file")
	destPath := flag.String("dest", "", "Destination playlist path, defaults to source file name without extension")
	destPlaylistFile := flag.String("dest-playlist", "", "Destination m3u / m3u8 playlist file, defaults to source file name")
	flag.Parse()

	if len(*srcFile) == 0 {
		fmt.Println("Source file required")
		os.Exit(1)
	}
	if len(*destPath) == 0 {
		srcFileBase := nameWithoutExtension(*srcFile)
		destPath = &srcFileBase
	}
	if len(*destPlaylistFile) == 0 {
		srcFileBase := path.Base(*srcFile)
		destPlaylistFile = &srcFileBase
	}

	musicFiles, err := readM3U(*srcFile)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Found %d music files\n", len(musicFiles))

	if err := createDirIfNotExists(*destPath); err != nil {
		panic(err)
	}

	for _, file := range musicFiles {
		destPath := path.Join(*destPath, file.DestPath)
		if err := copyFile(file.SrcPath, destPath); err != nil {
			fmt.Printf("Error copying %s to %s: %s\n", file.SrcPath, destPath, err)
		}
	}

	destPlayListPath := path.Join(*destPath, *destPlaylistFile)
	if err := writeM3U(destPlayListPath, musicFiles); err != nil {
		panic(err)
	}
}

///////////////////////////////////////////////////////////////////////////////

func readM3U(file string) ([]MusicFile, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\r")

	var musicFiles []MusicFile
	musicFileIdx := 1

	for lineIdx := 0; lineIdx < len(lines); lineIdx++ {
		line := lines[lineIdx]

		if !strings.HasPrefix(line, "#EXTINF") {
			continue
		}

		lineIdx++

		srcPath := lines[lineIdx]
		destName := fmt.Sprintf("%05d.%s", musicFileIdx, path.Ext(srcPath))

		musicFiles = append(musicFiles, MusicFile{
			SrcPath:  srcPath,
			DestPath: destName,
			ExtInf:   line,
		})

		musicFileIdx++
	}

	return musicFiles, nil
}

func writeM3U(file string, musicFiles []MusicFile) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString("#EXTM3U\r")

	for _, file := range musicFiles {
		f.WriteString(file.ExtInf)
		f.WriteString("\r")
		f.WriteString(file.DestPath)
		f.WriteString("\r")
	}

	return nil
}

func createDirIfNotExists(dir string) error {
	stat, err := os.Stat(dir)

	if os.IsNotExist(err) {
		err := os.Mkdir(dir, 0755)
		if err != nil {
			return err
		}
	} else if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	return nil
}

func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)

	return err
}

func nameWithoutExtension(name string) string {
	base := path.Base(name)
	ext := path.Ext(base)

	if len(ext) > 0 {
		lenExt := len([]rune(ext))
		lenName := len([]rune(base))
		if lenExt == lenName-1 {
			//.file
			return base
		}
		base = string([]rune(base)[0 : lenName-lenExt])
	}

	return base
}
