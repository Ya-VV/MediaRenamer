package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
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

type processingAttr struct {
	toSkip       bool
	doByName     bool
	doByExiftool bool
}

var exiftoolExist bool
var timeNow = time.Now()
var exifBirthday int64 = 2002

func puts(s ...string) {
	fmt.Println(s)
}
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
func checkEt(logger *log.Logger) {
	out, err := exec.Command("/usr/bin/env", "exiftool", "-ver").Output()
	if err == nil {
		cmdOut := string(out)
		cmdOut = strings.TrimSuffix(cmdOut, "\n")
		etVersion, err := strconv.ParseFloat(cmdOut, 64)
		check(err)
		logger.Println("ExifTool installed. Version: ", etVersion)
		exiftoolExist = true
	} else {
		logger.Println("ExifTool not found!")
		logger.Println("Will be processed only files who have TimeStamp in the name.")
		exiftoolExist = false
	}
}

//Ask to workdir
func getConfig(logger *log.Logger) string {
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
		logger.Printf("Your choise is a: %v\n", input)
	}
	return input
}
func walkingOnFilesystem(workDir string, logger *log.Logger) ([]string, []string) {
	//fileExt: array fo file extensions to processing
	fileExt := []string{
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
	//для хранения списка подходящих файлов с датой в имени, где каждый item - полный путь;
	var dirFiles []string
	//для хранения списка подходящих файлов для exiftool, где каждый item - полный путь;
	var forExifTool []string

	err := filepath.Walk(workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() && match(`^\..*`, info.Name()) {
			logger.Printf("skipping a dir without errors: %+v \n", info.Name())
			return filepath.SkipDir
		}
		// fmt.Printf("visited file or dir: %q\n", path)

		//проверка на подходящее расширение файла (в нижнем регистре) со слайса fileExt + признак по которому обрабатывать
		if _, ok := find(fileExt, filepath.Ext(strings.ToLower(path))); ok {
			fProcessing := fileToProcessing(path, logger)
			// не добавляю в мапу для обработки если файл в этом не нуждается
			if !fProcessing.toSkip {
				if fProcessing.doByExiftool && exiftoolExist {
					forExifTool = append(forExifTool, path)
				} else {
					dirFiles = append(dirFiles, path)
				}
			}
		}
		return nil
	})

	if err != nil {
		logger.Printf("error walking the path %q: %v\n", workDir, err)
		log.Fatal(err)
	}
	logger.Println("Found " + strconv.Itoa(len(dirFiles)) + " files for processing without exiftool")
	logger.Println("Found " + strconv.Itoa(len(forExifTool)) + " files for processing via exiftool")
	return dirFiles, forExifTool
}
func fileToProcessing(file string, logger *log.Logger) processingAttr {
	var filematched processingAttr
	fileNameBase := filepath.Base(file)
	logger.Println("fileToProcessing; basename of file to processing: " + fileNameBase)
	patternToSkip := `(^\d{8}_\d{6}\.)|(^\d{8}_\d{6}\(\d+\)\.)`                                        //шаблон файлов обработанных раннее
	patternDateInName := `.*\d{4}[\._:-]?\d{2}[\._:-]?\d{2}[\._:-]?\s?\d{2}[\._:-]?\d{2}[\._:-]?\d{2}` //шаблон файлов имеющих дату в имени
	switch {
	case match(`^\..*`, fileNameBase):
		logger.Println("fName: " + fileNameBase + " func: fileToProcessing:match; skip file")
		filematched.toSkip = true
		return filematched
	case match(patternToSkip, fileNameBase):
		logger.Println("fName: " + fileNameBase + " func: fileToProcessing:match; skip file")
		filematched.toSkip = true
		return filematched
	case match(patternDateInName, fileNameBase):
		logger.Println("fName: " + fileNameBase + " func: fileToProcessing:match; pattern by DateInName")
		filematched.doByName = true
		return filematched
	default:
		logger.Println("fName: " + fileNameBase + " func: fileToProcessing:match; pattern by doExif")
		filematched.doByExiftool = true
		return filematched
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
func renamer(fullPath string, newName string, logger *log.Logger) {
	logger.Println("renamer:start, newName: " + newName)
	path := filepath.Dir(fullPath) + "/"
	extFile := filepath.Ext(fullPath)
	fullNewName := path + newName + extFile
	logger.Println("renamer:newFullName: " + fullNewName)
	if fileExists(fullNewName) {
		nextName := newName
		logger.Println("renamer:fileExists, newName: " + newName)
		for count := 1; fileExists(path + nextName + extFile); count++ {
			nextName = newName + "(" + strconv.Itoa(count) + ")"
		}
		fullNewName = path + nextName + extFile
		logger.Println("renamer:fileExists, newFullName: " + fullNewName)
	}

	err := os.Rename(fullPath, fullNewName)
	check(err)
}
func getExif(et *exiftool.Exiftool, filePath string, logger *log.Logger) (string, error) {
	fileInfos := et.ExtractMetadata(filePath)
	fileExifStrings := []string{"CreateDate", "DateTimeOriginal", "ModifyDate", "Date", "FileModifyDate", "File Modification Date/Time"}
	timeLayout := regexp.MustCompile(`.*(\d{4})[\._:-]?(\d{2})[\._:-]?(\d{2})[\._:-]?\s?(\d{2})[\._:-]?(\d{2})[\._:-]?(\d{2}).*`)
	for _, fileInfo := range fileInfos {
		if fileInfo.Err != nil {
			logger.Printf("Error concerning %v: %v\n", fileInfo.File, fileInfo.Err)
			continue
		}
		if verbose {
			for k, v := range fileInfo.Fields {
				fmt.Printf("[%v] %v\n", k, v)
			}
		}
		for _, exifString := range fileExifStrings {
			if exifTime, err := fileInfo.GetString(exifString); err == nil {
				logger.Printf("getExif:checkField; Exif field <<<%v>>> matched\n", exifString)
				exifSliceParsed := timeLayout.FindStringSubmatch(exifTime)
				if err := areYearActual(exifSliceParsed[1], logger); err != nil {
					logger.Printf("ERROR: exif data (file year) corrupted: %v. Checking next exif string\n", exifSliceParsed[1])
					continue
				}
				exifDateParsed := exifSliceParsed[1] + exifSliceParsed[2] + exifSliceParsed[3] + "_" + exifSliceParsed[4] + exifSliceParsed[5] + exifSliceParsed[6]
				return exifDateParsed, nil
			}
		}
	}
	return "", errors.New("ERROR: exif data corrupted")
}
func fsTimeStamp(item string) (string, error) {
	fInfo, err := os.Stat(item)
	if err != nil {
		return "", nil
	}
	fTimestamp := fInfo.ModTime()
	fModTimeNewName := fTimestamp.Format(stdLongYear + stdZeroMonth + stdZeroDay + "_" + stdHour + stdZeroMinute + stdZeroSecond)
	return fModTimeNewName, nil
}
func useFSTimeStamp(fPath string, logger *log.Logger) {
	newName, err := fsTimeStamp(fPath)
	check(err)
	logger.Println("fsTimeStamp:rename; newName: " + newName)
	renamer(fPath, newName, logger)
}
func areYearActual(parsedYearStr string, logger *log.Logger) error {
	year, err := strconv.ParseInt(parsedYearStr, 10, 64)
	check(err)
	if year > int64(timeNow.Year()) {
		logger.Printf("Parsed year is corrupted: %v. Biger that now: %v\n", year, int64(timeNow.Year()))
		return errors.New("Parsed year is corrupted")
	} else if year < exifBirthday {
		logger.Printf("Parsed year is corrupted: %v. Less that exifBirthday: %v\n", year, exifBirthday)
		return errors.New("Parsed year is corrupted")
	}
	return nil
}
