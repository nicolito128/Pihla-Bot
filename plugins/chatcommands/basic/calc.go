package basic

import (
	"fmt"
	"strconv"

	"github.com/nicolito128/Pihla-Bot/client"
	"github.com/nicolito128/Pihla-Bot/commands"
	"github.com/nicolito128/Pihla-Bot/utils"
	"github.com/nicolito128/go-calculator"
)

var CalcCommand = &commands.Command[*client.Message]{
	Name: "calc",

	Description: "A calculator.",

	Usage: "calc [expression]",

	AllowPM: true,

	Handler: func(m *client.Message) error {
		exp := m.Content
		if exp == "" || utils.ToID(exp) == "" {
			return fmt.Errorf("invalid expression: you must provided a valid mathematical expression to calculate")
		}

		result, err := calculator.Resolve(exp)
		if err != nil {
			return fmt.Errorf("something goes wrong: %w", err)
		}

		value := strconv.FormatFloat(result, 'f', -1, 64)
		m.Send(fmt.Sprintf("Result: ``%s``", value))
		return nil
	},
}
