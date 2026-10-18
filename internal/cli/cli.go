package cli

import "github.com/Gulliby-development/youngkit/internal/base"

// Run dispatches a command and writes its complete result atomically.
func Run(args []string, io base.IO) int {
	var output string
	var err error

	if len(args) == 0 {
		err = base.Fail("missing command (count, list, rank, unrank, conjugate, hooks, compare, neighbors, rsk, unrsk, syt or core)")
	} else {
		switch args[0] {
		case "count":
			output, err = RunCount(args[1:], io)
		case "list":
			output, err = RunList(args[1:], io)
		case "rank":
			output, err = RunRank(args[1:], io)
		case "unrank":
			output, err = RunUnrank(args[1:], io)
		case "conjugate":
			output, err = RunConjugate(args[1:], io)
		case "hooks":
			output, err = RunHooks(args[1:], io)
		case "compare":
			output, err = RunCompare(args[1:], io)
		case "neighbors":
			output, err = RunNeighbors(args[1:], io)
		case "rsk":
			output, err = RunRsk(args[1:], io)
		case "unrsk":
			output, err = RunUnrsk(args[1:], io)
		case "syt":
			output, err = RunSyt(args[1:], io)
		case "core":
			output, err = RunCore(args[1:], io)
		default:
			err = base.Fail("unknown command " + base.Show(args[0]))
		}
	}

	if err != nil {
		io.Err("youngkit: " + err.Error() + "\n")
		return 2
	}

	io.Out(output)
	return 0
}
