package mylibrary

/* Print message */
func PrintMessage(who string) string {
	return "Hello, " + who
}

type Salutation struct {
	name     string
	greeting string
}
