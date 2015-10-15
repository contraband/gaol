package commands

type Destroy struct{}

func (command *Destroy) Execute(handles []string) error {
	client := globalClient()

	for _, handle := range handles {
		err := client.Destroy(handle)
		failIf(err)
	}

	return nil
}
