package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/barasher/go-exiftool"
)

const (
	verbose           = false
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

var patternToSkip = `(^\d{8}_\d{6}\.)|(^\d{8}_\d{6}\(\d+\)\.)|(^\d{8}_\d{6}_\(\d+\)\.)` //шаблон файлов обработанных раннее
var patternDateInName = `^[A-Z]{3}_\d{8}_\d{6}`                                         //шаблон файлов имеющих дату в имени
var patternDateInName2 = `^\d{4}[_:-]\d{2}[_:-]\d{2}[_:-]\d{6}`                         //шаблон файлов имеющих дату в имени
var patternDateInName3 = `^.*\d{4}[_:-]\d{2}[_:-]\d{2}[_:-]\d{2}[_:-]\d{2}[_:-]\d{2}`   //шаблон файлов имеющих дату в имени

type processingAttr struct {
	toSkip       bool
	doByName     bool
	doByName2    bool
	doByName3    bool
	doByExiftool bool
}

func main() {
	workDir := getConfig()
	// log.Println("=== App started ===")
	dirFiles := walkingOnFilesystem(workDir)
	var et *exiftool.Exiftool
	var err error
	mustCompile1 := regexp.MustCompile(`^[A-Z]{3}_(\d{8})_(\d{6})`)
	mustCompile2 := regexp.MustCompile(`^(\d{4})[_:-](\d{2})[_:-](\d{2})[_:-](\d{6})`)
	mustCompile3 := regexp.MustCompile(`^.*(\d{4})[_:-](\d{2})[_:-](\d{2})[_:-](\d{2})[_:-](\d{2})[_:-](\d{2})`)
	initEt := func() {
		if et == nil {
			et, err = exiftool.NewExiftool()
			if err != nil {
				panic(fmt.Errorf("Error when intializing: %v", err))
			}
		}
	}
	if len(dirFiles) > 0 {
		initEt()
		defer et.Close()
	} else {
		fmt.Println("Nothin to do!\nBye :)")
		os.Exit(0)
	}
	for key, val := range dirFiles {
		fmt.Println(key)
		switch {
		case val.doByName:
			nameSlice := mustCompile1.FindStringSubmatch(filepath.Base(key))
			newName := nameSlice[1] + "_" + nameSlice[2]
			renamer(key, newName)
		case val.doByName2:
			nameSlice := mustCompile2.FindStringSubmatch(filepath.Base(key))
			newName := nameSlice[1] + nameSlice[2] + nameSlice[3] + "_" + nameSlice[4]
			renamer(key, newName)
		case val.doByName3:
			nameSlice := mustCompile3.FindStringSubmatch(filepath.Base(key))
			newName := nameSlice[1] + nameSlice[2] + nameSlice[3] + "_" + nameSlice[4] + nameSlice[5] + nameSlice[6]
			renamer(key, newName)
		case val.doByExiftool:
			exifData, err := getExif(et, key)
			if err != nil { //если не получилось вынуть exif
				fmt.Println("===> exifData: ", exifData)
				fmt.Println("===> error: ", err)
				log.Println("Exif data FAILED -> go to filesystem maketime data: ", key)
				fInfo, err := os.Stat(key)
				check(err)
				fTimestamp := fInfo.ModTime()
				newName := fTimestamp.Format(stdLongYear + stdZeroMonth + stdZeroDay + "_" + stdHour + stdZeroMinute + stdZeroSecond)
				renamer(key, newName)
			} else {
				newName := exifData.Format(stdLongYear + stdZeroMonth + stdZeroDay + "_" + stdHour + stdZeroMinute + stdZeroSecond)
				renamer(key, newName)
			}
		default:
			puts("Look like something wrong in main::for::switch block", key)
		}
	}
} //main END
func puts(s ...string) {
	fmt.Println(s)
}
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
func getConfig() string { //получаю каталог который необходимо обработать
	var input string

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
	return input
}
func walkingOnFilesystem(workDir string) map[string]processingAttr {
	// fmt.Println("Walking on filesystem:")
	fileExt := []string{ //обрабатываемые файлы
		"3FR", ".3G2", ".3GP2", ".3GP", ".3GPP", ".A", ".AA", ".AAE", ".AAX", ".ACR", ".AFM", ".ACFM", ".AMFM", ".AI", ".AIT", ".AIFF",
		".AIF", ".AIFC", ".APE", ".ARQ", ".ARW", ".ASF", ".AVI", ".AVIF", ".BMP", ".DIB", ".BPG", ".BTF", ".CHM", ".COS", ".CR2", ".CR3",
		".CRW", ".CIFF", ".CS1", ".CSV", ".DCM", ".DC3", ".DIC", ".DICM", ".DCP", ".DCR", ".DFONT", ".DIVX", ".DJVU", ".DJV", ".DNG",
		".DOC", ".DOT", ".DOCX", ".DOCM", ".DOTX", ".DOTM", ".DPX", ".DR4", ".DYLIB", ".DV", ".DVB", ".DVR-MS", ".EIP", ".EPS", ".EPSF",
		".PS", ".EPUB", ".ERF", ".EXE", ".DLL", ".EXIF", ".EXR", ".EXV", ".F4A", ".F4B", ".F4P", ".F4V", ".FFF", ".FFF", ".FLA", ".FLAC",
		".FLIF", ".FLV", ".FPF", ".FPX", ".GIF", ".GPR", ".GZ", ".GZIP", ".HDP", ".WDP", ".JXR", ".HDR", ".HEIC", ".HEIF", ".HIF", ".HTML",
		".HTM", ".XHTML", ".ICC", ".ICM", ".ICS", ".ICAL", ".IDML", ".IIQ", ".IND", ".INDD", ".INDT", ".INSV", ".INX", ".ISO", ".ITC", ".J2C",
		".J2K", ".JPC", ".JP2", ".JPF", ".JPM", ".JPX", ".JPEG", ".JPG", ".JPE", ".JSON", ".K25", ".KDC", ".KEY", ".KTH", ".LA", ".LFP",
		".LFR", ".LNK", ".LRV", ".M2TS", ".MTS", ".M2T", ".TS", ".M4A", ".M4B", ".M4P", ".M4V", ".MAX", ".MEF", ".MIE", ".MIFF", ".MIF",
		".MKA", ".MKV", ".MKS", ".MOBI", ".AZW", ".AZW3", ".MODD", ".MOI", ".MOS", ".MOV", ".QT", ".MP3", ".MP4", ".MPC", ".MPEG", ".MPG",
		".M2V", ".MPO", ".MQV", ".MRW", ".MXF", ".NEF", ".NMBTEMPLATE", ".NRW", ".NUMBERS", ".O", ".ODB", ".ODC", ".ODF", ".ODG", ".", ".ODI",
		".ODP", ".ODS", ".ODT", ".OFR", ".OGG", ".OGV", ".OPUS", ".ORF", ".OTF", ".PAC", ".PAGES", ".PCD", ".PCX", ".PDB", ".PRC", ".PDF",
		".PEF", ".PFA", ".PFB", ".PFM", ".PGF", ".PICT", ".PCT", ".PLIST", ".PMP", ".PNG", ".JNG", ".MNG", ".PPM", ".PBM", ".PGM", ".PPT",
		".PPS", ".POT", ".POTX", ".POTM", ".PPAX", ".PPAM", ".PPSX", ".PPSM", ".PPTX", ".PPTM", ".PSD", ".PSB", ".PSDT", ".PSP", ".PSPIMAGE",
		".QTIF", ".QTI", ".QIF", ".R3D", ".RA", ".RAF", ".RAM", ".RPM", ".RAR", ".RAW", ".RAW", ".RIFF", ".RIF", ".RM", ".RV", ".RMVB", ".RSRC",
		".RTF", ".RW2", ".RWL", ".RWZ", ".SEQ", ".SKETCH", ".SO", ".SR2", ".SRF", ".SRW", ".SVG", ".SWF", ".THM", ".THMX", ".TIFF", ".TIF", ".TTF",
		".TTC", ".TORRENT", ".TXT", ".VCF", ".VCARD", ".VOB", ".VRD", ".VSD", ".WAV", ".WEBM", ".WEBP", ".WMA", ".WMV", ".WTV", ".WV", ".X3F", ".XCF",
		".XLS", ".XLT", ".XLSX", ".XLSM", ".XLSB", ".XLTX", ".XLTM", ".XMP", ".ZIP",
	}
	subDirToSkip := "skip"
	dirFiles := make(map[string]processingAttr) //для хранения всего списка подходящих файлов, где: key- полный путь; val- полное имя файла

	err := filepath.Walk(workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() && info.Name() == subDirToSkip {
			fmt.Printf("skipping a dir without errors: %+v \n", info.Name())
			return filepath.SkipDir
		}
		// fmt.Printf("visited file or dir: %q\n", path)

		//проверка на подходящее расширение файла (в нижнем регистре) со слайса fileExt + признак по которому обрабатывать
		if _, ok := find(fileExt, filepath.Ext(strings.ToLower(path))); ok {
			fProcessing, err := fileToProcessing(path)
			check(err)
			// не добавляю в мапу для обработки если файл в этом не нуждается
			if !fProcessing.toSkip {
				dirFiles[path] = fProcessing
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", workDir, err)
		log.Fatal(err)
	}

	return dirFiles
}
func fileToProcessing(file string) (processingAttr, error) {
	var filematched processingAttr
	switch {
	case match(patternToSkip, filepath.Base(file)):
		puts(filepath.Base(file), "---> skip file")
		filematched.toSkip = true
		return filematched, nil
	case match(patternDateInName, filepath.Base(file)):
		puts(filepath.Base(file), "---> pattern by inName")
		filematched.doByName = true
		return filematched, nil
	case match(patternDateInName2, filepath.Base(file)):
		puts(filepath.Base(file), "---> pattern by inName")
		filematched.doByName2 = true
		return filematched, nil
	case match(patternDateInName3, filepath.Base(file)):
		puts(filepath.Base(file), "---> pattern by inName")
		filematched.doByName3 = true
		return filematched, nil
	default:
		puts(filepath.Base(file), "---> pattern by doExif")
		filematched.doByExiftool = true
		return filematched, nil
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
		if strings.ToLower(item) == val {
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
	puts("====================================================================================================")
	puts("fullpath: ", fullPath)
	puts("newname ---> ", newName)
	path := filepath.Dir(fullPath) + "/"
	extFile := filepath.Ext(fullPath)
	fullNewName := path + newName + extFile
	if fileExists(fullNewName) {
		nextName := newName
		for count := 1; fileExists(path + nextName + extFile); count++ {
			nextName = newName + "(" + strconv.Itoa(count) + ")"
		}
		fullNewName = path + nextName + extFile
		puts("New newName: ", fullNewName)
	}

	err := os.Rename(fullPath, fullNewName)
	check(err)
}
func getExif(et *exiftool.Exiftool, filePath string) (time.Time, error) {
	fname := filePath

	fileInfos := et.ExtractMetadata(fname)
	for _, fileInfo := range fileInfos {
		if fileInfo.Err != nil {
			fmt.Printf("Error concerning %v: %v\n", fileInfo.File, fileInfo.Err)
			continue
		}
		if verbose {
			for k, v := range fileInfo.Fields {
				fmt.Printf("[%v] %v\n", k, v)
			}
		}
		// [CreateDate] 		2008:06:18 22:02:45
		// [DateTimeOriginal] 	2008:06:18 22:02:45
		// [ModifyDate] 		2008:06:18 22:02:45
		tLayout1 := "2006:01:02 15:04:05"
		// [Date] 				2020:03:02 12:57:46.145865+02:00
		tLayout2 := "2006:01:02 15:04:05.999999-07:00"
		// [FileModifyDate] 	2019:01:14 11:08:20+02:00
		tLayout3 := "2006:01:02 15:04:05-07:00"
		if exifTime, err := fileInfo.GetString("CreateDate"); err == nil {
			puts("Exif field <<<CreateDate>>> matched")
			return parseExifTime(tLayout1, exifTime)
		} else if exifTime, err := fileInfo.GetString("DateTimeOriginal"); err == nil {
			puts("Exif field <<<DateTimeOriginal>>> matched")
			return parseExifTime(tLayout1, exifTime)
		} else if exifTime, err := fileInfo.GetString("Date"); err == nil {
			puts("Exif field <<<Date>>> matched")
			return parseExifTime(tLayout2, exifTime)
		} else if exifTime, err := fileInfo.GetString("ModifyDate"); err == nil {
			puts("Exif field <<<ModifyDate>>> matched")
			return parseExifTime(tLayout1, exifTime)
		} else if exifTime, err := fileInfo.GetString("FileModifyDate"); err == nil {
			puts("Exif field <<<FileModifyDate>>> matched")
			return parseExifTime(tLayout3, exifTime)
		}
	}
	return time.Time{}, errors.New("ERROR: exif data corrupted")
}
func parseExifTime(layout string, it string) (time.Time, error) {
	return time.Parse(layout, it)
}
