package ledger

type Ledger struct {
	SaveStateFile     func(string, StateFile) error
	LoadStateFile     func(string) (StateFile, error)
	ClearStateFile    func(string) error
	FindAllStateFiles func(string) ([]StateFile, error)
	SaveProxyState    func(string, ProxyState) error
	LoadProxyState    func(string) ProxyState
	ClearProxyState   func(string) error
	SaveInstallFile   func(string, InstallFile) error
	LoadInstallFile   func(string) (InstallFile, error)
}

func NewLedger() Ledger {
	return Ledger{
		SaveStateFile:     saveStateFile,
		LoadStateFile:     loadStateFile,
		ClearStateFile:    clearStateFile,
		FindAllStateFiles: findAll,

		SaveProxyState:  saveProxyState,
		LoadProxyState:  loadProxyState,
		ClearProxyState: clearProxyState,

		SaveInstallFile: saveInstallFile,
		LoadInstallFile: loadInstallFile,
	}
}
