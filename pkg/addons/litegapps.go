package addons

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type LiteGappsAddon struct {
	*BaseAddon
	dlLinks     map[string]map[string][]string
	apiLevelMap map[string]string
	extractTo   string
	currentArch string
}

func NewLiteGappsAddon() *LiteGappsAddon {
	versions := []string{"8.1.0", "9.0.0", "10.0.0", "11.0.0", "12.0.0", "12.0.0_64only", "13.0.0", "13.0.0_64only"}

	dlLinks := map[string]map[string][]string{
		"13.0.0": {
			"x86_64": {"https://downloads.sourceforge.net/project/litegapps/litegapps/x86_64/33/lite/2024-02-22/AUTO-LiteGapps-x86_64-13.0-20240222-official.zip", "d91a18a28cc2718c18726a59aedcb8da"},
			"arm64":  {"https://downloads.sourceforge.net/project/litegapps/litegapps/arm64/33/lite/2024-10-22/LiteGapps-arm64-13.0-20241022-official.zip", "a8b1181291fe70d1e838a8579218a47c"},
			"arm":    {"https://downloads.sourceforge.net/project/litegapps/litegapps/arm/33/lite/2024-08-15/AUTO-LiteGapps-arm-13.0-20240815-official.zip", "5a1d192a42ef97693f63d166dea89849"},
		},
		"13.0.0_64only": {
			"x86_64": {"https://downloads.sourceforge.net/project/litegapps/litegapps/x86_64/33/lite/2024-02-22/AUTO-LiteGapps-x86_64-13.0-20240222-official.zip", "d91a18a28cc2718c18726a59aedcb8da"},
			"arm64":  {"https://downloads.sourceforge.net/project/litegapps/litegapps/arm64/33/lite/2024-10-22/LiteGapps-arm64-13.0-20241022-official.zip", "a8b1181291fe70d1e838a8579218a47c"},
		},
		"12.0.0": {
			"arm64": {"https://downloads.sourceforge.net/project/litegapps/litegapps/arm64/31/lite/2024-10-10/AUTO-LiteGapps-arm64-12.0-20241010-official.zip", "ed3196b7d6048ef4adca6388a771cd84"},
			"arm":   {"https://downloads.sourceforge.net/project/litegapps/litegapps/arm/31/lite/v2.5/%5BAUTO%5DLiteGapps_arm_12.0_v2.5_official.zip", "35e1f98dd136114fc1ca74e3a0539cfa"},
		},
		"12.0.0_64only": {
			"arm64": {"https://downloads.sourceforge.net/project/litegapps/litegapps/arm64/31/lite/2024-10-10/AUTO-LiteGapps-arm64-12.0-20241010-official.zip", "ed3196b7d6048ef4adca6388a771cd84"},
		},
		"11.0.0": {
			"x86_64": {"https://downloads.sourceforge.net/project/litegapps/litegapps/x86_64/30/lite/2024-10-12/AUTO-LiteGapps-x86_64-11.0-20241012-official.zip", "5c2a6c354b6faa6973dd3f399bbe162d"},
			"x86":    {"https://downloads.sourceforge.net/project/litegapps/litegapps/x86/30/lite/2024-10-12/AUTO-LiteGapps-x86-11.0-20241012-official.zip", "7252ea97a1d66ae420f114bfe7089070"},
			"arm64":  {"https://downloads.sourceforge.net/project/litegapps/litegapps/arm64/30/lite/2024-10-21/LiteGapps-arm64-11.0-20241021-official.zip", "901fd830fe4968b6979f38169fe49ceb"},
			"arm":    {"https://downloads.sourceforge.net/project/litegapps/litegapps/arm/30/lite/2024-08-18/AUTO-LiteGapps-arm-11.0-20240818-official.zip", "d4b2471d94facc13c9e7a026f2dff80d"},
		},
		"10.0.0": {
			"x86_64": {"https://downloads.sourceforge.net/project/litegapps/litegapps/x86_64/29/lite/v2.6/%5BAUTO%5DLiteGapps_x86_64_10.0_v2.6_official.zip", "d2d70e3e59149e23bdc8975dd6fa49e1"},
			"x86":    {"https://downloads.sourceforge.net/project/litegapps/litegapps/x86/29/lite/v2.6/%5BAUTO%5DLiteGapps_x86_10.0_v2.6_official.zip", "14e20a4628dc3198bbe79774cb1c33dc"},
			"arm64":  {"https://downloads.sourceforge.net/project/litegapps/litegapps/arm64/29/lite/2024-10-22/LiteGapps-arm64-10.0-20241022-official.zip", "0d079569cb5e2687939993776abb538c"},
			"arm":    {"https://downloads.sourceforge.net/project/litegapps/litegapps/arm/29/lite/2024-08-18/AUTO-LiteGapps-arm-10.0-20240818-official.zip", "a467f73d2b5a1ff9882d070989db0f0e"},
		},
		"9.0.0": {
			"x86_64": {"https://downloads.sourceforge.net/project/litegapps/litegapps/x86_64/28/lite/v2.6/%5BAUTO%5DLiteGapps_x86_64_9.0_v2.6_official.zip", "fc17a35518af188015baf1a682eb9fc7"},
			"x86":    {"https://downloads.sourceforge.net/project/litegapps/litegapps/x86/28/lite/v2.6/%5BAUTO%5DLiteGapps_x86_9.0_v2.6_official.zip", "31981cd14199d6b3610064b09d96e278"},
			"arm64":  {"https://downloads.sourceforge.net/project/litegapps/litegapps/arm64/28/lite/2024-02-23/AUTO-LiteGapps-arm64-9.0-20240223-official.zip", "b8ccfbedbf003803af19346c610988c0"},
			"arm":    {"https://downloads.sourceforge.net/project/litegapps/litegapps/arm/28/lite/%5BAUTO%5DLiteGapps_arm_9.0_v2.5_official.zip", "8034245b695b6b31cd6a5d2ed5b2b670"},
		},
		"8.1.0": {
			"x86_64": {"https://downloads.sourceforge.net/project/litegapps/litegapps/x86_64/27/lite/v2.6/%5BAUTO%5DLiteGapps_x86_64_8.1_v2.6_official.zip", "eee0ebdea5eb7580cab9dec307b46f56"},
			"x86":    {"https://downloads.sourceforge.net/project/litegapps/litegapps/x86/27/lite/v2.6/%5BAUTO%5DLiteGapps_x86_8.1_v2.6_official.zip", "5739feb54fdf85dc1d870998aeeee43a"},
			"arm64":  {"https://downloads.sourceforge.net/project/litegapps/litegapps/arm64/27/lite/2024-02-22/AUTO-LiteGapps-arm64-8.1-20240222-official.zip", "35d4195595961dc229f617c30c5460bb"},
			"arm":    {"https://downloads.sourceforge.net/project/litegapps/litegapps/arm/27/lite/%5BAUTO%5DLiteGapps_arm_8.1_v2.5_official.zip", "b0f7f5ba418b1696005f4e3f5abe924f"},
		},
	}

	apiLevelMap := map[string]string{
		"13.0.0": "33",
		"12.0.0": "31",
		"11.0.0": "30",
		"10.0.0": "29",
		"9.0.0":  "28",
		"8.1.0":  "27",
	}

	arch := getHostArch()

	baseAddon := NewBaseAddon("LiteGapps", AddonTypeGapps, versions)
	return &LiteGappsAddon{
		BaseAddon:   baseAddon,
		dlLinks:     dlLinks,
		apiLevelMap: apiLevelMap,
		extractTo:   "/tmp/litegapps/extract",
		currentArch: arch,
	}
}

func (l *LiteGappsAddon) Download(version, arch string, onStatus func(string)) error {

	versionLinks, ok := l.dlLinks[version]
	if !ok {
		return fmt.Errorf("no download links for version %s", version)
	}

	archLinks, ok := versionLinks[arch]
	if !ok {
		return fmt.Errorf("LiteGapps not available for architecture %s on Android %s", arch, version)
	}

	url := archLinks[0]
	filename := filepath.Join(l.downloadDir, "litegapps.zip")

	if err := ensureDir(l.downloadDir); err != nil {
		return err
	}

	onStatus(fmt.Sprintf("Downloading LiteGapps for Android %s (%s)...", version, arch))
	return downloadFile(url, filename)
}

func (l *LiteGappsAddon) Extract(version, arch string, onStatus func(string)) error {
	filename := filepath.Join(l.downloadDir, "litegapps.zip")

	if err := ensureDir(l.extractTo); err != nil {
		return err
	}

	onStatus("Extracting LiteGapps archive...")
	return extractZip(filename, l.extractTo)
}

func (l *LiteGappsAddon) Copy(version, arch, outputDir string, onStatus func(string)) error {
	copyDir := filepath.Join(outputDir, "litegapps")

	if err := os.RemoveAll(copyDir); err != nil {
		return err
	}

	if err := ensureDir(copyDir); err != nil {
		return err
	}

	appUnpackDir := filepath.Join(l.extractTo, "appunpack")
	if err := ensureDir(appUnpackDir); err != nil {
		return err
	}

	tarFile := filepath.Join(l.extractTo, "files", "files.tar.xz")
	onStatus("Extracting files.tar.xz...")

	cmd := exec.Command("tar", "-xvf", tarFile, "-C", appUnpackDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract tar.xz: %v", err)
	}

	apiLevel := l.apiLevelMap[version]
	srcDir := filepath.Join(appUnpackDir, arch, apiLevel, "system")
	dstDir := filepath.Join(copyDir, "system")

	onStatus("Copying LiteGapps files...")
	return copyDir2(srcDir, dstDir)
}

func (l *LiteGappsAddon) Install(version, arch, outputDir string, onStatus func(string)) error {
	if err := l.Download(version, arch, onStatus); err != nil {
		return err
	}
	if err := l.Extract(version, arch, onStatus); err != nil {
		return err
	}
	return l.Copy(version, arch, outputDir, onStatus)
}

func (l *LiteGappsAddon) DockerfileInstructions() string {
	return fmt.Sprintf("COPY %s /\n", "litegapps")
}

func getHostArch() string {
	machine := runtime.GOARCH

	mapping := map[string]string{
		"386":   "x86",
		"amd64": "x86_64",
		"arm64": "arm64",
		"arm":   "arm",
	}

	if arch, ok := mapping[machine]; ok {
		return arch
	}
	return "x86_64"
}
