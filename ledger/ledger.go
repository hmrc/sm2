package ledger

type Ledger struct {
	SaveStateFile       func(string, StateFile) error
	LoadStateFile       func(string) (StateFile, error)
	ClearStateFile      func(string) error
	FindAllStateFiles   func(string) ([]StateFile, error)
	SaveProxyStateFile  func(string, ProxyStateFile) error
	LoadProxyStateFile  func(string) ProxyStateFile
	ClearProxyStateFile func(string) error
	SaveInstallFile     func(string, InstallFile) error
	LoadInstallFile     func(string) (InstallFile, error)
}

func NewLedger() Ledger {
	return Ledger{
		SaveStateFile:     saveStateFile,
		LoadStateFile:     loadStateFile,
		ClearStateFile:    clearStateFile,
		FindAllStateFiles: findAll,

		SaveProxyStateFile:  saveProxyStateFile,
		LoadProxyStateFile:  loadProxyStateFile,
		ClearProxyStateFile: clearProxyStateFile,

		SaveInstallFile: saveInstallFile,
		LoadInstallFile: loadInstallFile,
	}
}
