package base

import "strings"

// Spec describes a supported command-line option.
type Spec struct {
	Name             string
	Lo, Hi, Fallback int64
	Flag             bool
}

// Parsed contains validated option values and positional arguments.
type Parsed struct {
	Values map[string]int64
	Args   []string
}

// Parse scans argv before validating numeric option values. This ensures that
// structural scan errors take precedence over invalid option values.
func Parse(argv []string, specs []Spec) (Parsed, error) {
	specByName := make(map[string]Spec, len(specs))
	for _, spec := range specs {
		specByName[spec.Name] = spec
	}

	args := make([]string, 0, len(argv))
	rawValues := make(map[string]string)
	present := make(map[string]bool)

	for i := 0; i < len(argv); i++ {
		arg := argv[i]
		if !strings.HasPrefix(arg, "--") {
			args = append(args, arg)
			continue
		}

		name := arg
		value := ""
		hasValue := false
		if equals := strings.IndexByte(arg, '='); equals >= 0 {
			name = arg[:equals]
			value = arg[equals+1:]
			hasValue = true
		}

		spec, ok := specByName[name]
		if !ok {
			return Parsed{}, Fail("unknown option " + Show(arg))
		}
		if spec.Flag {
			if hasValue {
				return Parsed{}, Fail("option " + name + " takes no value")
			}
			present[name] = true
			continue
		}

		if !hasValue {
			if i+1 >= len(argv) {
				return Parsed{}, Fail(name + " needs a value")
			}
			i++
			value = argv[i]
		}
		rawValues[name] = value
		present[name] = true
	}

	values := make(map[string]int64, len(specs))
	for _, spec := range specs {
		if spec.Flag {
			if present[spec.Name] {
				values[spec.Name] = 1
			} else {
				values[spec.Name] = spec.Fallback
			}
			continue
		}

		if !present[spec.Name] {
			values[spec.Name] = spec.Fallback
			continue
		}

		value, err := ParseInt(spec.Name, rawValues[spec.Name], spec.Lo, spec.Hi)
		if err != nil {
			return Parsed{}, err
		}
		values[spec.Name] = value
	}

	return Parsed{Values: values, Args: args}, nil
}
