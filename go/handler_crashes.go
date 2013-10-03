package main

import (
	"atom"
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var blacklistedSumatraVersions = []string{"1.5.1", "1.6", "1.7", "1.8", "1.9",
	"2.0", "2.0.1", "2.1", "2.1.1", "2.2", "2.2.1", "2.3", "2.3.1", "2.3.2"}

func (c *Crash) Version() string {
	ver := *c.ProgramVersion
	if ver == "" {
		return "no ver"
	}
	if strings.HasSuffix(ver, " pre-release") {
		return ver[:len(ver)-len(" pre-release")]
	}
	return ver
}

func (c *Crash) CreatedOnSince() string {
	return TimeSinceNowAsString(c.CreatedOn)
}

func (c *Crash) ShortCrashingLine() string {
	s := *c.CrashingLine
	if len(s) <= 60 {
		return s
	}
	return s[:56] + "..."
}

func (c *Crash) ShortIpAddr() string {
	s := c.IpAddress()
	if len(s) <= 16 {
		return s
	}
	return s[:13] + "..."
}

type CrashesForDay struct {
	Day     string
	Crashes []*Crash
}

func (c *CrashesForDay) CrashesCount() int {
	return len(c.Crashes)
}

type AppDisplay struct {
	*App
	Days []CrashesForDay
}

// Reverse embeds a sort.Interface value and implements a reverse sort over
// that value.
type Reverse struct {
	// This embedded Interface permits Reverse to use the methods of
	// another Interface implementation.
	sort.Interface
}

// Less returns the opposite of the embedded implementation's Less method.
func (r Reverse) Less(i, j int) bool {
	return r.Interface.Less(j, i)
}

func CanSeeCrashes(r *http.Request, app string) bool {
	user := getSecureCookie(r).TwitterUser
	if user == "kjk" {
		return true
	}
	if user == "zeniko_ch" && (app == "SumatraPDF" || app == "") {
		return true
	}
	return false
}

var notLoggedIn = `<html><body>Need to <a href="/login?redirect=%s">login</a> to see crashes.</body></html>`
var loggedInButNoAccess = `<html><body>You're logged in as %s. No access. <a href="/logout?redirect=%s">logout</a>`

func serveCrashLoginLogout(w http.ResponseWriter, r *http.Request) {
	url := url.QueryEscape(r.URL.Path + "?" + r.URL.RawQuery)
	user := getSecureCookie(r).TwitterUser
	if user == "" {
		fmt.Fprintf(w, notLoggedIn, url)
	} else {
		fmt.Fprintf(w, loggedInButNoAccess, user, url)
	}
}

func NewAppDisplay(app *App, addCrashesPerDay bool) *AppDisplay {
	res := &AppDisplay{App: app}
	if !addCrashesPerDay {
		res.Days = make([]CrashesForDay, 0)
		return res
	}

	n := len(app.PerDayCrashes)
	res.Days = make([]CrashesForDay, n, n)

	days := make([]string, n)
	i := 0
	for day, _ := range app.PerDayCrashes {
		days[i] = day
		i += 1
	}
	sort.Sort(Reverse{sort.StringSlice(days)})
	for i, day := range days {
		crashesForDay := CrashesForDay{Day: day}
		crashesForDay.Crashes = app.PerDayCrashes[day]
		res.Days[i] = crashesForDay
	}
	return res
}

func showCrashesIndex(w http.ResponseWriter, r *http.Request) {
	apps := storeCrashes.GetApps()
	model := struct {
		Apps []*App
		User string
	}{
		Apps: apps,
		User: getSecureCookie(r).TwitterUser,
	}
	ExecTemplate(w, tmplCrashReportsIndex, model)
}

func showCrashesByIp(w http.ResponseWriter, r *http.Request, app *App, ipAddrInternal string) {
	appDisplay := NewAppDisplay(app, false)
	crashes := storeCrashes.GetCrashesForIpAddrInternal(app, ipAddrInternal)
	model := struct {
		App         *AppDisplay
		ShowSince   bool
		Crashes     []*Crash
		DayOrIpAddr string
	}{
		App:         appDisplay,
		ShowSince:   true,
		Crashes:     crashes,
		DayOrIpAddr: crashes[0].IpAddress(),
	}
	ExecTemplate(w, tmplCrashReportsAppIndex, model)
}

func showCrashesByCrashingLine(w http.ResponseWriter, r *http.Request, app *App, crashingLine string) {
	appDisplay := NewAppDisplay(app, false)
	crashes := storeCrashes.GetCrashesForCrashingLine(app, crashingLine)
	model := struct {
		App         *AppDisplay
		ShowSince   bool
		Crashes     []*Crash
		DayOrIpAddr string
	}{
		App:         appDisplay,
		ShowSince:   true,
		Crashes:     crashes,
		DayOrIpAddr: crashingLine,
	}
	ExecTemplate(w, tmplCrashReportsAppIndex, model)
}

var tmplCrashesRss = template.Must(template.New("crashesrss.html").Parse(`
  <p>{{ len .Crashes }} crashes for {{ .Day }}:</p>
  {{ $appName := .AppName }}
  <table>
    {{ range .Crashes }}
      <tr>
        <td><a href="/app/crashshow?crash_id={{ .Id }}">{{ .Id }}</a></td>
        <td>{{ .Version }}</td>
        <td><a href="/app/crashes?app_name={{$appName}}&crashing_line={{.CrashingLine}}">{{ .ShortCrashingLine }}</a></td>
      </tr>
    {{ end }}
  </table>
`))

// /app/crashesrss?app_name=${appName}
func handleCrashesRss(w http.ResponseWriter, r *http.Request) {
	appName := getTrimmedFormValue(r, "app_name")
	app := storeCrashes.GetAppByName(appName)
	if app == nil {
		logger.Errorf("handleCrashesRss(): invalid app '%s'", appName)
		serve404(w, r)
		return
	}
	// to minimize the number of times rss reader updates the entries, we
	// don't show crashes for today i.e. if first day is the same time as
	// current time, we skip it
	appDisplay := NewAppDisplay(app, true)
	pubDate := time.Now()
	firstDayIdx := -1
	if len(appDisplay.Days) > 0 {
		firstDayIdx = 0
		todayDay := time.Now().Format("2006-01-02")
		if todayDay == appDisplay.Days[1].Day {
			firstDayIdx = 1
		}
	}
	if firstDayIdx != -1 {
		pubDate, _ = time.Parse("2006-01-02", appDisplay.Days[firstDayIdx].Day)
	}

	feed := &atom.Feed{
		Title:   fmt.Sprintf("Crashes %s", appName),
		Link:    fmt.Sprintf("http://blog.kowalczyk.info/app/crashesrss?app_name=%s", appName),
		PubDate: pubDate}
	baseUrl := fmt.Sprintf("http://blog.kowalczyk.info/app/crashes?app_name=%s", appName)
	if firstDayIdx == -1 {
		e := &atom.Entry{
			Title:       fmt.Sprintf("Crashes for %s", appName),
			Link:        baseUrl,
			ContentHtml: fmt.Sprintf("There are no crashes for %s yet", appName),
			PubDate:     pubDate}
		feed.AddEntry(e)
	} else {
		maxDays := 10
		for i := firstDayIdx; i < len(appDisplay.Days) && maxDays > 0; maxDays-- {
			day := appDisplay.Days[i].Day
			crashes := appDisplay.Days[i].Crashes
			model := struct {
				Crashes []*Crash
				Day     string
				AppName string
			}{
				Day:     day,
				AppName: appName,
				Crashes: crashes,
			}
			var buf bytes.Buffer
			tmplCrashesRss.Execute(&buf, model)
			html := string(buf.Bytes())
			pubDate, _ = time.Parse("2006-01-02", day)
			e := &atom.Entry{
				Title:       fmt.Sprintf("%d %s crashes on %s", len(crashes), appName, day),
				Link:        fmt.Sprintf("%s&day=%s", baseUrl, day),
				ContentHtml: html,
				PubDate:     pubDate}
			feed.AddEntry(e)
			i += 1
		}
	}

	s, err := feed.GenXml()
	if err != nil {
		s = "Failed to generate XML feed"
	}
	w.Write([]byte(s))
}

// /app/crashes[?app_name=${appName}][&day=${day}][&ip_addr=${ipAddrInternal}]
// [&crashing_line=${crashingLine}a]
func handleCrashes(w http.ResponseWriter, r *http.Request) {
	appName := getTrimmedFormValue(r, "app_name")
	if appName == "" {
		showCrashesIndex(w, r)
		return
	}
	if !CanSeeCrashes(r, "") {
		serveCrashLoginLogout(w, r)
		return
	}
	app := storeCrashes.GetAppByName(appName)
	if app == nil {
		logger.Errorf("handleCrashes(): invalid app '%s'", appName)
		serve404(w, r)
		return
	}

	ipAddrInternal := getTrimmedFormValue(r, "ip_addr")
	if ipAddrInternal != "" {
		showCrashesByIp(w, r, app, ipAddrInternal)
		return
	}
	crashingLine := getTrimmedFormValue(r, "crashing_line")
	if crashingLine != "" {
		showCrashesByCrashingLine(w, r, app, crashingLine)
		return
	}

	day := getTrimmedFormValue(r, "day")

	appDisplay := NewAppDisplay(app, true)
	var crashes []*Crash
	for _, forDay := range appDisplay.Days {
		if day == forDay.Day {
			crashes = forDay.Crashes
			break
		}
	}
	if crashes == nil {
		crashes = appDisplay.Days[0].Crashes
		day = appDisplay.Days[0].Day
	}
	model := struct {
		App         *AppDisplay
		ShowSince   bool
		Crashes     []*Crash
		DayOrIpAddr string
	}{
		App:         appDisplay,
		ShowSince:   false,
		Crashes:     crashes,
		DayOrIpAddr: day,
	}
	ExecTemplate(w, tmplCrashReportsAppIndex, model)
}

func readCrashReport(sha1 []byte) ([]byte, error) {
	return ReadFileAll(storeCrashes.MessageFilePath(sha1))
}

// /app/crashshow?crash_id=${crash_id}
func handleCrashShow(w http.ResponseWriter, r *http.Request) {
	if !CanSeeCrashes(r, "") {
		serveCrashLoginLogout(w, r)
		return
	}
	crashIdStr := getTrimmedFormValue(r, "crash_id")
	crashId, err := strconv.Atoi(crashIdStr)
	if err != nil {
		serve404(w, r)
		return
	}
	crash := storeCrashes.GetCrashById(crashId)
	if crash == nil {
		serve404(w, r)
		return
	}
	crashData, err := readCrashReport(crash.Sha1[:])
	if err != nil {
		serve404(w, r)
		return
	}
	appName := crash.App.Name
	crashBody := string(crashData)
	model := struct {
		IndexUrl  string
		IpAddr    string
		AppName   string
		CrashBody template.HTML
	}{
		IndexUrl:  fmt.Sprintf("/app/crashes?app_name=%s", appName),
		IpAddr:    crash.IpAddress(),
		AppName:   appName,
		CrashBody: template.HTML(crashBody),
	}
	ExecTemplate(w, tmplCrashReport, model)
}

// Version is in the format:
// "Ver: 2.1.1"
func extractSumatraVersion(crashData []byte) string {
	l := FindLineWithPrefix(crashData, "Ver: ")
	if l == nil {
		return ""
	}
	return string(l[5:])
}

// Version is in the format:
// "Version:         0.3.3 (0.3.3)"
func extractMacVersion(crashData []byte) string {
	l := FindLineWithPrefix(crashData, "Version:")
	if l == nil {
		return ""
	}
	s := string(l)
	parts := strings.SplitN(s, ":", 2)
	ver := strings.TrimSpace(parts[1])
	parts = strings.Split(ver, " ")
	return parts[0]
}

var macApps = []string{"VisualAck"}

func isMacApp(name string) bool {
	for _, n := range macApps {
		if n == name {
			return true
		}
	}
	return false
}

func extractAppVer(appName string, crashData []byte) string {
	if appName == "SumatraPDF" {
		return extractSumatraVersion(crashData)
	}

	if isMacApp(appName) {
		return extractMacVersion(crashData)
	}
	return ""
}

// we don't need to process crashes from old version, so blacklist specific
// versions
func shouldSaveCrash(app, ver string) bool {
	if app == "SumatraPDF" {
		for _, v := range blacklistedSumatraVersions {
			if v == ver {
				return false
			}
		}
		return true
	}
	return true
}

// POST /app/crashsubmit?appname=${appName}&file=${crashData}
func handleCrashSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		serveErrorMsg(w, "GET not supported")
		return
	}
	ipAddr := getIpAddress(r)
	appName := getTrimmedFormValue(r, "appname")
	if appName == "" {
		logger.Noticef("handleCrashSubmit(): 'appName' is not defined")
		return
	}
	crashDataFile, _, err := r.FormFile("file")
	if err != nil {
		logger.Noticef("handleCrashSubmit(): 'file' is not defined, err = %s", err.Error())
		return
	}

	crashData, err := ioutil.ReadAll(crashDataFile)
	if err != nil {
		logger.Noticef("handleCrashSubmit(): ioutil.ReadAll() failed with %s", err.Error())
		return
	}

	appVer := extractAppVer(appName, crashData)
	if shouldSaveCrash(appName, appVer) {
		err = storeCrashes.SaveCrash(appName, appVer, ipAddr, crashData)
		if err != nil {
			logger.Noticef("handleCrashSubmit(): storeCrashes.SaveCrash() failed with %s", err.Error())
			return
		}
		logger.Noticef("handleCrashSubmit(): %s %s %s", appName, appVer, ipAddr)
	}
	w.Write([]byte(""))
}
