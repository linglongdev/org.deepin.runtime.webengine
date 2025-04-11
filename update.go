package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"
)

var (
	deepinRepoURL = "http://10.20.64.92:8080/crimson_runtime/stable_20260415"
	uosRepoURL    = "https://pools.uniontech.com/desktop-professional-V25"

	deepinCodename = "stable"
	uosCodename    = "snipe/2500u1"

	uosArch = []string{"mips64", "sw64"}

	sources = []string{
		"qt6-base",
		"qt6-svg",
		"qt6-declarative",
		"qt6-imageformats",
		"qt6-multimedia",
		"qt6-speech",
		"qt6-tools",
		"qt6-wayland",
		"qt6-translations",
		"qt6-5compat",

		"qt6-webengine",
		"qt6-webchannel",

		"dtkcore",
		"dtkdeclarative",
		"dtkgui",
		"dtkwidget",
		"dtklog",
		"dtkcommon",
		"qt5integration",
		"dde-qt5platform-plugins",

		"fcitx5-qt",
		"deepin-shortcut-viewer",
	}
)

func main() {
	writer := &bytes.Buffer{}
	excludePackage := func(pkg string) {
		fmt.Fprintf(os.Stderr, "  Exclude Binary %s\n", pkg)
		// fmt.Fprintf(writer, "  # linglong:gen_deb_source exclude %s\n", pkg)
	}
	installPackage := func(pkg string) {
		fmt.Fprintf(os.Stderr, "  Install Binary %s\n", pkg)
		fmt.Fprintf(writer, "  # linglong:gen_deb_source install %s\n", pkg)
	}

	for _, src := range sources {
		fmt.Fprintf(os.Stderr, "Source %s\n", src)
		fmt.Fprintf(writer, "  # source package %s\n", src)

		// 确定目录结构 (lib* 开头取前 4 位，否则取首字母)
		dir := src[:1]
		if strings.HasPrefix(src, "lib") && len(src) >= 4 {
			dir = src[:4]
		}

		// 获取网页内容并解析二进制包名
		fullURL := fmt.Sprintf("%s/pool/main/%s/%s/", deepinRepoURL, dir, src)
		pkgs, err := fetchBinaryPackages(fullURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Error fetching %s: %v\n", src, err)
			continue
		}

		for _, pkg := range pkgs {
			// 1. 通用过滤：排除 dbgsym, doc, examples
			if strings.HasSuffix(pkg, "dbgsym") {
				continue
			}
			if strings.HasSuffix(pkg, "-examples") {
				excludePackage(pkg)
				continue
			}
			if strings.HasSuffix(pkg, "-doc") || strings.Contains(pkg, "-doc-") {
				excludePackage(pkg)
				continue
			}

			// 过滤dtk5
			if src != "dtkcommon" {
				if strings.Contains(pkg, "dtk") && !strings.Contains(pkg, "dtk6") {
					excludePackage(pkg)
					continue
				}
			}
			if src == "dtkdeclarative" {
				if strings.HasPrefix(pkg, "qml-module") {
					excludePackage(pkg)
					continue
				}
			}
			if strings.HasPrefix(pkg, "dde-") && strings.Contains(pkg, "qt5") {
				excludePackage(pkg)
				continue
			}

			if slices.Contains([]string{"fcitx5-frontend-qt5", "libfcitx5-qt-dev", "libfcitx5-qt1"}, pkg) {
				excludePackage(pkg)
				continue
			}

			installPackage(pkg)
		}
	}
	// 更新各架构的 yaml 文件
	updateYamlFiles(writer.Bytes())
}

var re = regexp.MustCompile(`href="([^" \n]+)`)

// 从仓库获取包列表
func fetchBinaryPackages(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// 正则匹配 <a href="pkgname_version_arch.deb">
	matches := re.FindAllStringSubmatch(string(body), -1)
	result := []string{}
	for i := range matches {
		filename := matches[i][1]
		if !strings.Contains(filename, ".deb") {
			continue
		}
		fields := strings.Split(filename, "_")
		result = append(result, fields[0])
	}
	sort.Strings(result)
	result = slices.Compact(result)
	return result, nil
}

func updateYamlFiles(content []byte) {
	yamlFiles := []string{
		"linglong.yaml",
		"arm64/linglong.yaml",
		"loong64/linglong.yaml",
		"sw64/linglong.yaml",
		"mips64/linglong.yaml",
	}

	for _, file := range yamlFiles {
		fields := strings.Split(file, "/")
		var arch string
		if len(fields) == 2 {
			arch = fields[0]
		} else {
			arch = "amd64"
		}

		original, err := os.ReadFile(file)
		if err != nil {
			continue // 如果文件不存在则跳过
		}

		lines := strings.Split(string(original), "\n")
		var newLines []string
		for _, line := range lines {
			if strings.Contains(line, "linglong:gen_deb_source sources") {
				break // 找到锚点，停止读取旧内容
			}
			newLines = append(newLines, line)
		}
		if slices.Contains(uosArch, arch) {
			if arch == "mips64" {
				arch = "mips64el"
			}
			source := fmt.Sprintf("  # linglong:gen_deb_source sources %s %s %s main community", arch, uosRepoURL, uosCodename)
			newLines = append(newLines, source)
		} else {
			source := fmt.Sprintf("  # linglong:gen_deb_source sources %s %s %s main", arch, deepinRepoURL, deepinCodename)
			newLines = append(newLines, source)
		}
		newContent := strings.Join(newLines, "\n") + "\n" + string(content)

		newContent = strings.Replace(newContent, "id: org.deepin.runtime.dtk", "id: org.deepin.runtime.webengine", 1)

		err = os.WriteFile(file, []byte(newContent), 0644)
		if err != nil {
			panic(err)
		}
	}
}
