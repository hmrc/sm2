package cli

import (
	"flag"
	"fmt"
)

// Print a valid bash autocomplete config to stdout
// intended use would be to pipe it into a file in the os's autocomplete folder
func GenerateAutoCompletions() {
	opts := new(UserOption)
	flagset := buildFlagSet(opts)
	fmt.Println("# copy this content into a file and place it in your OS's autocomplete folder")
	fmt.Printf("complete -W \"")
	flagset.VisitAll(func(f *flag.Flag) {
		fmt.Printf("-%s ", f.Name)
	})

	fmt.Print("\" sm2\n")
}
