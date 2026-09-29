package targets

import (
	"path"
	"strings"

	"github.com/bugraskl/sweepkit/internal/core"
)

func linuxTargets(p Paths) []*core.Target {
	return []*core.Target{
		{
			ID: "thumbnails", Category: core.System, Default: true, MinAge: month,
			Name: txt("Thumbnail cache", "Küçük resim önbelleği"),
			Description: txt(
				"Previews file managers make for photos and videos, older than 30 days.",
				"Dosya yöneticilerinin fotoğraf ve videolar için ürettiği önizlemeler; 30 günden eski olanlar.",
			),
			Roots: nonEmpty(join(p.CacheHome, "thumbnails")),
		},
		{
			ID: "apt-cache", Category: core.Packages, Admin: true, Default: true,
			Name: txt("APT package cache", "APT paket önbelleği"),
			Description: txt(
				"Downloaded .deb files of packages that are already installed. Runs `apt-get clean`.",
				"Zaten kurulu paketlerin indirilmiş .deb dosyaları. `apt-get clean` çalıştırılır.",
			),
			Roots:   []string{"/var/cache/apt/archives"},
			Match:   hasExt(".deb"),
			Command: []string{"apt-get", "clean"},
		},
		{
			ID: "dnf-cache", Category: core.Packages, Admin: true, Default: true,
			Name: txt("DNF package cache", "DNF paket önbelleği"),
			Description: txt(
				"Downloaded .rpm files of packages that are already installed. Runs `dnf clean packages`.",
				"Zaten kurulu paketlerin indirilmiş .rpm dosyaları. `dnf clean packages` çalıştırılır.",
			),
			Roots:   []string{"/var/cache/dnf", "/var/cache/libdnf5"},
			Match:   hasExt(".rpm"),
			Command: []string{"dnf", "clean", "packages"},
		},
		{
			ID: "pacman-cache", Category: core.Packages, Admin: true, Default: true,
			Name: txt("Pacman package cache", "Pacman paket önbelleği"),
			Description: txt(
				"Old package versions in the pacman cache. The two newest versions of each package are kept, like `paccache -rk2`.",
				"pacman önbelleğindeki eski paket sürümleri. Her paketin en yeni iki sürümü tutulur; `paccache -rk2` gibi.",
			),
			Caveat: txt(
				"Downgrading to a removed version means downloading it again.",
				"Silinen bir sürüme geri dönmek için onu yeniden indirmeniz gerekir.",
			),
			Roots:     []string{"/var/cache/pacman/pkg"},
			Match:     func(rel string) bool { return strings.Contains(path.Base(rel), ".pkg.tar.") },
			Select:    keepNewest(2),
			BlockedBy: []string{"pacman"},
		},
		{
			ID: "journal", Category: core.Logs, Admin: true, Default: true, MinAge: month,
			Name: txt("Old system logs", "Eski sistem günlükleri"),
			Description: txt(
				"Archived systemd journal files older than 30 days. Runs `journalctl --vacuum-time=30d`.",
				"30 günden eski, arşivlenmiş systemd günlük dosyaları. `journalctl --vacuum-time=30d` çalıştırılır.",
			),
			Roots: []string{"/var/log/journal"},
			Match: func(rel string) bool {
				base := path.Base(rel)
				return strings.Contains(base, "@") && (strings.HasSuffix(base, ".journal") || strings.HasSuffix(base, ".journal~"))
			},
			Command: []string{"journalctl", "--vacuum-time=30d"},
		},
		{
			ID: "crash-dumps", Category: core.Logs, Admin: true, Default: true, MinAge: week,
			Name: txt("Crash dumps", "Çökme dökümleri"),
			Description: txt(
				"Core dumps systemd saved when programs crashed, older than 7 days.",
				"Programlar çöktüğünde systemd'nin kaydettiği bellek dökümleri; 7 günden eski olanlar.",
			),
			Roots: []string{"/var/lib/systemd/coredump"},
		},
	}
}
