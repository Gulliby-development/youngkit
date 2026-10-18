package cli_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/Gulliby-development/youngkit/internal/base"
	"github.com/Gulliby-development/youngkit/internal/cli"
)

type docsIO struct {
	reads int
	out   []string
	err   []string
}

func (io *docsIO) ReadStdin() []byte {
	io.reads++
	return nil
}

func (io *docsIO) Out(text string) {
	io.out = append(io.out, text)
}

func (io *docsIO) Err(text string) {
	io.err = append(io.err, text)
}

var _ base.IO = (*docsIO)(nil)

type docsFixture struct {
	command string
	args    []string
	output  string
}

type docsFence struct {
	language string
	content  string
}

func readDocsREADME(t *testing.T) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join("..", "..", "README.md"))
	if err != nil {
		t.Fatalf("read ../../README.md: %v", err)
	}
	return string(content)
}

func docsSection(markdown, title string) (string, error) {
	lines := strings.Split(markdown, "\n")
	header := "## " + title
	start := -1
	for i, line := range lines {
		if line != header {
			continue
		}
		if start != -1 {
			return "", fmt.Errorf("section %q occurs more than once", header)
		}
		start = i
	}
	if start == -1 {
		return "", fmt.Errorf("section %q is missing", header)
	}

	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			end = i
			break
		}
	}
	return strings.Join(lines[start+1:end], "\n"), nil
}

func docsFences(markdown string) ([]docsFence, error) {
	lines := strings.Split(markdown, "\n")
	var fences []docsFence
	for i := 0; i < len(lines); i++ {
		if !strings.HasPrefix(lines[i], "```") {
			continue
		}
		if lines[i] == "```" {
			return nil, fmt.Errorf("unexpected closing fence at line %d", i+1)
		}
		language := strings.TrimSpace(strings.TrimPrefix(lines[i], "```"))
		start := i + 1
		for i++; i < len(lines); i++ {
			if lines[i] != "```" {
				continue
			}
			fences = append(fences, docsFence{
				language: language,
				content:  strings.Join(lines[start:i], "\n") + "\n",
			})
			break
		}
		if i == len(lines) {
			return nil, fmt.Errorf("unterminated %q fence", language)
		}
	}
	return fences, nil
}

func TestDocumentationREADMEHasRequiredSectionsAndGuidance(t *testing.T) {
	readme := readDocsREADME(t)
	requiredSections := []string{
		"Install",
		"Usage",
		"Shapes and tableaux",
		"Commands",
		"Exit codes",
		"Layout",
		"Development",
	}
	for _, title := range requiredSections {
		if _, err := docsSection(readme, title); err != nil {
			t.Fatal(err)
		}
	}

	requiredProductText := []string{
		"offline, exact-integer command-line toolkit",
		"students and researchers in combinatorics",
		"Go's standard library",
		"go build -o youngkit .",
	}
	requiredShapeText := []string{
		"comma-separated, positive, nonincreasing parts",
		"`5,3,3,1`",
		"`0` denotes the empty shape",
		"leading-zero part",
		"trailing comma",
	}
	requiredTableauText := []string{
		"slash-separated rows and comma-separated entries",
		"`1,3,5/2,4/6`",
		"`0` denotes the empty tableau",
		"each integer from 1 through",
		"exactly once",
		"increasing from left to right",
		"top to bottom",
		"Empty rows",
		"exact integers",
	}
	requiredRuntimeText := []string{
		"status `0`",
		"status `2`",
		"youngkit: ",
		"standard error",
		"standard output",
		"gofmt -l .",
		"go vet ./...",
		"go test ./...",
	}
	for _, group := range [][]string{requiredProductText, requiredShapeText, requiredTableauText, requiredRuntimeText} {
		for _, text := range group {
			if !strings.Contains(readme, text) {
				t.Errorf("README is missing required text %q", text)
			}
		}
	}

	commands, err := docsSection(readme, "Commands")
	if err != nil {
		t.Fatal(err)
	}
	commandNames := []string{
		"count", "list", "rank", "unrank", "conjugate", "hooks",
		"compare", "neighbors", "rsk", "unrsk", "syt", "core",
	}
	found := make(map[string]int, len(commandNames))
	for _, paragraph := range strings.Split(commands, "\n\n") {
		paragraph = strings.TrimSpace(paragraph)
		for _, name := range commandNames {
			if strings.HasPrefix(paragraph, "`"+name+"` ") {
				found[name]++
				if len(strings.Fields(paragraph)) < 5 {
					t.Errorf("README paragraph for %s is not descriptive: %q", name, paragraph)
				}
			}
		}
	}
	for _, name := range commandNames {
		if found[name] != 1 {
			t.Errorf("README has %d command paragraphs for %s, want exactly one", found[name], name)
		}
	}
}

func TestDocumentationUsageLinesMatchNoArgumentDiagnostics(t *testing.T) {
	readme := readDocsREADME(t)
	usageSection, err := docsSection(readme, "Usage")
	if err != nil {
		t.Fatal(err)
	}
	usageFences, err := docsFences(usageSection)
	if err != nil {
		t.Fatal(err)
	}
	if len(usageFences) != 1 || usageFences[0].language != "text" {
		t.Fatalf("Usage has fences %#v, want one text usage block", usageFences)
	}

	fixtures := []struct {
		line string
		args []string
	}{
		{"youngkit count N [--max-part K] [--distinct]", []string{"count"}},
		{"youngkit list N [--max-part K] [--limit L]", []string{"list"}},
		{"youngkit rank SHAPE", []string{"rank"}},
		{"youngkit unrank N RANK", []string{"unrank"}},
		{"youngkit conjugate SHAPE", []string{"conjugate"}},
		{"youngkit hooks SHAPE [--entries K]", []string{"hooks"}},
		{"youngkit compare A B", []string{"compare"}},
		{"youngkit neighbors SHAPE", []string{"neighbors"}},
		{"youngkit rsk ENTRY... | youngkit rsk -", []string{"rsk"}},
		{"youngkit unrsk P Q", []string{"unrsk"}},
		{"youngkit syt SHAPE [--index K] [--limit L]", []string{"syt"}},
		{"youngkit core K SHAPE", []string{"core"}},
	}
	var wantBlock strings.Builder
	for _, fixture := range fixtures {
		wantBlock.WriteString(fixture.line)
		wantBlock.WriteByte('\n')
	}
	if got := usageFences[0].content; got != wantBlock.String() {
		t.Fatalf("README usage block = %q, want %q", got, wantBlock.String())
	}

	for _, fixture := range fixtures {
		if got := strings.Count(readme, fixture.line); got != 1 {
			t.Errorf("usage line %q occurs %d times in README, want once", fixture.line, got)
		}
		io := &docsIO{}
		if status := cli.Run(fixture.args, io); status != 2 {
			t.Errorf("cli.Run(%q) status = %d, want 2", fixture.args, status)
		}
		wantErr := "youngkit: usage: " + fixture.line + "\n"
		if len(io.out) != 0 {
			t.Errorf("cli.Run(%q) stdout calls = %#v, want none", fixture.args, io.out)
		}
		if len(io.err) != 1 || io.err[0] != wantErr {
			t.Errorf("cli.Run(%q) stderr calls = %#v, want one exact call %q", fixture.args, io.err, wantErr)
		}
		if io.reads != 0 {
			t.Errorf("cli.Run(%q) read stdin %d times, want none", fixture.args, io.reads)
		}
	}
}

func TestDocumentationExamplesMatchLiteralCommandsAndCLI(t *testing.T) {
	readme := readDocsREADME(t)
	fences, err := docsFences(readme)
	if err != nil {
		t.Fatal(err)
	}
	if len(fences) != 6 {
		t.Fatalf("README has %d fenced blocks, want the two commands, usage block and three outputs", len(fences))
	}
	languages := map[string]int{}
	for _, fence := range fences {
		languages[fence.language]++
	}
	if languages["sh"] != 2 || languages["text"] != 4 || len(languages) != 2 {
		t.Fatalf("README fence languages = %#v, want two sh and four text blocks", languages)
	}

	fixtures := []docsFixture{
		{
			command: "youngkit hooks 4,3,1 --entries 3",
			args:    []string{"hooks", "4,3,1", "--entries", "3"},
			output:  "shape       4,3,1\nsize        8\nhook        6 4 3 1\nhook        4 2 1\nhook        1\nsyt         70\nentries     3\nssyt        15\n",
		},
		{
			command: "youngkit rsk 2 5 1 4 3",
			args:    []string{"rsk", "2", "5", "1", "4", "3"},
			output:  "perm        2 5 1 4 3\nshape       2,2,1\nlis         2\nlds         3\np           1 3\np           2 4\np           5\nq           1 2\nq           3 4\nq           5\n",
		},
		{
			command: "youngkit core 3 5,4,1",
			args:    []string{"core", "3", "5,4,1"},
			output:  "shape       5,4,1\nk           3\ncore        2,1,1\nweight      2\nquotient    0=0\nquotient    1=1\nquotient    2=1\n",
		},
	}
	expected := make(map[string]docsFixture, len(fixtures))
	for _, fixture := range fixtures {
		expected[fixture.command] = fixture
	}

	lines := strings.Split(readme, "\n")
	extracted := make(map[string]string, len(fixtures))
	textOutputFences := 0
	for _, fence := range fences {
		if fence.language == "text" {
			textOutputFences++
		}
	}
	markers := 0
	for i, line := range lines {
		if !strings.HasPrefix(line, "`youngkit ") || !strings.HasSuffix(line, "` exits with 0 and prints:") {
			continue
		}
		markers++
		command := strings.TrimSuffix(strings.TrimPrefix(line, "`"), "` exits with 0 and prints:")
		fixture, ok := expected[command]
		if !ok {
			t.Fatalf("unexpected example command %q", command)
		}
		if _, duplicate := extracted[command]; duplicate {
			t.Fatalf("example command %q occurs more than once", command)
		}
		if i+2 >= len(lines) || lines[i+1] != "" || lines[i+2] != "```text" {
			t.Fatalf("example command %q is not followed by a text output block", command)
		}
		var output strings.Builder
		closed := false
		for j := i + 3; j < len(lines); j++ {
			if lines[j] == "```" {
				closed = true
				break
			}
			output.WriteString(lines[j])
			output.WriteByte('\n')
		}
		if !closed {
			t.Fatalf("example output for %q is unterminated", command)
		}
		extracted[command] = output.String()
		if output.String() != fixture.output {
			t.Errorf("README output for %q = %q, want literal output %q", command, output.String(), fixture.output)
		}
	}
	if markers != len(fixtures) || textOutputFences != 4 || len(extracted) != len(fixtures) {
		t.Fatalf("README has %d example markers, %d text blocks and %d extracted outputs; want 3, 4 and 3", markers, textOutputFences, len(extracted))
	}

	for _, fixture := range fixtures {
		io := &docsIO{}
		if status := cli.Run(fixture.args, io); status != 0 {
			t.Errorf("cli.Run(%q) status = %d, want 0 (stderr calls: %#v)", fixture.args, status, io.err)
		}
		if len(io.out) != 1 || io.out[0] != extracted[fixture.command] {
			t.Errorf("cli.Run(%q) stdout calls = %#v, want one exact call %q", fixture.args, io.out, extracted[fixture.command])
		}
		if len(io.err) != 0 {
			t.Errorf("cli.Run(%q) stderr calls = %#v, want none", fixture.args, io.err)
		}
		if io.reads != 0 {
			t.Errorf("cli.Run(%q) read stdin %d times, want none", fixture.args, io.reads)
		}
	}
}

func TestDocumentationLayoutMatchesRepository(t *testing.T) {
	readme := readDocsREADME(t)
	layout, err := docsSection(readme, "Layout")
	if err != nil {
		t.Fatal(err)
	}

	wantPaths := []string{
		"main.go",
		"internal/base",
		"internal/shape",
		"internal/count",
		"internal/tableau",
		"internal/cli",
	}
	requiredDescriptionText := []string{
		"entry point",
		"input",
		"Shape",
		"Partition",
		"Tableau",
		"commands",
	}
	lines := strings.Split(layout, "\n")
	var gotPaths []string
	descriptionIndex := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "- `") {
			t.Fatalf("unexpected content in Layout section: %q", line)
		}
		remainder := strings.TrimPrefix(line, "- `")
		closing := strings.IndexByte(remainder, '`')
		if closing <= 0 || !strings.HasPrefix(remainder[closing+1:], " - ") {
			t.Fatalf("Layout bullet has invalid format: %q", line)
		}
		path := remainder[:closing]
		description := strings.TrimPrefix(remainder[closing+1:], " - ")
		if description == "" || !strings.HasSuffix(description, ".") ||
			strings.Count(description, ".") != 1 || strings.ContainsAny(description, "!?") ||
			strings.Contains(description, "/") || strings.Contains(description, "`") {
			t.Errorf("Layout description for %q is not one sentence: %q", path, description)
		}
		if descriptionIndex < len(requiredDescriptionText) &&
			!strings.Contains(description, requiredDescriptionText[descriptionIndex]) {
			t.Errorf("Layout description for %q does not explain its contents: %q", path, description)
		}
		descriptionIndex++
		gotPaths = append(gotPaths, path)
	}
	if descriptionIndex != len(wantPaths) {
		t.Fatalf("Layout has %d descriptions, want %d", descriptionIndex, len(wantPaths))
	}
	if len(gotPaths) != len(wantPaths) {
		t.Fatalf("Layout paths = %#v, want exactly %#v", gotPaths, wantPaths)
	}
	for i, path := range wantPaths {
		if gotPaths[i] != path {
			t.Errorf("Layout path %d = %q, want %q", i+1, gotPaths[i], path)
		}
	}

	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	rootReal, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("resolve repository root symlinks: %v", err)
	}
	for _, path := range gotPaths {
		fullPath := filepath.Join(root, filepath.FromSlash(path))
		info, err := os.Stat(fullPath)
		if err != nil {
			t.Errorf("listed path %q does not exist: %v", path, err)
			continue
		}
		if path == "main.go" && info.IsDir() {
			t.Errorf("listed path %q is a directory, want a file", path)
		}
		if path != "main.go" && !info.IsDir() {
			t.Errorf("listed path %q is not a directory", path)
		}
		resolved, err := filepath.EvalSymlinks(fullPath)
		if err != nil {
			t.Errorf("resolve listed path %q: %v", path, err)
			continue
		}
		relative, err := filepath.Rel(rootReal, resolved)
		if err != nil {
			t.Errorf("make listed path %q repository-relative: %v", path, err)
			continue
		}
		if got := filepath.ToSlash(relative); got != path {
			t.Errorf("resolved listed path %q is %q relative to repository root", path, got)
		}
	}

	internalRoot := filepath.Join(root, "internal")
	goFolders := make(map[string]bool)
	err = filepath.WalkDir(internalRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" {
			return nil
		}
		folder, err := filepath.EvalSymlinks(filepath.Dir(path))
		if err != nil {
			return fmt.Errorf("resolve Go-bearing folder %q: %w", path, err)
		}
		relative, err := filepath.Rel(rootReal, folder)
		if err != nil {
			return fmt.Errorf("make Go-bearing folder %q repository-relative: %w", path, err)
		}
		goFolders[filepath.ToSlash(relative)] = true
		return nil
	})
	if err != nil {
		t.Fatalf("walk internal Go folders: %v", err)
	}

	listedFolders := make(map[string]bool)
	for _, path := range gotPaths {
		if strings.HasPrefix(path, "internal/") {
			listedFolders[path] = true
		}
	}
	var discovered []string
	for path := range goFolders {
		discovered = append(discovered, path)
	}
	sort.Strings(discovered)
	var listed []string
	for path := range listedFolders {
		listed = append(listed, path)
	}
	sort.Strings(listed)
	if len(goFolders) != len(listedFolders) {
		t.Fatalf("Go-bearing folders = %#v, listed internal folders = %#v", discovered, listed)
	}
	for i := range discovered {
		if discovered[i] != listed[i] {
			t.Errorf("Go-bearing folders = %#v, listed internal folders = %#v", discovered, listed)
			break
		}
	}
}

func requireDocsSectionText(t *testing.T, title, text string) {
	t.Helper()

	section, err := docsSection(readDocsREADME(t), title)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(section, text) {
		t.Errorf("README section %q is missing %q", title, text)
	}
}

func docsCommandParagraph(t *testing.T, command string) string {
	t.Helper()

	commands, err := docsSection(readDocsREADME(t), "Commands")
	if err != nil {
		t.Fatal(err)
	}
	prefix := "`" + command + "` "
	for _, paragraph := range strings.Split(commands, "\n\n") {
		paragraph = strings.TrimSpace(paragraph)
		if strings.HasPrefix(paragraph, prefix) {
			return paragraph
		}
	}
	t.Fatalf("README has no command paragraph for %q", command)
	return ""
}

func TestDocumentationAudienceAndProductPurpose(t *testing.T) {
	readme := readDocsREADME(t)
	for _, phrase := range []string{"offline, exact-integer command-line toolkit", "students and researchers in combinatorics"} {
		if !strings.Contains(readme, phrase) {
			t.Errorf("README introduction is missing %q", phrase)
		}
	}
}

func TestDocumentationInstallBuildAndOutputLocation(t *testing.T) {
	requireDocsSectionText(t, "Install", "go build -o youngkit .")
	requireDocsSectionText(t, "Install", "executable in the current directory")
}

func TestDocumentationCountCommandParagraph(t *testing.T) {
	paragraph := docsCommandParagraph(t, "count")
	for _, phrase := range []string{"counts partitions", "largest part", "distinct parts"} {
		if !strings.Contains(paragraph, phrase) {
			t.Errorf("count paragraph is missing %q", phrase)
		}
	}
}

func TestDocumentationListCommandParagraph(t *testing.T) {
	paragraph := docsCommandParagraph(t, "list")
	for _, phrase := range []string{"reverse lexicographic order", "largest-part bound", "result limit"} {
		if !strings.Contains(paragraph, phrase) {
			t.Errorf("list paragraph is missing %q", phrase)
		}
	}
}

func TestDocumentationRankCommandParagraph(t *testing.T) {
	if !strings.Contains(docsCommandParagraph(t, "rank"), "zero-based position") {
		t.Error("rank paragraph does not explain the zero-based position")
	}
}

func TestDocumentationUnrankCommandParagraph(t *testing.T) {
	if !strings.Contains(docsCommandParagraph(t, "unrank"), "zero-based partition position") {
		t.Error("unrank paragraph does not explain the partition position")
	}
}

func TestDocumentationConjugateCommandParagraph(t *testing.T) {
	if !strings.Contains(docsCommandParagraph(t, "conjugate"), "reflects a shape across its diagonal") {
		t.Error("conjugate paragraph does not explain reflection")
	}
}

func TestDocumentationHooksCommandParagraph(t *testing.T) {
	paragraph := docsCommandParagraph(t, "hooks")
	for _, phrase := range []string{"hook length", "standard-tableau count", "semistandard-tableau count"} {
		if !strings.Contains(paragraph, phrase) {
			t.Errorf("hooks paragraph is missing %q", phrase)
		}
	}
}

func TestDocumentationCompareCommandParagraph(t *testing.T) {
	paragraph := docsCommandParagraph(t, "compare")
	for _, phrase := range []string{"dominance", "partial sums"} {
		if !strings.Contains(paragraph, phrase) {
			t.Errorf("compare paragraph is missing %q", phrase)
		}
	}
}

func TestDocumentationNeighborsCommandParagraph(t *testing.T) {
	if !strings.Contains(docsCommandParagraph(t, "neighbors"), "adding or removing one box") {
		t.Error("neighbors paragraph does not explain one-box changes")
	}
}

func TestDocumentationRSKCommandParagraph(t *testing.T) {
	paragraph := docsCommandParagraph(t, "rsk")
	for _, phrase := range []string{"Robinson–Schensted", "longest increasing", "decreasing subsequence"} {
		if !strings.Contains(paragraph, phrase) {
			t.Errorf("rsk paragraph is missing %q", phrase)
		}
	}
}

func TestDocumentationUnrskCommandParagraph(t *testing.T) {
	if !strings.Contains(docsCommandParagraph(t, "unrsk"), "inverse Robinson–Schensted") {
		t.Error("unrsk paragraph does not explain the inverse bijection")
	}
}

func TestDocumentationSYTCommandParagraph(t *testing.T) {
	paragraph := docsCommandParagraph(t, "syt")
	for _, phrase := range []string{"`--index`", "jumps to a position", "`--limit`"} {
		if !strings.Contains(paragraph, phrase) {
			t.Errorf("syt paragraph is missing %q", phrase)
		}
	}
}

func TestDocumentationCoreCommandParagraph(t *testing.T) {
	paragraph := docsCommandParagraph(t, "core")
	for _, phrase := range []string{"`k`-core", "weight", "runner quotient shapes"} {
		if !strings.Contains(paragraph, phrase) {
			t.Errorf("core paragraph is missing %q", phrase)
		}
	}
}

func TestDocumentationShapeNotationUsesPositiveNonincreasingParts(t *testing.T) {
	requireDocsSectionText(t, "Shapes and tableaux", "comma-separated, positive, nonincreasing parts")
	requireDocsSectionText(t, "Shapes and tableaux", "`5,3,3,1`")
}

func TestDocumentationShapeNotationUsesZeroForEmptyShape(t *testing.T) {
	requireDocsSectionText(t, "Shapes and tableaux", "`0` denotes the empty shape")
}

func TestDocumentationShapeNotationNamesMalformedForms(t *testing.T) {
	requireDocsSectionText(t, "Shapes and tableaux", "leading-zero part (`03`)")
	requireDocsSectionText(t, "Shapes and tableaux", "trailing comma")
	requireDocsSectionText(t, "Shapes and tableaux", "empty shape string is rejected")
}

func TestDocumentationTableauNotationShowsRowsAndEntries(t *testing.T) {
	requireDocsSectionText(t, "Shapes and tableaux", "slash-separated rows and comma-separated entries")
	requireDocsSectionText(t, "Shapes and tableaux", "`1,3,5/2,4/6`")
}

func TestDocumentationTableauNotationUsesZeroForEmptyTableau(t *testing.T) {
	requireDocsSectionText(t, "Shapes and tableaux", "`0` denotes the empty tableau")
}

func TestDocumentationTableauNotationRejectsEmptyRows(t *testing.T) {
	requireDocsSectionText(t, "Shapes and tableaux", "Empty rows, malformed entries and tableaux that fail these conditions are rejected")
}

func TestDocumentationStandardTableauEntriesAreUniqueAndComplete(t *testing.T) {
	requireDocsSectionText(t, "Shapes and tableaux", "each integer from 1 through the tableau's size exactly once")
}

func TestDocumentationStandardTableauRowsIncrease(t *testing.T) {
	requireDocsSectionText(t, "Shapes and tableaux", "increasing from left to right in every row")
}

func TestDocumentationStandardTableauColumnsIncrease(t *testing.T) {
	requireDocsSectionText(t, "Shapes and tableaux", "from top to bottom in every column")
}

func TestDocumentationCountsAndRanksAreExactIntegers(t *testing.T) {
	requireDocsSectionText(t, "Shapes and tableaux", "Counts and ranks are exact integers")
}

func TestDocumentationExitCodesAndStreamBehavior(t *testing.T) {
	exitCodes, err := docsSection(readDocsREADME(t), "Exit codes")
	if err != nil {
		t.Fatal(err)
	}
	for _, phrase := range []string{
		"successful command exits with status `0`",
		"error exits with status `2`",
		"youngkit: `-prefixed diagnostic",
		"standard error",
		"nothing to standard output",
		"Successful output lines end with LF",
	} {
		if !strings.Contains(exitCodes, phrase) {
			t.Errorf("Exit codes section is missing %q", phrase)
		}
	}
}

func TestDocumentationDevelopmentListsAllGoChecks(t *testing.T) {
	development, err := docsSection(readDocsREADME(t), "Development")
	if err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{"gofmt -l .", "go vet ./...", "go test ./..."} {
		if !strings.Contains(development, command) {
			t.Errorf("Development section is missing %q", command)
		}
	}
}
