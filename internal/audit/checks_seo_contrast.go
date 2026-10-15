package audit

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

const (
	minContrastRatio  = 4.5
	minLargeTextRatio = 3.0
	asciiSpace        = " \t\n\f\r"
)

var (
	hexColorRegex = regexp.MustCompile(`^#([0-9a-f]{3}|[0-9a-f]{6})$`)
	rgbColorRegex = regexp.MustCompile(`^rgb\(\s*(\d{1,3})\s*,\s*(\d{1,3})\s*,\s*(\d{1,3})\s*\)$`)
	namedColors   = map[string][3]float64{"white": {255, 255, 255}, "black": {0, 0, 0}}
	largeTextTags = map[string]bool{"h1": true, "h2": true, "h3": true}
	contrastTags  = []string{"p", "a", "li", "button", "label", "h1", "h2", "h3", "h4", "h5", "h6"}
)

func seoColorContrast(l *checkList, s *PageSignals) {
	checked, low := 0, 0
	for _, name := range contrastTags {
		for _, t := range s.Tags[name] {
			if strings.Trim(t.Text, asciiSpace) == "" {
				continue
			}
			fg, bg, ok := inlineColorPair(t.Attrs["style"])
			if !ok {
				continue
			}
			checked++
			min := minContrastRatio
			if largeTextTags[name] {
				min = minLargeTextRatio
			}
			if contrastRatio(fg, bg) < min {
				low++
			}
		}
	}
	skipUnless(l, checked > 0, "color_contrast", fmt.Sprintf("%d of %d inline-colored text element(s) fall below WCAG AA contrast (4.5:1, 3:1 for h1-h3)", low, checked), "No inline color pairs to check", low == 0)
}

func inlineColorPair(style string) (fg, bg [3]float64, ok bool) {
	var haveFg, haveBg bool
	for _, decl := range strings.Split(style, ";") {
		prop, value, found := strings.Cut(decl, ":")
		if !found {
			continue
		}
		color, resolved := parseCSSColor(strings.ToLower(strings.Trim(value, asciiSpace)))
		switch strings.ToLower(strings.Trim(prop, asciiSpace)) {
		case "color":
			fg, haveFg = color, resolved
		case "background-color":
			bg, haveBg = color, resolved
		}
	}
	return fg, bg, haveFg && haveBg
}

func parseCSSColor(v string) ([3]float64, bool) {
	if c, ok := namedColors[v]; ok {
		return c, true
	}
	if m := hexColorRegex.FindStringSubmatch(v); m != nil {
		h := m[1]
		if len(h) == 3 {
			h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
		}
		var c [3]float64
		for i := range c {
			n, _ := strconv.ParseUint(h[i*2:i*2+2], 16, 8)
			c[i] = float64(n)
		}
		return c, true
	}
	if m := rgbColorRegex.FindStringSubmatch(v); m != nil {
		var c [3]float64
		for i := range c {
			n, _ := strconv.Atoi(m[i+1])
			if n > 255 {
				return c, false
			}
			c[i] = float64(n)
		}
		return c, true
	}
	return [3]float64{}, false
}

func relativeLuminance(c [3]float64) float64 {
	var lin [3]float64
	for i, v := range c {
		v /= 255
		if v <= 0.03928 {
			lin[i] = v / 12.92
		} else {
			lin[i] = math.Pow((v+0.055)/1.055, 2.4)
		}
	}
	return 0.2126*lin[0] + 0.7152*lin[1] + 0.0722*lin[2]
}

func contrastRatio(a, b [3]float64) float64 {
	la, lb := relativeLuminance(a), relativeLuminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}
