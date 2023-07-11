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
	fmt.Println("# copy this content into a file and place it in your OS's autocomplete folder")
	fmt.Println(
		`_serv_words()
{
	local count cur
	cur=${COMP_WORDS[COMP_CWORD]}
	count=${COMP_CWORD}
	words=$(sm2 --autocomplete --comp-cword $count)
	COMPREPLY=( $(compgen -W "$words" -- $cur) )
	return 0
}
complete -F _serv_words sm2
`)
}

func (sm *ServiceManager) GenerateAutocompleteResponse() string {
	count := sm.Commands.CompWordCount
	var words strings.Builder
	opts := new(cli.UserOption)
	cli.BuildFlagSet(opts).VisitAll(func(f *flag.Flag) {
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
