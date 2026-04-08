/**
# Copyright (c) Advanced Micro Devices, Inc. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
**/

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ROCm/container-toolkit/internal/logger"
)

var (
	versionFlag = flag.Bool("version", false, "Display version information")
)

func main() {
	flag.Parse()
	logger.Init(false)

	if *versionFlag {
		fmt.Println("AMD Container Runtime Hook version 1.0.0")
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: amd-container-runtime-hook <command>\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  prestart   - Configure GPU devices before container start\n")
		fmt.Fprintf(os.Stderr, "  poststop   - Release GPU resources after container stop\n")
		os.Exit(2)
	}

	command := args[0]
	switch command {
	case "prestart":
		if err := doPrestart(); err != nil {
			logger.Log.Printf("prestart hook failed: %v", err)
			os.Exit(1)
		}
	case "poststop":
		if err := doPoststop(); err != nil {
			logger.Log.Printf("poststop hook failed: %v", err)
			os.Exit(1)
		}
	case "poststart":
		// No-op for compatibility
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(2)
	}
}
