package metadata

type MetadataDownloader interface {
	DownloadRepository(owner string, name string, version string) error
	DownloadOrg(name string, version string) error
	SetCurrent(version string) error
}
