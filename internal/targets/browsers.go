package targets

import (
	"github.com/bugraskl/sweepkit/internal/core"
)

type browser struct {
	id, name string
	procs    []string
	roots    []string
}

// Cache folders Chromium-based browsers keep per profile and per install.
// Everything else in the profile, such as cookies, history, passwords and
// sessions, is left alone.
var chromiumCaches = []string{
	"Cache", "Code Cache", "GPUCache", "ShaderCache", "GrShaderCache",
	"GraphiteDawnCache", "DawnGraphiteCache", "DawnWebGPUCache",
}

// chromium finds cache folders directly in each base folder and in its
// profile folders one level down.
func chromium(bases ...string) []string {
	var patterns []string
	for _, base := range nonEmpty(bases...) {
		for _, c := range chromiumCaches {
			patterns = append(patterns, join(base, c), join(base, "*", c))
		}
	}
	return dirs(patterns...)
}

func firefox(profileDirs ...string) []string {
	var patterns []string
	for _, d := range nonEmpty(profileDirs...) {
		patterns = append(patterns, join(d, "*", "cache2"))
	}
	return dirs(patterns...)
}

func browsers(p Paths) []browser {
	if p.OS == "windows" {
		local, roaming := p.LocalAppData, p.RoamingAppData
		return []browser{
			{"chrome", "Google Chrome", []string{"chrome.exe"}, chromium(join(local, "Google", "Chrome", "User Data"))},
			{"edge", "Microsoft Edge", []string{"msedge.exe"}, chromium(join(local, "Microsoft", "Edge", "User Data"))},
			{"brave", "Brave", []string{"brave.exe"}, chromium(join(local, "BraveSoftware", "Brave-Browser", "User Data"))},
			{"vivaldi", "Vivaldi", []string{"vivaldi.exe"}, chromium(join(local, "Vivaldi", "User Data"))},
			{"opera", "Opera", []string{"opera.exe"}, chromium(
				join(local, "Opera Software", "Opera Stable"), join(roaming, "Opera Software", "Opera Stable"),
				join(local, "Opera Software", "Opera GX Stable"), join(roaming, "Opera Software", "Opera GX Stable"),
			)},
			{"firefox", "Firefox", []string{"firefox.exe"}, firefox(join(local, "Mozilla", "Firefox", "Profiles"))},
		}
	}
	cache, config, home := p.CacheHome, p.ConfigHome, p.Home
	return []browser{
		{"chrome", "Google Chrome", []string{"chrome"}, chromium(join(cache, "google-chrome"), join(config, "google-chrome"))},
		{"chromium", "Chromium", []string{"chromium", "chromium-browse", "chrome"}, chromium(
			join(cache, "chromium"), join(config, "chromium"),
			join(home, "snap", "chromium", "common", ".cache", "chromium"), join(home, "snap", "chromium", "common", "chromium"),
		)},
		{"edge", "Microsoft Edge", []string{"msedge"}, chromium(join(cache, "microsoft-edge"), join(config, "microsoft-edge"))},
		{"brave", "Brave", []string{"brave"}, chromium(join(cache, "BraveSoftware", "Brave-Browser"), join(config, "BraveSoftware", "Brave-Browser"))},
		{"vivaldi", "Vivaldi", []string{"vivaldi-bin", "vivaldi"}, chromium(join(cache, "vivaldi"), join(config, "vivaldi"))},
		{"opera", "Opera", []string{"opera"}, chromium(join(cache, "opera"), join(config, "opera"))},
		{"firefox", "Firefox", []string{"firefox", "firefox-bin"}, firefox(
			join(cache, "mozilla", "firefox"),
			join(home, "snap", "firefox", "common", ".cache", "mozilla", "firefox"),
			join(home, ".var", "app", "org.mozilla.firefox", "cache", "mozilla", "firefox"),
		)},
	}
}

func browserTargets(p Paths) []*core.Target {
	var ts []*core.Target
	for _, b := range browsers(p) {
		ts = append(ts, &core.Target{
			ID:       "browser-" + b.id,
			Category: core.Browser,
			Name:     txt(b.name+" cache", b.name+" önbelleği"),
			Description: txt(
				"Page resources and compiled scripts the browser saved to load sites faster. Cookies, history, passwords and open tabs are not touched.",
				"Tarayıcının siteleri daha hızlı açmak için kaydettiği sayfa kaynakları ve derlenmiş betikler. Çerezler, geçmiş, şifreler ve açık sekmeler silinmez.",
			),
			Caveat: txt(
				"Sites load a little slower the first time you visit them again.",
				"Siteler bir sonraki ilk ziyaretinizde biraz daha yavaş açılır.",
			),
			Roots:     b.roots,
			BlockedBy: b.procs,
		})
	}
	return ts
}
