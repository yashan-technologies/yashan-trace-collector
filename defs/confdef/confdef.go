package confdef

func InitConf(ytcConf string) error {
	if err := initYTCConf(ytcConf); err != nil {
		return err
	}
	if err := initStrategyConf(_ytcConf.StrategyPath); err != nil {
		return err
	}
	return nil
}
