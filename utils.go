package main

import (
	"fmt"
	"math"
	"os"
	"os/user"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return nil
	case syscall.Errno(0x23):
		return syscall.EAGAIN
	case syscall.Errno(0x16):
		return syscall.EINVAL
	case syscall.Errno(0x2):
		return syscall.ENOENT
	}
	return e
}

var (
	options   Options
	colorsMap map[string]string
	counter   int
)

const terminalWidth = 188

func ParseOptions(flags []string) Options {
	options := Options{}
	options.color = true
	for _, flag := range flags {
		if strings.Contains(flag, "--") {
			if strings.Contains(flag, "--dirs-first") {
				options.dirsFirst = true
			}
			if strings.Contains(flag, "--help") {
				options.help = true
			}
			if strings.Contains(flag, "--nocolor") {
				options.color = false
			}
		} else {
			if strings.Contains(flag, "1") {
				options.one = true
			}
			if strings.Contains(flag, "a") {
				options.all = true
			}
			if strings.Contains(flag, "d") {
				options.dir = true
			}
			if strings.Contains(flag, "h") {
				options.human = true
			}
			if strings.Contains(flag, "l") {
				options.long = true
			}
			if strings.Contains(flag, "r") {
				options.sortReverse = true
			}
			if strings.Contains(flag, "t") {
				options.sortTime = true
			}
			if strings.Contains(flag, "S") {
				options.sortSize = true
			}
			if strings.Contains(flag, "R") {
				options.recursive = true
			}
		}
	}
	return options
}

func ParseColors() map[string]string {
	colorsMap := make(map[string]string)
	colorsMap["end"] = "\x1b[0m"

	envColors := strings.Split(os.Getenv("LS_COLORS"), ":")
	for _, color := range envColors {
		if color == "" {
			continue
		}

		tmp := strings.Split(color, "=")
		colorID := fmt.Sprintf("\x1b[%sm", tmp[1])
		colorType := tmp[0]

		if colorType == "rs" {
			colorsMap["end"] = colorID
		} else if colorType == "di" {
			colorsMap["directory"] = colorID
		} else if colorType == "ln" {
			colorsMap["symlink"] = colorID
		} else if colorType == "mh" {
			colorsMap["multi_hardlink"] = colorID
		} else if colorType == "pi" {
			colorsMap["pipe"] = colorID
		} else if colorType == "so" {
			colorsMap["socket"] = colorID
		} else if colorType == "bd" {
			colorsMap["block"] = colorID
		} else if colorType == "cd" {
			colorsMap["character"] = colorID
		} else if colorType == "or" {
			colorsMap["link_orphan"] = colorID
		} else if colorType == "mi" {
			colorsMap["link_orphan_target"] = colorID
		} else if colorType == "su" {
			colorsMap["executable_suid"] = colorID
		} else if colorType == "sg" {
			colorsMap["executable_sgid"] = colorID
		} else if colorType == "tw" {
			colorsMap["directory_o+w_sticky"] = colorID
		} else if colorType == "ow" {
			colorsMap["directory_o+w"] = colorID
		} else if colorType == "st" {
			colorsMap["directory_sticky"] = colorID
		} else if colorType == "ex" {
			colorsMap["executable"] = colorID
		} else {
			colorsMap[colorType] = colorID
		}
	}
	return colorsMap
}

func CreateList(dirName string, pathInfo FileInfoPath) (List, int, error) {
	var list List
	list.permissions = pathInfo.info.Mode().String()
	if pathInfo.info.Mode()&os.ModeSymlink == os.ModeSymlink {
		// fmt.Println(dirName, pathInfo)
		list.permissions = strings.Replace(list.permissions, "L", "l", 1)
		var path string

		path = fmt.Sprintf("%s", dirName)

		//fmt.Println(path)
		if path == "" {
			path = "."
		}
		link, err := os.Readlink(fmt.Sprintf("%s/%s", path, pathInfo.path))
		if err != nil && !os.IsPermission(err) {
			return list, 0, err
		}
		// fmt.Println(strings.Join(splitedPath[:len(splitedPath)-1], "/"))
		// fmt.Println(dirName)
		list.linkName = link
		var linkPath string
		if len(dirName) == 0 {
			linkPath = fmt.Sprintf("%s", link)
		} else {
			if len(link) > 0 && link[0] == '/' {
				linkPath = fmt.Sprintf("%s", link)
			} else {
				linkPath = fmt.Sprintf("%s/%s", path, link)
			}
		}
		//fmt.Println(linkPath)
		list.linkColor, err = GetLinkColor(linkPath)
		if err != nil {
			return list, 0, err
		}
		_, err = os.Open(linkPath)
		if err != nil && !os.IsPermission(err) {
			if os.IsNotExist(err) {
				list.linkOrphan = true
			}
		}
	} else if list.permissions[0] == 'D' {
		list.permissions = list.permissions[1:]
	} else if list.permissions[0:2] == "ug" {
		list.permissions = strings.Replace(list.permissions, "ug", "-", 1)
		list.permissions = fmt.Sprintf("%ss%ss%s",
			list.permissions[0:3],
			list.permissions[4:6],
			list.permissions[7:])
	} else if list.permissions[0] == 'u' {
		list.permissions = strings.Replace(list.permissions, "u", "-", 1)
		list.permissions = fmt.Sprintf("%ss%s",
			list.permissions[0:3],
			list.permissions[4:])
	} else if list.permissions[0] == 'g' {
		list.permissions = strings.Replace(list.permissions, "g", "-", 1)
		list.permissions = fmt.Sprintf("%ss%s",
			list.permissions[0:6],
			list.permissions[7:])
	} else if list.permissions[0:2] == "dt" {
		list.permissions = strings.Replace(list.permissions, "dt", "d", 1)
		list.permissions = fmt.Sprintf("%st",
			list.permissions[0:len(list.permissions)-1])
	} else if list.permissions[0] == 'S' {
		list.permissions = "s" + list.permissions[1:]
	}

	sys := pathInfo.info.Sys()
	stat, ok := sys.(*syscall.Stat_t)
	//fmt.Printf("%#v\n", stat)
	//fmt.Println(os.FileMode(stat.Mode).Perm())

	//fmt.Println(sys)
	if !ok {
		return list, 0, fmt.Errorf("syscall failed")
	}

	hardLinksNum := uint64(stat.Nlink)
	list.hardLinks = fmt.Sprintf("%d", hardLinksNum)

	owner, err := user.LookupId(fmt.Sprintf("%d", stat.Uid))
	if err != nil {
		return list, 0, err
	}
	list.owner = owner.Username

	group, err := user.LookupGroupId(strconv.Itoa(int(stat.Gid)))
	if err != nil {
		return list, 0, err
	}
	list.group = group.Name

	if options.human {
		size := float64(pathInfo.info.Size())

		count := 0
		for size >= 1.0 {
			size /= 1024
			count++
		}

		if count < 0 {
			count = 0
		} else if count > 0 {
			size *= 1024
			count--
		}

		var suffix string
		if count == 0 {
			suffix = "B"
		} else if count == 1 {
			suffix = "K"
		} else if count == 2 {
			suffix = "M"
		} else if count == 3 {
			suffix = "G"
		} else if count == 4 {
			suffix = "T"
		} else if count == 5 {
			suffix = "P"
		} else if count == 6 {
			suffix = "E"
		} else {
			suffix = "?"
		}

		sizeStr := ""
		if count == 0 {
			sizeStr = fmt.Sprintf("%d%s", int64(size), suffix)
		} else {
			sizeStr = fmt.Sprintf("%.1f%s", size, suffix)
		}

		if len(sizeStr) > 3 &&
			sizeStr[len(sizeStr)-3:len(sizeStr)-1] == ".0" {
			sizeStr = sizeStr[0:len(sizeStr)-3] + suffix
		}

		list.size = sizeStr

	} else {
		list.size = fmt.Sprintf("%d", pathInfo.info.Size())
	}

	list.epochNano = pathInfo.info.ModTime().UnixNano()

	list.month = pathInfo.info.ModTime().Month().String()[0:3]

	list.day = fmt.Sprintf("%2d", pathInfo.info.ModTime().Day())

	now := time.Now().Unix()
	sixMonth := time.Now().Sub(time.Now().AddDate(0, -6, 0))
	var seconds int64 = int64(sixMonth.Seconds())
	epochSixMonth := now - seconds
	epochModified := pathInfo.info.ModTime().Unix()
	//fmt.Println(time.Unix(epochSixMonth, 0))

	//fmt.Println(time.Unix(epochModified, 0))

	var timeStr string
	if epochModified <= epochSixMonth ||
		epochModified >= (now+5) {
		timeStr = fmt.Sprintf("%d", pathInfo.info.ModTime().Year())
	} else {
		timeStr = fmt.Sprintf("%02d:%02d",
			pathInfo.info.ModTime().Hour(),
			pathInfo.info.ModTime().Minute())
	}

	list.time = timeStr

	list.name = pathInfo.path

	if pathInfo.info.Mode()&os.ModeCharDevice == os.ModeCharDevice {
		list.isCharacter = true
	} else if pathInfo.info.Mode()&os.ModeDevice == os.ModeDevice {
		list.permissions = "b" + list.permissions
		list.isBlock = true
	} else if pathInfo.info.Mode()&os.ModeNamedPipe == os.ModeNamedPipe {
		list.isPipe = true
	} else if pathInfo.info.Mode()&os.ModeSocket == os.ModeSocket {
		list.isSocket = true
	}
	if list.isBlock || list.isCharacter {
		list.major = fmt.Sprintf("%d", uint64(stat.Rdev/256))
		list.minor = fmt.Sprintf("%d", uint64(stat.Rdev%256))
	}
	return list, int(stat.Blocks) / 2, nil

}

func WriteListToOuptut(list []List, terminalWidth int) string {
	if len(list) == 0 {
		return ""
	}
	var output []string
	if options.long {
		var (
			permissionsWidth int = 0
			hardLinksWidth   int = 0
			ownerWidth       int = 0
			groupWidth       int = 0
			sizeWidth        int = 0
			majorWidth       int = 0
			minorWidth       int = 0
			timeWidth        int = 0
		)

		for _, l := range list {
			if len(l.permissions) > permissionsWidth {
				permissionsWidth = len(l.permissions)
			}
			if len(l.hardLinks) > hardLinksWidth {
				hardLinksWidth = len(l.hardLinks)
			}
			if len(l.owner) > ownerWidth {
				ownerWidth = len(l.owner)
			}
			if len(l.group) > groupWidth {
				groupWidth = len(l.group)
			}
			if len(l.major) > majorWidth {
				majorWidth = len(l.major)
			}
			if len(l.minor) > minorWidth {
				minorWidth = len(l.minor)
			}
			if len(l.size) > sizeWidth {
				sizeWidth = len(l.size)
			}
			if l.isBlock || l.isCharacter && len(l.major)+len(l.minor)+3 > sizeWidth {
				sizeWidth = len(l.major) + len(l.minor) + 3
			}
			if len(l.time) > timeWidth {
				timeWidth = len(l.time)
			}
		}

		for _, l := range list {
			str := ""
			// permissions
			str += l.permissions
			for i := 0; i < permissionsWidth-len(l.permissions); i++ {
				str += " "
			}
			str += " "

			// number of hard links (right justified)
			for i := 0; i < hardLinksWidth-len(l.hardLinks); i++ {
				str += " "
			}
			str += l.hardLinks
			str += " "

			// owner
			str += l.owner
			for i := 0; i < ownerWidth-len(l.owner); i++ {
				str += " "
			}
			str += " "

			// group
			str += l.group
			for i := 0; i < groupWidth-len(l.group); i++ {
				str += " "
			}
			str += " "

			// size
			if l.isBlock || l.isCharacter {
				for i := 0; i < majorWidth-len(l.major); i++ {
					str += " "
				}
				str += l.major
				str += ", "
				for i := 0; i < minorWidth-len(l.minor); i++ {
					str += " "
				}
				str += l.minor
				str += " "
			} else {
				for i := 0; i < sizeWidth-len(l.size); i++ {
					str += " "
				}
				str += l.size
				str += " "
			}

			// month
			str += l.month
			str += " "

			// day
			str += l.day
			str += " "

			// time
			for i := 0; i < timeWidth-len(l.time); i++ {
				str += " "
			}
			str += l.time
			str += " "

			// name
			str += WriteName(l)
			output = append(output, str)
		}
	} else if options.one {
		for _, l := range list {
			output = append(output, WriteName(l))
		}
	} else {
		separator := "  "

		// calculate the number of rows needed for column output
		rows := 1
		var colWidth []int
		for {
			colsFloat := float64(len(list)) / float64(rows)
			colsFloat = math.Ceil(colsFloat)
			cols := int(colsFloat)

			colWidth = make([]int, cols)
			for i, _ := range colWidth {
				colWidth[i] = 0
			}

			colList := make([]int, cols)
			for i := 0; i < len(colList); i++ {
				colList[i] = 0
			}

			// calculate necessary column widths
			// also calculate the number of list per column
			for i := 0; i < len(list); i++ {
				col := i / rows
				if colWidth[col] < len(list[i].name) {
					colWidth[col] = len(list[i].name)
				}
				colList[col]++
			}

			// calculate the maximum width of each row
			rowMaxLength := 0
			for i := 0; i < cols; i++ {
				rowMaxLength += colWidth[i]
			}
			rowMaxLength += len(separator) * (cols - 1)

			if rowMaxLength > terminalWidth && rows >= len(list) {
				break
			} else if rowMaxLength > terminalWidth {
				rows++
			} else {
				firstCol := colList[0]
				lastCol := colList[len(colList)-1]

				// prevent short last (right-hand) columns
				if lastCol <= firstCol/2 &&
					firstCol-lastCol >= 5 {
					rows++
				} else {
					break
				}
			}
		}
		str := ""
		for r := 0; r < rows; r++ {
			for i, l := range list {
				if i%rows == r {
					str += WriteName(l)
					for s := 0; s < colWidth[i/rows]-len(l.name); s++ {
						str += " "
					}
					str += separator
				}
			}
			if r != rows-1 {
				str += "\n"
			}
		}
		output = append(output, str)
	}
	return strings.Join(output, "\n")
}

func WriteName(l List) string {
	str := ""
	if options.color {
		appliedColor := false

		hardLinksNum, _ := strconv.Atoi(l.hardLinks)

		// "file.name.txt" -> "*.txt"
		name := strings.Split(l.name, ".")
		extension := ""
		if len(name) > 1 {
			extension = fmt.Sprintf("*.%s", name[len(name)-1])
		}

		if extension != "" && colorsMap[extension] != "" {
			str += colorsMap[extension]
			appliedColor = true
		} else if GetColor(l) != "" {
			str += GetColor(l)
			appliedColor = true
		} else if hardLinksNum > 1 { // multiple hardlinks
			str += colorsMap["multi_hardlink"]
			appliedColor = true
		}

		str += l.name
		if appliedColor {
			str += colorsMap["end"]
		}
	} else {
		str += l.name
	}

	if l.permissions[0] == 'l' && options.long {
		if l.linkOrphan {
			str += fmt.Sprintf(" -> %s%s%s",
				colorsMap["link_orphan_target"],
				l.linkName,
				colorsMap["end"])
		} else {
			str += fmt.Sprintf(" -> %s%s%s", l.linkColor, l.linkName, colorsMap["end"])
		}
	}
	return str
}

func ListDirFiles(dir List) ([]List, int, error) {
	l := make([]List, 0)
	size := 0

	if options.all {
		info, err := os.Stat(dir.name)
		if err != nil {
			return l, 0, err
		}
		list, blocksize, err := CreateList(dir.name,
			FileInfoPath{".", info})
		size += blocksize
		if err != nil {
			return l, 0, err
		}

		infodot, err := os.Stat(dir.name + "/..")
		if err != nil {
			return l, 0, err
		}

		listDot, blocksize, err := CreateList(dir.name,
			FileInfoPath{"..", infodot})
		size += blocksize
		if err != nil {
			return l, 0, err
		}

		l = append(l, list)
		l = append(l, listDot)
	}
	files, err := ReadDir(dir.name)
	// for _, v := range files {
	// 	fmt.Println(v.Name())
	// }
	if err != nil {
		return l, 0, err
	}

	for _, f := range files {
		if []rune(f.Name())[0] == rune('.') && !options.all {
			continue
		}

		_l, blocksize, err := CreateList(dir.name,
			FileInfoPath{f.Name(), f})
		size += blocksize
		if err != nil && !os.IsPermission(err) {
			return l, 0, err
		}
		l = append(l, _l)
	}
	SortList(l)
	return l, size, nil
}

func SortList(listings []List) {
	compareFunc := CompareName
	if options.sortTime {
		compareFunc = CompareTime
	} else if options.sortSize {
		compareFunc = CompareSize
	}
	// for _, v := range listings {
	// 	fmt.Println(v.name)
	// }
	// fmt.Println("-----------")
	for i := len(listings) - 1; i >= 0; i-- {
		for j := 0; j < i; j++ {
			if compareFunc(listings[j], listings[j+1]) == 1 {
				listings[j], listings[j+1] = listings[j+1], listings[j]
			} else if compareFunc(listings[j], listings[j+1]) == 0 {
				if CompareName(listings[j], listings[j+1]) == 1 {
					listings[j], listings[j+1] = listings[j+1], listings[j]
				}
			}
		}
	}

	if options.sortReverse {
		middleIndex := (len(listings) / 2)
		if len(listings)%2 == 0 {
			middleIndex--
		}

		for i := 0; i <= middleIndex; i++ {
			frontIndex := i
			rearIndex := len(listings) - 1 - i

			if frontIndex == rearIndex {
				break
			}

			tmp := listings[frontIndex]
			listings[frontIndex] = listings[rearIndex]
			listings[rearIndex] = tmp
		}
	}
	// for _, v := range listings {
	// 	fmt.Println(v.name)
	// }
	// fmt.Println("==============")
}

func SortDirsFirst(listings []List) []List {

	sortedList := make([]List, 0)

	for _, l := range listings {
		if l.permissions[0] == 'd' {
			sortedList = append(sortedList, l)
		}
	}
	for _, l := range listings {
		if l.permissions[0] != 'd' {
			sortedList = append(sortedList, l)
		}
	}

	return sortedList
}

func CompareName(a, b List) int {
	nameA := strings.ToLower(a.name)
	tmpA := ""
	for _, v := range nameA {
		if v >= 'a' && v <= 'z' || v >= 'A' && v <= 'Z' || v >= '0' && v <= '9' {
			tmpA += string(v)
		}
	}
	nameB := strings.ToLower(b.name)
	tmpB := ""
	for _, v := range nameB {
		if v >= 'a' && v <= 'z' || v >= 'A' && v <= 'Z' || v >= '0' && v <= '9' {
			tmpB += string(v)
		}
	}
	if tmpB == "" && tmpA == "" {
		tmpA = nameA
		tmpB = nameB
	}
	return strings.Compare(tmpA, tmpB)
}

func CompareTime(a, b List) int {
	if a.epochNano > b.epochNano {
		return -1
	} else if a.epochNano == b.epochNano {
		return 0
	}

	return 1
}

func CompareSize(a, b List) int {
	sizeA, _ := strconv.Atoi(a.size)
	sizeB, _ := strconv.Atoi(b.size)

	if sizeA >= sizeB {
		return -1
	}

	return 1
}

func ReadDir(dirname string) ([]os.FileInfo, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	return list, nil
}

func BubbleSort(arr []List, reverse bool) []string {
	compareFunc := CompareName
	if options.sortTime {
		compareFunc = CompareTime
	} else if options.sortSize {
		compareFunc = CompareSize
	}

	for i := len(arr) - 1; i >= 0; i-- {
		for j := 0; j < i; j++ {
			if compareFunc(arr[j], arr[j+1]) == 1 {
				arr[j], arr[j+1] = arr[j+1], arr[j]
			} else if compareFunc(arr[j], arr[j+1]) == 0 {
				if CompareName(arr[j], arr[j+1]) == 1 {
					arr[j], arr[j+1] = arr[j+1], arr[j]
				}
			}
		}
	}
	dirs := make([]string, len(arr))
	if options.sortReverse {
		cnt := 0
		for i := len(arr) - 1; i >= 0; i-- {
			dirs[cnt] = arr[i].name
			cnt++
		}
	} else {
		for i, v := range arr {
			dirs[i] = v.name
		}
	}

	return dirs
}

func GetColor(l List) string {
	if l.permissions[0] == 'd' &&
		l.permissions[8] == 'w' && l.permissions[9] == 't' {
		return colorsMap["directory_o+w_sticky"]
	} else if l.permissions[0] == 'd' && l.permissions[9] == 't' {
		return colorsMap["directory_sticky"]
	} else if l.permissions[0] == 'd' && l.permissions[8] == 'w' {
		return colorsMap["directory_o+w"]
	} else if l.permissions[0] == 'd' { // directory
		return colorsMap["directory"]
	} else if l.permissions[0] == 'l' && l.linkOrphan { // orphan link
		return colorsMap["link_orphan"]
	} else if l.permissions[0] == 'l' { // symlink
		return colorsMap["symlink"]
	} else if l.permissions[3] == 's' { // setuid
		return colorsMap["executable_suid"]
	} else if l.permissions[6] == 's' { // setgid
		return colorsMap["executable_sgid"]
	} else if strings.Contains(l.permissions, "x") { // executable
		return colorsMap["executable"]
	} else if l.isSocket { // socket
		return colorsMap["socket"]
	} else if l.isPipe { // pipe
		return colorsMap["pipe"]
	} else if l.isBlock { // block
		return colorsMap["block"]
	} else if l.isCharacter { // character
		return colorsMap["character"]
	}
	return ""
}

func GetLinkColor(linkPath string) (string, error) {
	info, err := os.Stat(linkPath)
	var linkOrphan, isCharacter, isBlock, isPipe, isSocket bool
	if err != nil && !os.IsPermission(err) {
		if os.IsNotExist(err) {
			linkOrphan = true
			return "", nil
		} else {
			return "", err
		}
	}
	permissions := info.Mode().String()
	if info.Mode()&os.ModeCharDevice == os.ModeCharDevice {
		isCharacter = true
	} else if info.Mode()&os.ModeDevice == os.ModeDevice {
		isBlock = true
	} else if info.Mode()&os.ModeNamedPipe == os.ModeNamedPipe {
		isPipe = true
	} else if info.Mode()&os.ModeSocket == os.ModeSocket {
		isSocket = true
	}

	if permissions[0] == 'd' &&
		permissions[8] == 'w' && permissions[9] == 't' {
		return colorsMap["directory_o+w_sticky"], nil
	} else if permissions[0] == 'd' && permissions[9] == 't' {
		return colorsMap["directory_sticky"], nil
	} else if permissions[0] == 'd' && permissions[8] == 'w' {
		return colorsMap["directory_o+w"], nil
	} else if permissions[0] == 'd' { // directory
		return colorsMap["directory"], nil
	} else if permissions[0] == 'l' && linkOrphan { // orphan link
		return colorsMap["link_orphan"], nil
	} else if permissions[0] == 'l' { // symlink
		return colorsMap["symlink"], nil
	} else if permissions[3] == 's' { // setuid
		return colorsMap["executable_suid"], nil
	} else if permissions[6] == 's' { // setgid
		return colorsMap["executable_sgid"], nil
	} else if strings.Contains(permissions, "x") { // executable
		return colorsMap["executable"], nil
	} else if isSocket { // socket
		return colorsMap["socket"], nil
	} else if isPipe { // pipe
		return colorsMap["pipe"], nil
	} else if isBlock { // block
		return colorsMap["block"], nil
	} else if isCharacter { // character
		return colorsMap["character"], nil
	}
	return "", nil
}
