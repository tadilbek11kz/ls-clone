package main

import "C"

import (
	"fmt"
	"os"
	"strings"
)

type List struct {
	permissions string
	hardLinks   string
	owner       string
	group       string
	size        string
	epochNano   int64
	month       string
	day         string
	time        string
	name        string
	linkName    string
	linkColor   string
	major       string
	minor       string
	linkOrphan  bool
	isSocket    bool
	isPipe      bool
	isBlock     bool
	isCharacter bool
}

type Dir struct {
	name      string
	epochNano int64
	size      int64
}

type Options struct {
	all         bool
	long        bool
	human       bool
	one         bool
	dir         bool
	color       bool
	sortReverse bool
	sortTime    bool
	sortSize    bool
	help        bool
	dirsFirst   bool
	recursive   bool
}

type FileInfoPath struct {
	path string
	info os.FileInfo
}

func ls(lsOutput *[]string, files []string) error {
	var output []string
	var dirsList []List
	var filesList []List
	var size int

	if len(files) == 0 {
		currentDir, err := os.Lstat(".")
		if err != nil && os.IsPermission(err) {
			lsOutputLen := len(*lsOutput)
			if lsOutputLen == 0 {
				*lsOutput = append(*lsOutput, fmt.Sprintf("ls: cannot open directory .: Permission denied"))
			} else {
				(*lsOutput)[lsOutputLen-1] = fmt.Sprintf("%v\nls: cannot open directory .: Permission denied", (*lsOutput)[lsOutputLen-1])
			}
			return err
		} else if err != nil {
			return err
		}
		dirList, _, err := CreateList(".", FileInfoPath{".", currentDir})
		if err != nil && os.IsPermission(err) {
			lsOutputLen := len(*lsOutput)
			if lsOutputLen == 0 {
				*lsOutput = append(*lsOutput, fmt.Sprintf("ls: cannot open directory .: Permission denied"))
			} else {
				(*lsOutput)[lsOutputLen-1] = fmt.Sprintf("%v\nls: cannot open directory .: Permission denied", (*lsOutput)[lsOutputLen-1])
			}
			return err
		} else if err != nil {
			return err
		}
		if options.dir {
			filesList = append(filesList, dirList)
		} else {
			dirsList = append(dirsList, dirList)
		}
	}

	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil && os.IsNotExist(err) {
			return fmt.Errorf("cannot access %s: no such file or directory", f)
		} else if err != nil {
			lsOutputLen := len(*lsOutput)
			if lsOutputLen == 0 {
				*lsOutput = append(*lsOutput, "ls: "+err.Error())
			} else {
				(*lsOutput)[lsOutputLen-1] = fmt.Sprintf("%v\n%s", (*lsOutput)[lsOutputLen-1], err.Error())
			}
			continue
		}
		path := f
		infoPath := f
		if !info.IsDir() {
			splitedPath := strings.Split(path, "/")
			path = strings.Join(splitedPath[:len(splitedPath)-1], "/")
		}
		fileList, blocksize, err := CreateList(path, FileInfoPath{infoPath, info})
		if err != nil {
			lsOutputLen := len(*lsOutput)
			if lsOutputLen == 0 {
				*lsOutput = append(*lsOutput, "ls: "+err.Error())
			} else {
				(*lsOutput)[lsOutputLen-1] = fmt.Sprintf("%v\n%s", (*lsOutput)[lsOutputLen-1], err.Error())
			}
			continue
		}

		if options.dir {
			filesList = append(filesList, fileList)
		} else {
			if info.IsDir() {
				dirsList = append(dirsList, fileList)
			} else {
				filesList = append(filesList, fileList)
				size += blocksize
			}
		}
	}

	filesNum := len(filesList)
	dirsNum := len(dirsList)
	SortList(filesList)
	SortList(dirsList)

	if filesNum > 0 {
		toWrite := WriteListToOuptut(filesList, terminalWidth)
		if len(files) > 1 {
			toWrite += "\n"
		}
		if len(toWrite) > 0 {
			output = append(output, toWrite)
		}
		size = 0
	}

	if (filesNum > 0 && dirsNum > 0) || (dirsNum > 1) {
		for index, d := range dirsList {
			if index == 0 {
				output = append(output, fmt.Sprintf("%v:", d.name))
			} else {
				output = append(output, fmt.Sprintf("\n%v:", d.name))
			}

			listings, blocksize, err := ListDirFiles(d)
			size += blocksize
			if options.long {
				output = append(output, fmt.Sprintf("total %v", size))
			}
			size = 0
			if err != nil {
				lsOutputLen := len(*lsOutput)
				if lsOutputLen == 0 {
					*lsOutput = append(*lsOutput, "ls: "+err.Error())
				} else {
					(*lsOutput)[lsOutputLen-1] = fmt.Sprintf("%v\n%s", (*lsOutput)[lsOutputLen-1], err.Error())
				}
				continue
			}

			if options.dirsFirst {
				listings = SortDirsFirst(listings)
			}

			if len(listings) > 0 {
				toWrite := WriteListToOuptut(listings, terminalWidth)
				if len(toWrite) > 0 {
					output = append(output, toWrite)
				}
			}
		}
	} else if dirsNum == 1 {
		for _, d := range dirsList {

			listings, blocksize, err := ListDirFiles(d)
			size += blocksize
			if err != nil {
				lsOutputLen := len(*lsOutput)
				if lsOutputLen == 0 {
					*lsOutput = append(*lsOutput, "ls: "+err.Error())
				} else {
					(*lsOutput)[lsOutputLen-1] = fmt.Sprintf("%v\n%s", (*lsOutput)[lsOutputLen-1], err.Error())
				}
				continue
			}
			if options.recursive {
				output = append(output, fmt.Sprintf("%v:", d.name))
			}
			if options.dirsFirst {
				listings = SortDirsFirst(listings)
			}
			if options.long {
				output = append(output, fmt.Sprintf("total %v", size))
			}
			size = 0
			toWrite := WriteListToOuptut(listings, terminalWidth)
			if len(toWrite) > 0 {
				output = append(output, toWrite)
			}
		}
	}

	//
	// list the files now if --dirs-first
	//

	if output != nil {
		*lsOutput = append(*lsOutput, strings.Join(output, "\n"))
	}
	return nil
}

func recursion(output *[]string, files []string) error {
	var list []os.FileInfo
	for _, dir := range files {
		err := ls(output, []string{dir})
		if err != nil && !os.IsPermission(err) {
			return err
		}
		path := dir
		var dirs []List
		fi, err := os.Stat(dir)
		if err != nil && !os.IsPermission(err) {
			return err
		}
		if fi.Mode().IsDir() {
			list, err = ReadDir(dir)
		}
		if err != nil && !os.IsPermission(err) {
			return err
		}
		for _, dirInfo := range list {
			var dirTemp List
			invisibleCheck := string(dirInfo.Name()[0]) != "."
			if options.all {
				invisibleCheck = true
			}
			if dirInfo.IsDir() && invisibleCheck {
				for path[len(path)-1] == '/' {
					path = path[:len(path)-1]
				}
				dirTemp.name = path + "/" + dirInfo.Name()
				dirTemp.size = fmt.Sprintf("%d", dirInfo.Size())
				dirTemp.epochNano = dirInfo.ModTime().UnixNano()
				dirs = append(dirs, dirTemp)
			}
		}
		if len(dirs) > 0 {
			sortedlist := BubbleSort(dirs, options.sortReverse)
			err = recursion(output, sortedlist)
		}
		if err != nil && !os.IsPermission(err) {
			return err
		}
	}

	return nil
}

func main() {
	var err error
	args := os.Args[1:]
	var flags []string
	var files []string
	var output string
	var result []string
	for _, arg := range args {
		if len(arg) > 1 && []rune(arg)[0] == '-' {
			// add to the options list
			flags = append(flags, arg)
		} else {
			// add to the files/directories list
			files = append(files, arg)
		}
	}

	options = ParseOptions(flags)

	if options.help {
		help := "usage:  ls [OPTIONS] [FILES]\n\n" +
			"OPTIONS:\n" +
			"    --dirs-first  list directories first\n" +
			"    --help        display usage information\n" +
			"    --nocolor     remove color formatting\n" +
			"    -1            one entry per line\n" +
			"    -a            include entries starting with '.'\n" +
			"    -d            list directories like files\n" +
			"    -h            list sizes with human-readable units\n" +
			"    -l            long listing\n" +
			"    -r            reverse any sorting\n" +
			"    -t            sort entries by modify time\n" +
			"    -S            sort entries by size\n"
		fmt.Println(help)
		return
	}

	if options.color {
		colorsMap = ParseColors()
	}
	if !options.recursive {
		var tmp []string
		err = ls(&tmp, files)
		output = strings.Join(tmp, "\n")
	} else {
		if len(files) == 0 {
			files = append(files, ".")
		}
		err = recursion(&result, files)
	}

	if err != nil {
		fmt.Printf("ls: %v\n", err.Error())
		os.Exit(1)
	}
	if options.recursive {
		fmt.Println(strings.Join(result, "\n\n"))
	} else {
		fmt.Println(output)
	}
}
