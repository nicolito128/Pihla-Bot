package commands

type HandleFunc[T any] func(T) error

type Command[T any] struct {
	Name string

	Description string

	Usage string

	AllowPM bool

	Permissions Permission

	Handler HandleFunc[T]

	SubCommands map[string]*Command[T]
}

func New[T any](name, desc string, handler HandleFunc[T]) *Command[T] {
	return &Command[T]{
		Name:        name,
		Description: desc,
		Handler:     handler,
		SubCommands: make(map[string]*Command[T]),
	}
}

func FindDeeperSubCommand[T any](cmd *Command[T], body []string) (*Command[T], []string) {
	deepSubCommand := cmd
	rest := body
	for i, part := range body {
		subcmd, ok := cmd.SubCommands[part]
		if ok {
			if subcmd.Name == "" {
				subcmd.Name = part
			}

			deepSubCommand, rest = FindDeeperSubCommand(subcmd, body[i+1:])
		}
	}

	return deepSubCommand, rest
}
