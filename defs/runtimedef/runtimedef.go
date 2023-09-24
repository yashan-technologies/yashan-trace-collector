package runtimedef

func InitRuntime() error {
	if err := initYTCHome(); err != nil {
		return err
	}
	if err := initExecuteable(); err != nil {
		return err
	}
	if err := initOSRelease(); err != nil {
		return err
	}
	if err := initExecuter(); err != nil {
		return err
	}
	initRootUsername()
	return nil
}
