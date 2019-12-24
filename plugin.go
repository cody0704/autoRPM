package main

import (
	"bufio"
	"errors"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"fmt"
	"html/template"
	"io"
	"os"

	strftime "github.com/lestrrat/go-strftime"
)

var (
	errMissingPackagingePath = errors.New("Error: missing rpm packaging path")
)

type rpmParms struct {
	RPMNAME          string
	ParsePath        string
	Packager         string
	PackagingPath    string
	ProjectName      string
	URL              string
	Vendor           string
	APINAME          string
	Version          string
	Release          string
	Requires         string
	ParseInstall     string
	ParseFiles       string
	ParsePosts       string
	ParseEchoInstall string
	ParseEchoINI     string
	ParseEchoLog     string
}

const rpmDirs = "/root/rpmbuild/RPMS"

const rpmTMPL = `
Name:           {{ .APINAME }}
Version:        {{ .Version }}
Release:        {{ .Release}}%{?dist}
Summary:        {{ .APINAME }}.
URL:            {{ .URL }}
Vendor:         {{ .Vendor }}
License:        Copyright
Packager:       {{ .Packager }}
BuildArch:      x86_64
Requires:       {{ .Requires }}

%description
This is {{ .APINAME }} for {{ .ProjectName }}.

%prep
cp -r %{_sourcedir}/{{ .PackagingPath }} %{_builddir}/

%install
{{ .ParseInstall }}

mkdir -p %{buildroot}/home/crontab/
cp -r {{ .ParsePath }}/crontab_setting.sh %{buildroot}/home/crontab/

%files
{{ .ParseFiles }}
/home/crontab/

%clean
rm -rf %{buildroot}

%pre
echo "This will install {{ .APINAME }}"

%post
{{ .ParsePosts }}
echo "################################################"
echo "Note: information"
echo "1.install path: 
{{ .ParseEchoInstall }}"
echo "2.ini path: 
{{ .ParseEchoINI }}"
echo "3.logs path: 
{{ .ParseEchoLog }}"
echo "4.RUN: sh /home/crontab/crontab_setting.sh (setup crontab)"
sh /home/crontab/crontab_setting.sh
echo "################################################"
`

type (
	// Config for the plugin.
	Config struct {
		RPMNAME       string
		PackagePath   string
		Packager      string
		PackagingPath string
		ProjectName   string
		URL           string
		Vendor        string
		APINAME       string
		Version       string
		Requires      string
		GitEnable     bool
		Reponstiry    string
		GitToken      string
	}
	// Plugin structure
	Plugin struct {
		Config Config
		Writer io.Writer
	}
)

func (p Plugin) exec() {
	var parseInstall string
	var parseFiles string
	var parsePosts string
	var parseEchoInstall string
	var parseEchoINI string
	var parseEchoLog string

	var echoInstall []string
	var files []string
	var posts []string

	var reponstiryPath string
	reponstiryPath, _ = parseRPM(`([\w-]+\/[\w-]+)$`, p.Config.Reponstiry, false)

	parsePath, _ := parseRPM(`(\w+$)`, p.Config.PackagingPath, false)

	p.log("Reponstiry:")
	p.log("====================")
	p.log(p.Config.Reponstiry)
	p.log(reponstiryPath)
	p.log("====================")

	dirs := []string{"SPECS", "BUILD", "BUILDROOT", "RPMS", "SRPMS"}

	for _, dir := range dirs {
		p.cmd("mkdir: /github/home/rpmbuild/"+dir,
			"mkdir", "-p", "/github/home/rpmbuild/"+dir)
	}

	p.cmd("ls:/github/home/rpmbuild/",
		"ls", "-al", "/github/home/rpmbuild/")

	gitRepo := "https://" + p.Config.GitToken + ":x-oauth-basic@github.com/" + reponstiryPath + ".git"

	p.cmd("Exec:",
		"git", "clone", gitRepo, "/github/home/rpmbuild/SOURCES")

	if len(p.Config.PackagePath) == 0 {
		p.Config.PackagePath = "."
	}

	list, err := getDirList(p.Config.PackagePath + p.Config.PackagingPath)
	if err != nil {
		log.Fatal(err)
	}

	type rpmSetting struct {
		file string
		path string
	}

	var getPath []rpmSetting
	for _, v := range list {
		prog, _ := os.Open(v)
		rd := bufio.NewReader(prog)
	parse:
		for {
			line, err := rd.ReadString('\n')
			if err != nil || err == io.EOF {
				break
			}

			switch {
			case strings.Contains(line, "configfile"):
				if matchPath, ok := parseRPM(`[\'\"]([\w\/\.]+)[\'\"]`, line, true); ok {
					matchFile, _ := parseRPM(`services\/([\w\W]+)\/\w+\.\w+$`, v, false)
					getPath = append(getPath, rpmSetting{file: matchFile, path: matchPath})
					break parse
				}

			case strings.Contains(line, "cfg"):
				if matchPath, ok := parseRPM(`\([\'\"]([\w\/\.]+)[\'\"]\)`, line, true); ok {
					matchFile, _ := parseRPM(`services\/([\w\W]+)\/\w+\.\w+$`, v, false)
					getPath = append(getPath, rpmSetting{file: matchFile, path: matchPath})
					break parse
				}

			case strings.Contains(line, "log_dir"):
				if matchPath, ok := parseRPM(`[\=,\s]+([\w\/\.]+)`, line, true); ok {
					getPath = append(getPath, rpmSetting{path: matchPath})
					break parse
				}
			}
		}
	}

	for _, data := range getPath {
		files = append(files, data.path)
		if data.file != "" {
			parseInstall = parseInstall + "mkdir -p %{buildroot}" + data.path + "/\n"

			parseInstall = parseInstall + fmt.Sprintf("cp %s/*.pl %s%s/\n", data.file, "%{buildroot}", data.path)
			parseInstall = parseInstall + fmt.Sprintf("cp %s/*.ini %s%s/\n", data.file, "%{buildroot}", data.path)

			posts = append(posts, "chmod 770 "+data.path+"/*.pl\n")
			echoInstall = append(echoInstall, data.path+"\n")
			parseEchoINI = parseEchoINI + data.path + "/" + data.file + "/" + data.file + ".ini" + "\n"
		} else {
			parseInstall = parseInstall + "mkdir -p %{buildroot}" + data.path + "/\n"
			parseEchoLog = parseEchoLog + data.path + "/\n"
		}
	}
	echoInstall = removeRepByLoop(echoInstall)
	files = removeRepByLoop(files)
	posts = removeRepByLoop(posts)

	for _, echo := range echoInstall {
		parseEchoInstall = parseEchoInstall + echo
	}

	for _, file := range files {
		parseFiles = parseFiles + file + "/\n"
	}

	for _, post := range posts {
		parsePosts = parsePosts + post
	}

	strfobj, _ := strftime.New("%Y%m%d%H%M%S")
	release := strfobj.FormatString(time.Now())

	autoRPM := rpmParms{
		ParsePath:        parsePath,
		ProjectName:      p.Config.ProjectName,
		PackagingPath:    p.Config.PackagingPath,
		Packager:         p.Config.Packager,
		URL:              p.Config.URL,
		Vendor:           p.Config.Vendor,
		APINAME:          p.Config.APINAME,
		RPMNAME:          p.Config.RPMNAME,
		Version:          p.Config.Version,
		Release:          release,
		Requires:         p.Config.Requires,
		ParseInstall:     parseInstall,
		ParseFiles:       parseFiles,
		ParsePosts:       parsePosts,
		ParseEchoInstall: parseEchoInstall,
		ParseEchoINI:     parseEchoINI,
		ParseEchoLog:     parseEchoLog,
	}

	tmpl := template.New("rpm_spec")
	tmpl.Parse(rpmTMPL)

	specFile, _ := os.OpenFile("/github/home/rpmbuild/SPECS/"+parsePath+".spec", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 777)
	defer specFile.Close()

	tmpl.Execute(specFile, autoRPM)

	if p.Config.GitEnable {
		p.cmd("ls:/github/home/rpmbuild/SOURCES/",
			"ls", "-al", "/github/home/rpmbuild/SOURCES")

		p.cmd("cat:"+"/github/home/rpmbuild/SPECS/"+parsePath+".spec",
			"cat", "/github/home/rpmbuild/SPECS/"+parsePath+".spec")

		p.cmd("build:",
			"rpmbuild", "-ba", "/github/home/rpmbuild/SPECS/"+parsePath+".spec")

		os.Chdir("/github/home/rpmbuild/SOURCES")

		p.cmd("git config user.email",
			"git", "config", "--local", "user.email", "action@github.com")

		p.cmd("git config user.name",
			"git", "config", "--local", "user.name", "GitHub Action")

		p.cmd("git config push.default",
			"git", "config", "--global", "push.default", "simple")

		p.cmd("git checkout rpm",
			"git", "checkout", "-b", "rpm")

		p.cmd("git pull",
			"git", "pull", gitRepo, "rpm")

		p.cmd("rm -rf .github",
			"rm", "-rf", ".github")

		p.cmd("cp -r /github/home/rpmbuild/RPMS /github/home/rpmbuild/SOURCES/",
			"cp", "-r", "/github/home/rpmbuild/RPMS", "/github/home/rpmbuild/SOURCES/")

		rpmFile := fmt.Sprintf("%s-%s-%s.el7.x86_64.rpm", p.Config.APINAME, p.Config.Version, release)
		rpmLatestFile := fmt.Sprintf("%s-latest.rpm", p.Config.RPMNAME)

		p.cmd("cp /github/home/rpmbuild/RPMS /github/home/rpmbuild/SOURCES/",
			"cp", "/github/home/rpmbuild/RPMS/x86_64/"+rpmFile, "/github/home/rpmbuild/SOURCES/RPMS/x86_64/"+rpmLatestFile)

		p.cmd("git add .",
			"git", "add", "-A", ".")

		p.cmd("git commit",
			"git", "commit", "-m", "\"RPM build latest\"", "-a")

		p.cmd("git push",
			"git", "push", "--set-upstream", gitRepo, "rpm")
	}
}

// Exec executes the plugin.
func (p Plugin) Exec() error {
	if len(p.Config.PackagingPath) == 0 {
		return errMissingPackagingePath
	}

	fmt.Println("================================================")
	fmt.Println("✅ Successfully executed commands packaging rpm.")
	fmt.Println("================================================")

	p.exec()

	return nil
}

func (p Plugin) log(message ...interface{}) {
	if p.Writer == nil {
		p.Writer = os.Stdout
	}

	fmt.Fprintf(p.Writer, "%s", fmt.Sprintln(message...))
}

func getDirList(dirpath string) ([]string, error) {
	var dirList []string
	err := filepath.Walk(dirpath,
		func(path string, f os.FileInfo, err error) error {
			if f == nil {
				return err
			}
			if !f.IsDir() {
				switch filepath.Ext(path) {
				case ".ini":
					dirList = append(dirList, path)
				case ".pl":
					dirList = append(dirList, path)
				}

				return nil
			}

			return nil
		})
	return dirList, err
}

func removeRepByLoop(slc []string) []string {
	result := []string{}
	for i := range slc {
		flag := true
		for j := range result {
			if slc[i] == result[j] {
				flag = false
				break
			}
		}
		if flag {
			result = append(result, slc[i])
		}
	}
	return result
}

func parseRPM(pattern string, s string, is bool) (string, bool) {
	match := regexp.MustCompile(pattern).FindStringSubmatch(s)
	if match != nil {
		if is {
			return filepath.Dir(match[1]), true
		}

		return match[1], true
	}
	return "", false
}

func (p Plugin) cmd(info string, command string, options ...string) error {
	cmd := exec.Command(command, options...)
	out, err := cmd.CombinedOutput()
	p.log(info)
	p.log("====================")
	p.log(string(out))
	if err != nil && info != "git pull" {
		log.Fatal(err)
	}
	p.log("====================")

	return nil
}
