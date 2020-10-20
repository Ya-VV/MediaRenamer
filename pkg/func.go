package pkg

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
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

type processingAttr struct {
	toSkip       bool
	doByName     bool
	doByExiftool bool
}
type calcMd5 struct {
	strPath string
	hash    []byte
}

var exiftoolExist, verbose, checkDublesFlag bool
var removedCount, skippedCount int
var timeNow = time.Now()
var exifBirthday int64 = 2002
var workDir string
var allFiles = make(map[string][]byte)
var timeExp = regexp.MustCompile(`.*(?P<year>\d{4})[\._:-]?(?P<month>\d{2})[\._:-]?(?P<day>\d{2}).+(?P<hour>\d{2})[\._:-]?(?P<min>\d{2})[\._:-]?(?P<sec>\d{2}).*`)
var defNewNameLayout = "2006-01-02_150405"

//SetVerbose to assign verbose output
func SetVerbose(v bool) {
	if v {
		verbose = v
		fmt.Printf("Setted verbose flag: %v\n", v)
	}
}

//SetCheckDublesFlag set to check dubles  from arguments
func SetCheckDublesFlag(v bool) {
	if v {
		checkDublesFlag = v
		fmt.Printf("Setted checkDublesFlag to: %v\n", v)
	}
}

//SetWorkDir set work directory from arguments
func SetWorkDir(s string) {
	workDir = s
	fmt.Println("Set workdir to: ", s)
}

//Check or ask workdir
func checkWorkDir(logger *log.Logger) string {
	if workDir != "" {
		if !checkPath(workDir) {
			log.Fatal("Dir is not exist")
		}
	} else {
		fmt.Print("Put collection full path: ")
		reader := bufio.NewReader(os.Stdin)
		inputData, err := reader.ReadString('\n')
		check(err)
		workDir = strings.TrimSpace(inputData)
		if !checkPath(workDir) {
			log.Fatal("Dir is not exist")
		}
		logger.Printf("Your choise is a: %v\n", workDir)
		//search for duplicates
		//enable verbose mode?
	}
	return workDir
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

func walkingOnFilesystem(workDir string, logger *log.Logger) ([]string, []string) {
	logger.Println("Started search of supported files on selected path.")
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
		if verbose {
			logger.Printf("visited file or dir: %q\n", path)
		}
		//проверка на подходящее расширение файла
		if _, ok := find(fileExt, filepath.Ext(strings.ToLower(path))); ok {
			if checkDublesFlag {
				allFiles[path] = []byte{}
			}
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
		log.Panic(err)
	}
	logger.Println("Found " + strconv.Itoa(len(dirFiles)) + " files for processing without exiftool")
	logger.Println("Found " + strconv.Itoa(len(forExifTool)) + " files for processing via exiftool")

	if checkDublesFlag {

		calcMd5chan := make(chan calcMd5)
		for path := range allFiles {
			go md5Calculate(path, calcMd5chan, logger)
		}
		for i := 0; i < len(allFiles); i++ {
			itemMd5 := <-calcMd5chan
			allFiles[itemMd5.strPath] = itemMd5.hash
		}

		for key, val := range allFiles {
			delete(allFiles, key)
			foundDubles := []string{}
			for k, v := range allFiles {
				res := bytes.Compare(v, val)
				if res == 0 && k != key {
					logger.Println("Found dublicate of file: ", key, "\n\t--->", k)
					foundDubles = append(foundDubles, k)
				}
			}
			if len(foundDubles) > 0 {
				for _, item := range foundDubles {
					delete(allFiles, item)
					err := os.Remove(item)
					check(err)
					if verbose {
						logger.Println("Removed file: ", item)
					}
					removedCount++
				}
			}
		}
	}

	return dirFiles, forExifTool
}

func fileToProcessing(file string, logger *log.Logger) processingAttr {
	var filematched processingAttr
	fileNameBase := filepath.Base(file)
	if verbose {
		logger.Println("fileToProcessing; basename of file to processing: " + fileNameBase)
	}
	patternToSkip := `(^\d{4}-\d{2}-\d{2}_\d{6}\.)|(^\d{4}-\d{2}-\d{2}_\d{6}\(\d+\)\.)`       //шаблон файлов обработанных раннее
	patternDateInName := `.*\d{4}[\._:-]?\d{2}[\._:-]?\d{2}.+\d{2}[\._:-]?\d{2}[\._:-]?\d{2}` //шаблон файлов имеющих дату в имени
	switch {
	case match(`^\..*`, fileNameBase):
		if verbose {
			logger.Println("fName: " + fileNameBase + " func: fileToProcessing:match; skip file")
		}
		filematched.toSkip = true
		return filematched
	case match(patternToSkip, fileNameBase):
		if verbose {
			logger.Println("fName: " + fileNameBase + " func: fileToProcessing:match; skip file")
		}
		filematched.toSkip = true
		return filematched
	case match(patternDateInName, fileNameBase):
		if verbose {
			logger.Println("fName: " + fileNameBase + " func: fileToProcessing:match; pattern by DateInName")
		}
		filematched.doByName = true
		return filematched
	default:
		if verbose {
			logger.Println("fName: " + fileNameBase + " func: fileToProcessing:match; pattern by doExif")
		}
		filematched.doByExiftool = true
		return filematched
	}
}
func md5Calculate(s string, channel chan calcMd5, logger *log.Logger) {
	f, err := os.Open(s)
	if err != nil {
		log.Panic(err)
	}
	defer f.Close()
	logger.Println("Calculate md5sum of: ", s)
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Panic(err)
	}
	channel <- calcMd5{strPath: s, hash: h.Sum(nil)}
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
	newName := ""
	allDetectedDate := make(map[string]time.Time)
	fileInfos := et.ExtractMetadata(filePath)
	fileExifStrings := []string{"CreateDate", "Create Date", "DateTimeOriginal", "Date/Time Original", "ModifyDate", "Modify Date", "Date", "Profile Date Time", "Media Create Date", "Media Modify Date", "Track Create Date", "Track Modify Date", "File Modification Date/Time", "FileModifyDate"}
	if verbose {
		logger.Println("Exif data of the file: \n", fileInfos[0].Fields)
	}
	for _, fileInfo := range fileInfos {
		if fileInfo.Err != nil {
			logger.Printf("Error concerning %v: %v\n", fileInfo.File, fileInfo.Err)
			continue
		}
		for _, exifString := range fileExifStrings {
			exifTime, err := fileInfo.GetString(exifString)
			if err == nil {
				logger.Printf("getExif:checkField; Exif field <<<%v>>> matched\n", exifString)
				suppositionName, err := parseAndCheckDate(exifTime, logger)
				if err != nil {
					logger.Println("ERROR: exif data corrupted. Checking next exif string.")
					continue
				} else {
					t, err := time.Parse(defNewNameLayout, suppositionName)
					check(err)
					allDetectedDate[suppositionName] = t
				}
			}
		}
		for k, v := range allDetectedDate {
			if newName == "" {
				newName = k
				continue
			} else {
				if allDetectedDate[newName].After(v) {
					newName = k
				}
			}
		}
		return newName, nil
	}
	return "", errors.New("ERROR: exif data corrupted")
}
func fsTimeStamp(item string) (string, error) {
	fInfo, err := os.Stat(item)
	if err != nil {
		return "", err
	}
	fTimestamp := fInfo.ModTime()
	fModTimeNewName := fTimestamp.Format(defNewNameLayout)
	return fModTimeNewName, nil
}
func useFSTimeStamp(fPath string, logger *log.Logger) error {
	newName, err := fsTimeStamp(fPath)
	if err != nil {
		return err
	}
	_, err = parseAndCheckDate(newName, logger)
	if err != nil {
		return err
	}
	logger.Println("fsTimeStamp:rename; newName: " + newName)
	renamer(fPath, newName, logger)
	return nil
}
func parseAndCheckDate(str string, logger *log.Logger) (string, error) {
	exifSliceParsed := timeExp.FindStringSubmatch(str)
	result := make(map[string]string)
	for i, name := range timeExp.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = exifSliceParsed[i]
		}
	}
	err := areDateActual(result, logger)
	if err != nil {
		logger.Println(err)
		return "", err
	}
	newName := result["year"] + "-" + result["month"] + "-" + result["day"] + "_" + result["hour"] + result["min"] + result["sec"]
	return newName, nil
}
func areDateActual(result map[string]string, logger *log.Logger) error {
	parseStr := result["year"] + result["month"] + result["day"] + result["hour"] + result["min"] + result["sec"]
	parseTime, err := time.Parse("20060102150405", parseStr)
	if err != nil {
		return err
	}
	if parseTime.Year() > timeNow.Year() {
		return fmt.Errorf("Parsed year is corrupted: %v. Biger that now: %v", parseTime.Year(), timeNow.Year())
	} else if int64(parseTime.Year()) < exifBirthday {
		return fmt.Errorf("Parsed year is corrupted: %v. Less that exifBirthday: %v", int64(parseTime.Year()), exifBirthday)
	}
	return nil
}
