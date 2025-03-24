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

package main

import (
	"os"

	"github.com/ROCm/container-runtime/internal/logger"
	"github.com/ROCm/container-runtime/internal/runtime"
)

func main() {
	logger.Init(false)
	logger.Log.Printf("Creating ROCm container runtime with args %v", os.Args)
	rt, err := runtime.New(os.Args)
	if err != nil {
		logger.Log.Printf("Failed to create container runtime, err = %v", err)
		os.Exit(1)
	}

	logger.Log.Printf("Running ROCm container runtime")
	err = rt.Run()
	if err != nil {
		logger.Log.Printf("Failed to run container runtime, err = %v", err)
		os.Exit(1)
	}
}
