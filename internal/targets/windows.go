package targets

import "github.com/bugraskl/sweepkit/internal/core"

func windowsTargets(p Paths) []*core.Target {
	wer := func(base string) []string {
		return nonEmpty(
			join(base, "Microsoft", "Windows", "WER", "ReportArchive"),
			join(base, "Microsoft", "Windows", "WER", "ReportQueue"),
		)
	}
	return []*core.Target{
		{
			ID: "temp", Category: core.System, Default: true, MinAge: day,
			Name: txt("Temporary files", "Geçici dosyalar"),
			Description: txt(
				"Files programs left behind in your temp folder. Only files untouched for 24 hours are removed.",
				"Programların geçici klasörünüzde bıraktığı dosyalar. Yalnızca 24 saattir dokunulmamış olanlar silinir.",
			),
			Roots: p.UserTemp,
		},
		{
			ID: "system-temp", Category: core.System, Admin: true, Default: true, MinAge: day,
			Name: txt("System temporary files", "Sistem geçici dosyaları"),
			Description: txt(
				`Leftovers from installers and services in Windows\Temp, older than 24 hours.`,
				`Kurulum programlarının ve servislerin Windows\Temp klasöründe bıraktıkları; 24 saatten eski olanlar.`,
			),
			Roots: nonEmpty(join(p.SystemRoot, "Temp")),
		},
		{
			ID: "windows-update", Category: core.System, Admin: true, MinAge: 10 * day,
			Name: txt("Windows Update downloads", "Windows Update indirmeleri"),
			Description: txt(
				"Update packages downloaded more than 10 days ago. Updates that are already installed do not need them.",
				"10 günden önce indirilmiş güncelleme paketleri. Kurulmuş güncellemelerin bunlara ihtiyacı yoktur.",
			),
			Caveat: txt(
				"If an update is still waiting to be installed, Windows downloads it again.",
				"Kurulmayı bekleyen bir güncelleme varsa Windows onu yeniden indirir.",
			),
			Roots: nonEmpty(join(p.SystemRoot, "SoftwareDistribution", "Download")),
		},
		{
			ID: "crash-dumps", Category: core.Logs, Default: true, MinAge: week,
			Name: txt("Crash dumps", "Çökme dökümleri"),
			Description: txt(
				"Memory dumps saved when a program crashed, older than 7 days. They are only useful for reporting that crash.",
				"Bir program çöktüğünde kaydedilen bellek dökümleri; 7 günden eski olanlar. Yalnızca o çökmeyi raporlamak için işe yarar.",
			),
			Roots: nonEmpty(join(p.LocalAppData, "CrashDumps")),
		},
		{
			ID: "system-crash-dumps", Category: core.Logs, Admin: true, Default: true, MinAge: week,
			Name: txt("System crash dumps", "Sistem çökme dökümleri"),
			Description: txt(
				"Dumps from blue screens and driver resets, older than 7 days. A full memory dump can take several gigabytes.",
				"Mavi ekran ve sürücü sıfırlamalarından kalan dökümler; 7 günden eski olanlar. Tam bellek dökümü birkaç GB tutabilir.",
			),
			Roots: nonEmpty(join(p.SystemRoot, "Minidump"), join(p.SystemRoot, "LiveKernelReports")),
			Files: nonEmpty(join(p.SystemRoot, "MEMORY.DMP")),
			Match: hasExt(".dmp"),
		},
		{
			ID: "error-reports", Category: core.Logs, Default: true, MinAge: week,
			Name: txt("Error reports", "Hata raporları"),
			Description: txt(
				"Windows Error Reporting files for your account, older than 7 days.",
				"Hesabınıza ait Windows Hata Raporlama dosyaları; 7 günden eski olanlar.",
			),
			Roots: wer(p.LocalAppData),
		},
		{
			ID: "system-error-reports", Category: core.Logs, Admin: true, Default: true, MinAge: week,
			Name: txt("System error reports", "Sistem hata raporları"),
			Description: txt(
				"Windows Error Reporting files shared by all users, older than 7 days.",
				"Tüm kullanıcıların ortak Windows Hata Raporlama dosyaları; 7 günden eski olanlar.",
			),
			Roots: wer(p.ProgramData),
		},
		{
			ID: "shader-cache", Category: core.Graphics, MinAge: day,
			Name: txt("GPU shader caches", "GPU gölgelendirici önbellekleri"),
			Description: txt(
				"Shaders compiled by DirectX and by the NVIDIA, AMD and Intel drivers.",
				"DirectX'in ve NVIDIA, AMD, Intel sürücülerinin derlediği gölgelendiriciler.",
			),
			Caveat: txt(
				"Games may stutter briefly while their shaders are compiled again.",
				"Gölgelendiriciler yeniden derlenirken oyunlarda kısa süreli takılmalar olabilir.",
			),
			Roots: nonEmpty(
				join(p.LocalAppData, "D3DSCache"),
				join(p.LocalAppData, "NVIDIA", "DXCache"),
				join(p.LocalAppData, "NVIDIA", "GLCache"),
				join(p.LocalLow, "NVIDIA", "PerDriverVersion", "DXCache"),
				join(p.LocalAppData, "AMD", "DxCache"),
				join(p.LocalAppData, "AMD", "DxcCache"),
				join(p.LocalAppData, "AMD", "GLCache"),
				join(p.LocalAppData, "AMD", "VkCache"),
				join(p.LocalLow, "Intel", "ShaderCache"),
			),
		},
	}
}
