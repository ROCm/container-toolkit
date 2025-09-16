/**
# Copyright (c) Advanced Micro Devices, Inc. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the \"License\");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an \"AS IS\" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
**/

package logger

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
	"sync"
)

var (
	Log       *log.Logger
	logdir    = "/var/log/"
	logfile   = "amd-container-runtime.log"
	logPrefix = "amd-container-runtime "
	once      sync.Once
)

// SetLogPrefix sets prefix in the log to be exporter or testrunner
func SetLogPrefix(prefix string) {
	logPrefix = prefix
}

// SetLogFile sets the log file name
func SetLogFile(file string) {
	logfile = file
}

// SetLogDir sets the path to the directory of logs
func SetLogDir() {

	isWriteable := func(path string) bool {
		// Create a temporary file in the specified directory.
		// os.CreateTemp will return an error if the directory is not writable.
		file, err := os.CreateTemp(path, "tmp-test-")
		if err != nil {
			return false
		}
		file.Close()           // Close the file
		os.Remove(file.Name()) // Clean up the temporary file

		return true
	}

	if os.Getenv("LOGDIR") != "" {
		logdir = os.Getenv("LOGDIR")

		//check if the user has permission to write to this location
		if !isWriteable(logdir) {
			log.Fatalf("User doesn't have write permission for the specified directory: %v", logdir)
		}

		return
	}

	// Get the current user's information.
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("Failed to get current user: %v", err)
	}
	// for root user, log dir is /var/log
	if currentUser.Uid != "0" {
		//Non-Root user, setting log directory to user's home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get user home directory: %v", err)
		}
		logdir = homeDir
	}
}

func initLogger(console bool) {
	if console {
		Log = log.New(os.Stdout, logPrefix, log.Lmsgprefix)
	} else {
		SetLogDir()

		outfile, err := os.OpenFile(filepath.Join(logdir, logfile),
			os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			return
		}

		Log = log.New(outfile, "", 0)
	}

	Log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func Init(console bool) {
	init := func() {
		initLogger(console)
	}
	once.Do(init)
}
