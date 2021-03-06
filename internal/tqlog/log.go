/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package tqlog

import (
	"fmt"
	"log"
	"os"
	"strings"
)

//FeedbackFn is used to print messages in the CLI. It's a callback from ishell
type FeedbackFn func(string, ...interface{})

// LogFile is based on log package. Supports log levels and printing messages to stdout
type LogFile struct {
	file   *os.File
	logger *log.Logger
}

var tranqapLog *LogFile
var printFeedback FeedbackFn

// Init bootstraps the logger. printShell effectively is the ishell instance.
// It is used to print messages on the screen
func Init(fname string, feedbackFn func(string, ...interface{})) error {
	if len(fname) > 0 {
		f, err := os.OpenFile(fname, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("Error opening log file %s: %s", fname, err)
		}
		tranqapLog = &LogFile{f, log.New(f, "", log.LstdFlags)}
	}

	printFeedback = feedbackFn

	return nil
}

func (l *LogFile) logError(format string, a ...interface{}) {
	var msgFormat strings.Builder

	if string(format[len(format)-1]) != "\n" {
		fmt.Fprintf(&msgFormat, "ERROR: %s\n", format)
	} else {
		fmt.Fprintf(&msgFormat, "ERROR: %s", format)
	}

	l.logger.Printf(msgFormat.String(), a...)
}

func (l *LogFile) logInfo(format string, a ...interface{}) {
	var msgFormat strings.Builder
	fmt.Fprintf(&msgFormat, "INFO: %s", format)
	l.logger.Printf(msgFormat.String(), a...)
}

//
// Exported wrappers
//

// Error logs with prefix ERROR in file and stdout
func Error(format string, a ...interface{}) {
	if tranqapLog == nil {
		return
	}

	tranqapLog.logError(format, a...)
}

// Info logs only in file
func Info(format string, a ...interface{}) {
	if tranqapLog == nil {
		return
	}

	tranqapLog.logInfo(format, a...)
}

// Feedback prints on the shell
func Feedback(format string, a ...interface{}) {
	if printFeedback != nil {
		printFeedback(format, a...)
	}
}

// Close the log file
func Close() {
	if tranqapLog == nil {
		return
	}

	tranqapLog.file.Close()
}
