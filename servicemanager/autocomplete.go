package servicemanager

import (
	"flag"
	"fmt"
	"sm2/cli"
	"sort"
	"strings"
)

// Print a valid bash autocomplete config to stdout
// intended use would be to pipe it into a file in the os's autocomplete folder
func GenerateAutoCompletionScript() {
	fmt.Println("# Below is a bash completion script for tab completion")
	fmt.Println(
		`_serv_words()
{
	local count cur
	cur=${COMP_WORDS[COMP_CWORD]}
	prev=${COMP_WORDS[COMP_CWORD-1]}
	count=${COMP_CWORD}
	words=$(sm2 --autocomplete --comp-cword $count --comp-pword \"$prev\" )
	COMPREPLY=( $(compgen -W "$words" -- $cur) )
	return 0
}
complete -F _serv_words sm2`)
}

func (sm *ServiceManager) GenerateAutocompleteResponse() string {
	count := sm.Commands.CompWordCount
	prev := strings.ReplaceAll(sm.Commands.CompPreviousWord, "\"", "")
	if dontComplete(prev) {
		return ""
	}
	var words strings.Builder
	opts := new(cli.UserOption)
	flagSet := cli.BuildFlagSet(opts)
	flagSet.VisitAll(func(f *flag.Flag) {
		if len(f.Name) == 1 {
			words.WriteString(fmt.Sprintf("-%s ", f.Name))
		} else {
			words.WriteString(fmt.Sprintf("--%s ", f.Name))
		}
	})

	if count >= 2 {
		keys := make([]string, len(sm.Services)+len(sm.Profiles))
		i := 0
		for k := range sm.Services {
			keys[i] = k
			i++

		}

		for k := range sm.Profiles {
			keys[i] = k
			i++

		}
		sort.Strings(keys)
		words.WriteString(strings.Join(keys, " "))
	}
	return words.String()
}

// Non boolean arguments can't be autocompleted
func dontComplete(previousWord string) bool {
	switch strings.ReplaceAll(previousWord, "--", "-") {
	case
		"-appendArgs",
		"-comp-cword",
		"-comp-pword",
		"-config",
		"-debug",
		"-logs",
		"-port",
		"-ports",
		"-search",
		"-wait",
		"-workers",
		"-delay-seconds":
		return true
	}
	return false
}
