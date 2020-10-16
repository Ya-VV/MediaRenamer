package pkg

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/barasher/go-exiftool"
	homedir "github.com/mitchellh/go-homedir"
)

//LetsGo basis action
func LetsGo() {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	logFile, err := os.OpenFile(home+"/"+"yarenamer-"+timeNow.Format("2006-01-02_15-04-05")+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Panic(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	logger := log.New(mw, "INFO: ", log.Flags()&^(log.Ldate|log.Ltime))
	defer logFile.Close()

	logger.Println("yaRenamer started!")

	checkEt(logger)
	workDir = checkWorkDir(logger)
	dirFiles, forExifTool := walkingOnFilesystem(workDir, logger)
	totalFiles := len(dirFiles) + len(forExifTool)
	if totalFiles == 0 {
		logger.Println("Nothin to do!\nBye :)")
		os.Exit(0)
	}

	if len(dirFiles) > 0 {
		for _, item := range dirFiles {
			if !fileExists(item) {
				continue
			}
			logger.SetPrefix(filepath.Base(item) + " ")
			newName, err := parseAndCheckDate(filepath.Base(item), logger)
			if err == nil {
				renamer(item, newName, logger)
				logger.Println("DateInName: new name is a: " + newName + " of file: " + item)
			} else {
				logger.Println(err)
				logger.Println("Moved to exifRenamer func; when DateInName data corrupted")
				forExifTool = append(forExifTool, item)
				continue
			}
		}
	}
	if len(forExifTool) > 0 && exiftoolExist {
		logger.SetPrefix("INFO: ")
		et, err := exiftool.NewExiftool()
		if err != nil {
			panic(fmt.Errorf("Error when intializing: %v", err))
		}
		defer et.Close()
		for _, item := range forExifTool {
			if !fileExists(item) {
				if verbose {
					logger.Println("exifProcessing::skip, the file no exist: ", item)
				}
				continue
			}
			logger.SetPrefix(filepath.Base(item) + " ")
			exifDateParsed, err := getExif(et, item, logger)
			if err != nil { //if getExif data is failed
				logger.Println("func:fsTimeStamp; when exif data corrupted.")
				err := useFSTimeStamp(item, logger)
				if err != nil {
					logger.Println("func::useFSTimeStamp FAILED, skipped file: ", item)
					logger.Println(err)
					totalFiles--
					skippedCount++
				}
			} else {
				renamer(item, exifDateParsed, logger)
				logger.Println("exifToolRename; newName: " + exifDateParsed)
			}
		}
	} else {
		if !exiftoolExist {
			logger.Println("SKIPPED: ", len(forExifTool), " files in ExifTool processing. Because exiftool is not installed.")
			totalFiles = totalFiles - len(forExifTool)
		}
	}
	logger.SetPrefix("INFO: ")
	logger.Println("Total files processed: ", totalFiles)
	if skippedCount > 0 {
		logger.Println("Total files skipped: ", skippedCount)
	}
	if checkDublesFlag {
		logger.Println("Total removed dublicates: ", removedCount)
	}
}
