/*
 * Copyright (c) 2024 OceanBase.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ob

import (
	"testing"

	"github.com/oceanbase/obshell/ob/agent/constant"
	"github.com/oceanbase/obshell/ob/agent/repository/model/oceanbase"
)

func TestCalculateDiskSpaceRequirementsForObshellInstall(t *testing.T) {
	packages := []oceanbase.UpgradePkgInfo{
		{Name: constant.PKG_OCEANBASE_CE_LIBS, Size: 100, PayloadSize: 50},
		{Name: constant.PKG_OBSHELL, Size: 200, PayloadSize: 80},
		{Name: constant.PKG_OBSHELL, Size: 300, PayloadSize: 70},
	}

	t.Run("shared file system uses the larger transient peak", func(t *testing.T) {
		upgradeSize, binSize := calculateDiskSpaceRequirements(packages, true)
		// Persistent download/extraction is 800 bytes. The final OBShell temp
		// (300) is larger than the private RPM snapshot (80).
		if upgradeSize != 1100 {
			t.Fatalf("shared file system requirement is %d, want 1100", upgradeSize)
		}
		if binSize != 0 {
			t.Fatalf("shared OBShell bin requirement is %d, want 0", binSize)
		}
	})

	t.Run("separate file systems reserve both locations", func(t *testing.T) {
		upgradeSize, binSize := calculateDiskSpaceRequirements(packages, false)
		if upgradeSize != 880 {
			t.Fatalf("upgrade file system requirement is %d, want 880", upgradeSize)
		}
		if binSize != 300 {
			t.Fatalf("OBShell bin file system requirement is %d, want 300", binSize)
		}
	})
}

func TestCalculateLegacyDiskSpaceRequirementForNonObshellUpgrade(t *testing.T) {
	packages := []oceanbase.UpgradePkgInfo{
		{Name: constant.PKG_OCEANBASE_CE, Size: 100, PayloadSize: 50, ChunkCount: 2},
		{Name: constant.PKG_OCEANBASE_CE_LIBS, Size: 200, PayloadSize: 75, ChunkCount: 3},
	}

	// The legacy path intentionally ignores ChunkCount and keeps the original
	// Size + PayloadSize calculation for non-OBShell upgrades.
	if actual := calculateLegacyDiskSpaceRequirement(packages); actual != 425 {
		t.Fatalf("legacy non-OBShell disk requirement is %d, want 425", actual)
	}
}

func TestRpmDownloadSizeUpperBoundIncludesCompleteRpm(t *testing.T) {
	info := oceanbase.UpgradePkgInfo{PayloadSize: 80, ChunkCount: 1}
	if actual := rpmDownloadSizeUpperBound(info); actual != constant.CHUNK_SIZE {
		t.Fatalf("complete RPM upper bound is %d, want one full chunk (%d)", actual, constant.CHUNK_SIZE)
	}

	info = oceanbase.UpgradePkgInfo{PayloadSize: constant.CHUNK_SIZE + 1, ChunkCount: 1}
	if actual := rpmDownloadSizeUpperBound(info); actual != info.PayloadSize {
		t.Fatalf("upper bound smaller than recorded payload: got %d, want %d", actual, info.PayloadSize)
	}
}

func TestAddDiskSpaceSafetyMarginRoundsUp(t *testing.T) {
	testCases := []struct {
		size uint64
		want uint64
	}{
		{size: 0, want: 0},
		{size: 1, want: 2},
		{size: 10, want: 11},
		{size: 11, want: 13},
		{size: 1100, want: 1210},
		{size: ^uint64(0), want: ^uint64(0)},
	}

	for _, testCase := range testCases {
		if actual := addDiskSpaceSafetyMargin(testCase.size); actual != testCase.want {
			t.Errorf("safety margin for %d is %d, want %d", testCase.size, actual, testCase.want)
		}
	}
}
