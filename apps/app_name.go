package apps

type AppName string

const (
	Sonarr            AppName = "sonarr"
	Radarr            AppName = "radarr"
	Lidarr            AppName = "lidarr"
	Readarr           AppName = "readarr"
	Prowlarr          AppName = "prowlarr"
	Whispar           AppName = "whispar"
	Bazarr            AppName = "bazarr"
	Plex              AppName = "plex"
	Jellyfin          AppName = "jellyfin"
	Emby              AppName = "emby"
	Overseer          AppName = "overseer"
	Tautulli          AppName = "tautulli"
	Nzbget            AppName = "nzbget"
	ruTorrent         AppName = "rutorrent"
	Sabnzbd           AppName = "sabnzbd"
	VueTorrent        AppName = "vuetorrent"
	qBittorrent       AppName = "qbittorrent"
	Deluge            AppName = "deluge"
	Jackett           AppName = "jackett"
	Librespeed        AppName = "librespeed"
	Synclounge        AppName = "synclounge"
	Lazylibrarian     AppName = "lazylibrarian"
	CalibreWeb        AppName = "calibreweb"
	Transmission      AppName = "transmission"
	Mylar             AppName = "mylar"
	Duplicati         AppName = "duplicati"
	Xbackbone         AppName = "xbackbone"
	Filebrowser       AppName = "filebrowser"
	Organizr          AppName = "organizr"
	Unraid            AppName = "unraid"
	Ombi              AppName = "ombi"
	Gitea             AppName = "gitea"
	Pihole            AppName = "pihole"
	Dozzle            AppName = "dozzle"
	Nzbhydra          AppName = "nzbhydra"
	Portaine          AppName = "portaine"
	Guacamol          AppName = "guacamol"
	Netdat            AppName = "netdat"
	Requestr          AppName = "requestr"
	Adguar            AppName = "adguar"
	Gap               AppName = "gap"
	Bitwarde          AppName = "bitwarde"
	Duplicac          AppName = "duplicac"
	Kitan             AppName = "kitan"
	Resilio           AppName = "resilio"
	Moviematc         AppName = "moviematc"
	Peti              AppName = "peti"
	Floo              AppName = "floo"
	UptimeKuma        AppName = "uptime-kuma"
	NginxProxyManager AppName = "nginx-proxy-manager"
	TheLounge         AppName = "thelounge"
	Grafana           AppName = "grafana"
	Monitorr          AppName = "monitorr"
	Logarr            AppName = "logarr"
	PLPP              AppName = "plpp"
	Webtools          AppName = "webtools"
)

func (appName AppName) IsSupported() bool {
	supportedNames := map[AppName]bool{
		Sonarr:            true,
		Radarr:            true,
		Lidarr:            true,
		Readarr:           true,
		Prowlarr:          true,
		Whispar:           true,
		Bazarr:            true,
		Plex:              true,
		Jellyfin:          true,
		Emby:              true,
		Overseer:          true,
		Tautulli:          true,
		Nzbget:            true,
		ruTorrent:         true,
		Sabnzbd:           true,
		VueTorrent:        true,
		qBittorrent:       true,
		Deluge:            true,
		Jackett:           true,
		Librespeed:        true,
		Synclounge:        true,
		Lazylibrarian:     true,
		CalibreWeb:        true,
		Transmission:      true,
		Mylar:             true,
		Duplicati:         true,
		Xbackbone:         true,
		Filebrowser:       true,
		Organizr:          true,
		Unraid:            true,
		Ombi:              true,
		Gitea:             true,
		Pihole:            true,
		Dozzle:            true,
		Nzbhydra:          true,
		Portaine:          true,
		Guacamol:          true,
		Netdat:            true,
		Requestr:          true,
		Adguar:            true,
		Gap:               true,
		Bitwarde:          true,
		Duplicac:          true,
		Kitan:             true,
		Resilio:           true,
		Moviematc:         true,
		Peti:              true,
		Floo:              true,
		UptimeKuma:        true,
		NginxProxyManager: true,
		TheLounge:         true,
		Grafana:           true,
		Monitorr:          true,
		Logarr:            true,
		PLPP:              true,
		Webtools:          true,
	}
	_, exists := supportedNames[appName]
	return exists
}
