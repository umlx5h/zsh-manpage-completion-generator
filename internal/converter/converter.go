package converter

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var (
	shortOptReg = regexp.MustCompile(` -s '?([\w#?]+)'?`) // support -?, -#
	longOptReg  = regexp.MustCompile(` -l '?([\w-]+)'?`)  // support aaa-bbb
	oldOptReg   = regexp.MustCompile(` -o '?([\w+?]+)'?`) // support -??, -???
	descMsgReg  = regexp.MustCompile(` -d '(.*)'$`)
)

type Converter struct {
	r       io.Reader
	cmdName string
	opts    []Opt
}

type Opt struct {
	src           string // for debug
	shortOptNames []string
	longOptNames  []string
	oldOptNames   []string
	descMsg       string
}

func (o *Opt) getMergeOpts() []string {
	// merge short -> old -> long
	var mergeOpts []string
	appendMergeOpts := func(optNames []string, prefix string, escapeFn func(opt string) string) {
		var opts []string
		for _, n := range optNames {
			opts = append(opts, prefix+escapeFn(n))
		}
		mergeOpts = append(mergeOpts, opts...)
	}

	escapeNoop := func(opt string) string {
		return opt
	}

	appendMergeOpts(o.shortOptNames, "-", escapeShortOpt)
	appendMergeOpts(o.oldOptNames, "-", escapeOldOpt)
	appendMergeOpts(o.longOptNames, "--", escapeNoop)

	return mergeOpts
}

func NewConverter(r io.Reader, cmdName string) *Converter {
	return &Converter{r: r, cmdName: cmdName}
}

func escapeDescMsg(opt string) string {
	e := opt

	// escape: 「\'HUP\'」  -> 「'"'"'HUP'"'"'」
	e = strings.ReplaceAll(e, `\'`, `'"'"'`)

	// delete whitespace
	e = strings.Join(strings.Fields(e), " ")

	// escape: []
	e = strings.ReplaceAll(e, `[`, `\[`)
	e = strings.ReplaceAll(e, `]`, `\]`)

	return e
}

func escapeCommonOpt(opt string) string {
	e := opt

	// support -?
	e = strings.ReplaceAll(e, `?`, `\?`)

	return e
}

func escapeShortOpt(opt string) string {
	e := opt

	e = escapeCommonOpt(e)

	// support -#
	e = strings.ReplaceAll(e, `#`, `\#`)

	return e
}

func escapeOldOpt(opt string) string {
	e := opt

	e = escapeCommonOpt(e)

	return e
}

func SplitLines(s string) ([]string, error) {
	var lines []string
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if sc.Err() != nil {
		return nil, sc.Err()
	}
	return lines, nil
}

func (c *Converter) parse() error {
	b, err := io.ReadAll(c.r)
	if err != nil {
		return err
	}

	lines, err := SplitLines(string(b))
	if err != nil {
		return err
	}

	var optLines []string
	for _, l := range lines {
		if strings.HasPrefix(l, "complete -c") {
			optLines = append(optLines, l)
		}
	}
	if len(optLines) == 0 {
		return fmt.Errorf("complete commands don't exist")
	}

	var opts []Opt
	for _, line := range optLines {
		// get only opts portion, deleting -d portion
		optLine := descMsgReg.ReplaceAllString(line, "")

		var (
			shortOptNames []string
			longOptNames  []string
			oldOptNames   []string
			descMsg       string
		)

		optParsers := []struct {
			optNames      *[]string
			reg           *regexp.Regexp
			excludeFilter func(match string) (excluded bool)
		}{
			{&shortOptNames, shortOptReg, nil},
			{&longOptNames, longOptReg, func(match string) (excluded bool) {
				// Eliminate invalid option that contain consecutive hyphens
				return strings.Contains(match, "---")
			}},
			{&oldOptNames, oldOptReg, nil},
		}

		for _, p := range optParsers {
			for _, matches := range p.reg.FindAllStringSubmatch(optLine, -1) {
				for i, match := range matches {
					// only one group
					if i == 1 {
						if p.excludeFilter != nil && p.excludeFilter(match) {
							break
						}
						*p.optNames = append(*p.optNames, match)
					}
				}
			}
		}

		if len(shortOptNames)+len(longOptNames)+len(oldOptNames) == 0 {
			// pp.Println("no option skipped:", map[string]any{
			// 	"cmdName": c.cmdName,
			// 	"line":    line,
			// })
			continue
		}

		// get description text
		matches := descMsgReg.FindStringSubmatch(line)
		if len(matches) >= 1+1 {
			// -d must be only one option
			descMsg = matches[1]
		}

		o := Opt{
			shortOptNames: shortOptNames,
			longOptNames:  longOptNames,
			oldOptNames:   oldOptNames,
			src:           line,
		}

		if descMsg != "" {
			o.descMsg = descMsg
		}

		opts = append(opts, o)
	}

	if len(opts) == 0 {
		return fmt.Errorf("not found opts")
	}

	c.opts = opts
	return nil
}

const zshCompTemplate = `#compdef %s

_arguments \
         '*:file:_files' \
`

func getZshCompTemplate(commandName string) string {
	return fmt.Sprintf(zshCompTemplate, commandName)
}

func (c *Converter) Convert() (fileContent string, err error) {
	if err := c.parse(); err != nil {
		return "", fmt.Errorf("convert error: %w", err)
	}

	str := strings.Builder{}
	str.WriteString(getZshCompTemplate(c.cmdName))

	for i, opt := range c.opts {
		var args string

		allOpts := opt.getMergeOpts()

		// allOpts must be >= 1
		if len(allOpts) == 1 {
			args = fmt.Sprintf("'%s", allOpts[0])
		} else {
			args = fmt.Sprintf("{%s}'", strings.Join(allOpts, ","))
		}

		str.WriteString(fmt.Sprintf("\t\t%s", args))
		if opt.descMsg != "" {
			str.WriteString(fmt.Sprintf("[%s]", escapeDescMsg(opt.descMsg)))
		}
		str.WriteString("'")

		if i+1 < len(c.opts) {
			// line break except last option
			str.WriteString(" \\\n")
		}
	}
	str.WriteString("\n")

	return str.String(), nil
}
