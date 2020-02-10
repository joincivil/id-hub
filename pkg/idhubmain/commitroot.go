package idhubmain

// RunCommitRoot whatever the current root is for the root_merkletree is saved to the smart contract
func RunCommitRoot() error {
	config := populateConfig()
	rootCron := &RootCron{}
	return rootCron.StartCron(config)
}
