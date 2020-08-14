package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
)

const (
	stdLongMonth      = "January"
	stdMonth          = "Jan"
	stdNumMonth       = "1"
	stdZeroMonth      = "01"
	stdLongWeekDay    = "Monday"
	stdWeekDay        = "Mon"
	stdDay            = "2"
	stdUnderDay       = "_2"
	stdZeroDay        = "02"
	stdHour           = "15"
	stdHour12         = "3"
	stdZeroHour12     = "03"
	stdMinute         = "4"
	stdZeroMinute     = "04"
	stdSecond         = "5"
	stdZeroSecond     = "05"
	stdLongYear       = "2006"
	stdYear           = "06"
	stdPM             = "PM"
	stdpm             = "pm"
	stdTZ             = "MST"
	stdISO8601TZ      = "Z0700"  // prints Z for UTC
	stdISO8601ColonTZ = "Z07:00" // prints Z for UTC
	stdNumTZ          = "-0700"  // always numeric
	stdNumShortTZ     = "-07"    // always numeric
	stdNumColonTZ     = "-07:00" // always numeric
)

func main() {
	var input string     //для хранения директории с медиафайлами
	fileExt := []string{ //обрабатываемые файлы
		".jpg", ".jpeg", ".arw", ".png", ".nef", ".cr2",
		".mts", ".mp4", ".3gp", ".m4v", ".mov", ".avi",
	}
	// if fileExists("/usr/bin/exiftool") {
	// 	fmt.Println("ExifTool here")
	// } else {
	// 	log.Fatal("Install ExifTool first")
	// }
	if len(os.Args) == 2 {
		input = os.Args[1]
		if !checkPath(input) {
			log.Fatal("Dir is not exist")
		}
	} else {
		fmt.Print("Put collection path: ")
		reader := bufio.NewReader(os.Stdin)
		inputData, err := reader.ReadString('\n')
		check(err)
		input = strings.TrimSpace(inputData)
		if !checkPath(input) {
			log.Fatal("Dir is not exist")
		}
		fmt.Printf("Your choise is a: %v\n", input)
	}
	os.Chdir(input)
	log.Println("=== App started ===")
	subDirToSkip := "skip"

	dirFiles := make(map[string]string) //key- полный путь; val- полное имя файла

	fmt.Println("On filesystem:")
	err := filepath.Walk(input, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() && info.Name() == subDirToSkip {
			fmt.Printf("skipping a dir without errors: %+v \n", info.Name())
			return filepath.SkipDir
		}
		fmt.Printf("visited file or dir: %q\n", path)

		//проверка на подходящее расширение файла (в нижнем регистре) со слайса fileExt
		if _, ok := find(fileExt, filepath.Ext(strings.ToLower(path))); ok {
			dirFiles[path] = filepath.Base(path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", input, err)
		log.Fatal(err)
	}
	for key, val := range dirFiles {
		fmt.Println(key, "    ", val)
		patternToSkip := `(^\d{8}_\d{6}\.)|(^\d{8}_\d{6}\(\d+\)\.)|(^\d{8}_\d{6}_\(\d+\)\.)` //шаблон файлов обработанных раннее
		patternDateInName := `^[A-Z]{3}_\d{8}_\d{6}`                                         //шаблон файлов имеющих дату в имени
		// patternDateInName2 := `^\d{4}[_:-]\d{2}[_:-]\d{2}[_:-]\d{6}`                          //шаблон файлов имеющих дату в имени
		patternDateInNameLong := `^.*\d{4}[_:-]\d{2}[_:-]\d{2}[_:-]\d{2}[_:-]\d{2}[_:-]\d{2}` //шаблон файлов имеющих дату в имени
		matched := match(patternToSkip, val)

		if matched {
			fmt.Println(val, "---> skip file")
			continue
		} else {
			switch {
			case match(patternDateInName, val):
				r := regexp.MustCompile(`^[A-Z]{3}_(?P<one>\d{8})_(?P<two>\d{6})`)
				nameSlice := r.FindStringSubmatch(val)
				nameSlice = nameSlice[1:] //убираю элемент в котором содержится val
				newName := nameSlice[0] + "_" + nameSlice[1]
				renamer(key, newName)
			case match(patternDateInNameLong, val):
				r := regexp.MustCompile(`^.*\(?P<one>\d{4})[_:-](?P<two>\d{2})[_:-](?P<three>\d{2})[_:-](?P<four>\d{2})[_:-](?P<five>\d{2})[_:-](?P<six>\d{2})`)
				nameSlice := r.FindStringSubmatch(val)
				newName := nameSlice[1] + nameSlice[2] + nameSlice[3] + "_" + nameSlice[4] + nameSlice[5] + nameSlice[6]
				renamer(key, newName)
			default:
				exifData, err := getExif(key)
				if err != nil { //если не получилось вынуть exif
					log.Println("Exif data FAILED -> go to filesystem maketime data")
					fInfo, err := os.Stat(key)
					check(err)
					fTimestamp := fInfo.ModTime()
					newName := fTimestamp.Format(stdLongYear + stdZeroMonth + stdZeroDay + "_" + stdHour + stdZeroMinute + stdZeroSecond)
					renamer(key, newName)
				} else {
					newName := exifData.Format(stdLongYear + stdZeroMonth + stdZeroDay + "_" + stdHour + stdZeroMinute + stdZeroSecond)
					renamer(key, newName)
				}
			}
		}
	}
} //main END

func check(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
func checkPath(somePath string) bool {
	info, err := os.Stat(somePath)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}
func find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}
func match(pattern string, text string) bool {
	m, err := regexp.Match(pattern, []byte(text))
	check(err)
	return m
}
func renamer(fullPath string, newName string) {
	path := filepath.Dir(fullPath) + "/"
	extFile := filepath.Ext(fullPath)
	if fileExists(path + newName + extFile) {
		// shortName := strings.ReplaceAll(newName, extFile, "")
		for count := 0; fileExists(path + newName + extFile); {
			count++
			newName = newName + "(" + strconv.Itoa(count) + ")"
		}
	}

	err := os.Rename(fullPath, newName+extFile)
	check(err)
}
func getExif(filePath string) (time.Time, error) {
	fname := filePath

	f, err := os.Open(fname)
	check(err)

	exif.RegisterParsers(mknote.All...)

	x, err := exif.Decode(f)
	if err != nil {
		f.Close()
	}

	tm, _ := x.DateTime()
	fmt.Println("Taken: ", tm)

	if f.Close() != nil {
		fmt.Println(err)
	}

	return tm, err
}
