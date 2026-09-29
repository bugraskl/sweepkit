package targets

import (
	"regexp"

	"github.com/bugraskl/sweepkit/internal/core"
)

var (
	// The Go build cache keeps its entries in folders named 00 to ff.
	goCacheLayout = regexp.MustCompile(`^[0-9a-f]{2}$`)
	// Yarn 1 keeps its cache in a folder named after the cache format, e.g. v6.
	yarnCacheLayout = regexp.MustCompile(`^v[0-9]+$`)
)

func developerTargets(p Paths) []*core.Target {
	getenv := p.Getenv
	if getenv == nil {
		getenv = func(string) string { return "" }
	}
	win := p.OS == "windows"
	pick := func(windows, linux string) string {
		if win {
			return windows
		}
		return linux
	}
	caveat := txt(
		"The next install or build downloads or compiles these again.",
		"Sonraki kurulum ya da derleme bunları yeniden indirir veya derler.",
	)

	npm := firstAbs(getenv("npm_config_cache"), getenv("NPM_CONFIG_CACHE"), pick(join(p.LocalAppData, "npm-cache"), join(p.Home, ".npm")))
	pip := firstAbs(getenv("PIP_CACHE_DIR"), pick(join(p.LocalAppData, "pip", "Cache"), join(p.CacheHome, "pip")))
	yarn := firstAbs(getenv("YARN_CACHE_FOLDER"), pick(join(p.LocalAppData, "Yarn", "Cache"), join(p.CacheHome, "yarn")))
	nuget := pick(join(p.LocalAppData, "NuGet"), join(p.DataHome, "NuGet"))

	return []*core.Target{
		{
			ID: "npm-cache", Category: core.Developer, MinAge: hour,
			Name: txt("npm cache", "npm önbelleği"),
			Description: txt(
				"Package archives npm keeps to speed up installs, like `npm cache clean --force`.",
				"npm'in kurulumları hızlandırmak için sakladığı paket arşivleri; `npm cache clean --force` gibi.",
			),
			Caveat: caveat,
			// Only the _cacache folder: a misconfigured cache path can never
			// widen the target beyond npm's own layout.
			Roots: nonEmpty(join(npm, "_cacache")),
		},
		{
			ID: "pip-cache", Category: core.Developer, MinAge: hour,
			Name: txt("pip cache", "pip önbelleği"),
			Description: txt(
				"Python packages pip downloaded or built, like `pip cache purge`.",
				"pip'in indirdiği veya derlediği Python paketleri; `pip cache purge` gibi.",
			),
			Caveat: caveat,
			Roots:  nonEmpty(join(pip, "http"), join(pip, "http-v2"), join(pip, "wheels")),
		},
		{
			ID: "go-build-cache", Category: core.Developer, MinAge: hour,
			Name: txt("Go build cache", "Go derleme önbelleği"),
			Description: txt(
				"Compiled packages the Go toolchain keeps, like `go clean -cache`.",
				"Go araç zincirinin sakladığı derlenmiş paketler; `go clean -cache` gibi.",
			),
			Caveat: caveat,
			Roots:  nonEmpty(p.GoCache),
			Match:  firstDirMatches(goCacheLayout),
		},
		{
			ID: "yarn-cache", Category: core.Developer, MinAge: hour,
			Name: txt("Yarn cache", "Yarn önbelleği"),
			Description: txt(
				"Packages cached by Yarn 1, like `yarn cache clean`.",
				"Yarn 1'in önbelleğe aldığı paketler; `yarn cache clean` gibi.",
			),
			Caveat: caveat,
			Roots:  nonEmpty(yarn),
			Match:  firstDirMatches(yarnCacheLayout),
		},
		{
			ID: "nuget-http-cache", Category: core.Developer, MinAge: hour,
			Name: txt("NuGet HTTP cache", "NuGet HTTP önbelleği"),
			Description: txt(
				"Package downloads cached by NuGet and dotnet, like `dotnet nuget locals http-cache --clear`.",
				"NuGet ve dotnet'in önbelleğe aldığı paket indirmeleri; `dotnet nuget locals http-cache --clear` gibi.",
			),
			Caveat: caveat,
			Roots:  nonEmpty(firstAbs(getenv("NUGET_HTTP_CACHE_PATH")), join(nuget, "v3-cache"), join(nuget, "http-cache")),
			// NuGet stores every cached response as a .dat file.
			Match: hasExt(".dat"),
		},
	}
}
