package commands

type Ping struct{}

func (command *Ping) Execute(args []string) error {
	err := globalClient().Ping()
	failIf(err)

	return nil
}
