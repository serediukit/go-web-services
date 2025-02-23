package main

import (
	"fmt"
	"io"
	"os"
	"sort"
)

func dirTree(out io.Writer, path string, printFiles bool) error {
	return dirTreeLevel(out, path, printFiles, []bool{})
}

func dirTreeLevel(out io.Writer, path string, printFiles bool, prevPaths []bool) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	filesDir, err := file.ReadDir(0)
	if err != nil {
		return err
	}

	sort.Slice(filesDir, func(i, j int) bool {
		return filesDir[i].Name() < filesDir[j].Name()
	})

	lastIdx := getLastIdx(filesDir, printFiles)

	for idx, fileDir := range filesDir {
		isLast := idx == lastIdx
		info, err := fileDir.Info()
		if err != nil {
			return err
		}
		if fileDir.IsDir() {
			printDir(out, info, isLast, prevPaths)
			err := dirTreeLevel(out, path+"/"+fileDir.Name(), printFiles, append(prevPaths, isLast))
			if err != nil {
				return err
			}
		} else if printFiles {
			printFile(out, info, isLast, prevPaths)
		}
	}

	return nil
}

func getLastIdx(files []os.DirEntry, printFiles bool) int {
	idx := len(files) - 1
	if !printFiles {
		for !files[idx].IsDir() && idx > 0 {
			idx--
		}
	}
	return idx
}

func printDir(out io.Writer, fileInfo os.FileInfo, isLast bool, prevPaths []bool) {
	path := getPath(isLast, prevPaths)
	_, err := out.Write([]byte(fmt.Sprintf("%s%s\n", path, fileInfo.Name())))
	if err != nil {
		return
	}
}

func printFile(out io.Writer, fileInfo os.FileInfo, isLast bool, prevPaths []bool) {
	path := getPath(isLast, prevPaths)

	size := "empty"

	if fileInfo.Size() > 0 {
		size = fmt.Sprintf("%db", fileInfo.Size())
	}

	_, err := out.Write([]byte(fmt.Sprintf("%s%s (%s)\n", path, fileInfo.Name(), size)))
	if err != nil {
		return
	}
}

func getPath(isLast bool, prevPaths []bool) string {
	path := ""

	for _, prevPath := range prevPaths {
		if !prevPath {
			path += "│\t"
		} else {
			path += "\t"
		}
	}

	pathEnd := "├───"
	if isLast {
		pathEnd = "└───"
	}

	return path + pathEnd
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
