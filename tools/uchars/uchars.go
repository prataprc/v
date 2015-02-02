package main

import "unicode"
import "flag"
import "strings"
import "fmt"
import "log"
import "sort"

var options struct {
	lists      []string
	maps       []string
	categories []string
	scripts    []string
	properties []string
}

var mapbyFns = map[string]func(r rune) string{
	"category": runesCategory,
	"script":   runesScript,
	"property": runesProperties,
}

func argParse() {
	var list, mapp, catg, scripts, properties string

	flag.StringVar(&list, "list", "",
		"list comma separated one or more tables names"+
			`[controls,digits,ranges,categories,scripts,properties`)
	flag.StringVar(&mapp, "map", "",
		"map listed runes by categories,scripts,properties")
	flag.StringVar(&catg, "categories", "",
		"list comma separated one or more category names")
	flag.StringVar(&scripts, "scripts", "",
		"list comma separated one or more script names")
	flag.StringVar(&properties, "properties", "",
		"list comma separated one or more property names")

	flag.Parse()
	// list
	if list != "" {
		options.lists = strings.Split(list, ",")
	}
	// map
	if mapp != "" {
		options.maps = strings.Split(mapp, ",")
	}
	// categories
	if catg == "all" {
		options.categories = make([]string, 0)
		for name := range unicode.Categories {
			options.categories = append(options.categories, name)
		}
		sort.Strings(options.categories)
	} else if catg != "" {
		options.categories = strings.Split(catg, ",")
	}
	// scripts
	if scripts == "all" {
		options.scripts = make([]string, 0)
		for name := range unicode.Scripts {
			options.scripts = append(options.scripts, name)
		}
		sort.Strings(options.scripts)
	} else if scripts != "" {
		options.scripts = strings.Split(scripts, ",")
	}
	// properties
	if properties == "all" {
		options.properties = make([]string, 0)
		for name := range unicode.Properties {
			options.properties = append(options.properties, name)
		}
		sort.Strings(options.properties)
	} else if properties != "" {
		options.properties = strings.Split(properties, ",")
	}
}

func main() {
	argParse()
	if options.lists != nil {
		for _, name := range options.lists {
			list(name)
			fmt.Println("\n")
		}
	}
}

func list(name string) {
	switch name {
	case "controls":
		listCharAs(name, unicode.IsControl)
	case "digits":
		listCharAs(name, unicode.IsDigit)
	case "graphics":
		listCharAs(name, unicode.IsGraphic)
	case "letter":
		listCharAs(name, unicode.IsLetter)
	case "lower":
		listCharAs(name, unicode.IsLower)
	case "mark":
		listCharAs(name, unicode.IsMark)
	case "number":
		listCharAs(name, unicode.IsNumber)
	case "print":
		listCharAs(name, unicode.IsPrint)
	case "punct":
		listCharAs(name, unicode.IsPunct)
	case "space":
		listCharAs(name, unicode.IsSpace)
	case "symbol":
		listCharAs(name, unicode.IsSymbol)
	case "title":
		listCharAs(name, unicode.IsTitle)
	case "upper":
		listCharAs(name, unicode.IsUpper)

	case "ranges":
		listRanges()
	case "categories":
		listCategories()
	case "scripts":
		listScripts()
	case "properties":
		listProperties()
	default:
		log.Fatalf("unknown list %s\n", name)
	}
}

func listRanges() {
	// see whether to range on categories
	if len(options.categories) > 0 {
		fmt.Println("categories:", options.categories)
		for _, catg := range options.categories {
			rt, ok := unicode.Categories[catg]
			if ok {
				fmt.Printf("%s:\n", catg)
				listWords("    ", reprRangeTable(rt), 60, 1)
			} else {
				log.Fatalf("Invalid category %v\n", catg)
			}
			fmt.Println("\n")
		}
	}
	// see whether to range on scripts (natural-languages)
	if len(options.scripts) > 0 {
		fmt.Println("scripts:", options.scripts)
		for _, script := range options.scripts {
			rt, ok := unicode.Scripts[script]
			if ok {
				fmt.Printf("%s:\n", script)
				listWords("    ", reprRangeTable(rt), 60, 1)
			} else {
				log.Fatalf("Invalid script %v\n", script)
			}
			fmt.Println("\n")
		}
	}
	// see whether to range on properties
	if len(options.properties) > 0 {
		fmt.Println("properties:", options.properties)
		for _, property := range options.properties {
			rt, ok := unicode.Properties[property]
			if ok {
				fmt.Printf("%s:\n", property)
				listWords("    ", reprRangeTable(rt), 60, 1)
			} else {
				log.Fatalf("Invalid property %v\n", property)
			}
			fmt.Println("\n")
		}
	}
}

func listCategories() {
	catgs := make([]string, 0, len(unicode.Categories))
	for k := range unicode.Categories {
		catgs = append(catgs, k)
	}
	sort.Strings(catgs)

	fmt.Println("categories:")
	listWords("    ", catgs, 4, 18)
}

func listScripts() {
	scripts := make([]string, 0, len(unicode.Scripts))
	for k := range unicode.Scripts {
		scripts = append(scripts, k)
	}
	sort.Strings(scripts)

	fmt.Println("scripts:")
	listWords("    ", scripts, 25, 3)
}

func listProperties() {
	properties := make([]string, 0, len(unicode.Properties))
	for k := range unicode.Properties {
		properties = append(properties, k)
	}
	sort.Strings(properties)

	fmt.Println("properties:")
	listWords("    ", properties, 40, 2)
}

func listCharAs(name string, filterfn func(r rune) bool) {

	runes := make([]rune, 0)
	strs := make([]string, 0)
	for r := rune(0); r < unicode.MaxRune; r++ {
		if filterfn(r) {
			runes = append(runes, r)
			strs = append(strs, fmt.Sprintf("%q", r))
		}
	}

	if len(options.maps) == 0 {
		fmt.Printf("%s:\n", name)
		listWords("    ", strs, 10, 7)
	}

	for _, mapby := range options.maps {
		fn := mapbyFns[mapby]
		if fn == nil {
			log.Fatalf("unknown mapby %v\n", mapby)
		}
		m := make(map[string][]string)
		ks := make([]string, 0)
		for _, r := range runes {
			k := fn(r)
			if _, ok := m[k]; !ok {
				m[k] = make([]string, 0)
				ks = append(ks, k)
			}
			m[k] = append(m[k], fmt.Sprintf("%q", r))
		}
		sort.Strings(ks)

		for _, k := range ks {
			fmt.Printf("%s, %s:\n", name, k)
			listWords("    ", m[k], 10, 7)
			fmt.Println("\n")
		}
	}
}

func listWords(prefix string, words []string, width, perline int) {
	format := fmt.Sprintf("%%-%dv", width)
	i := 1
	fmt.Printf("%s", prefix)
	for _, word := range words {
		fmt.Printf(format, word)
		if i >= perline {
			fmt.Printf("\n%s", prefix)
			i = 0
		}
		i++
	}
}

func reprRangeTable(rt *unicode.RangeTable) []string {
	strs := make([]string, 0)
	for _, r16 := range rt.R16 {
		l, h, s := r16.Lo, r16.Hi, r16.Stride
		str := fmt.Sprintf("R16 %12q(%8x) -> %12q(%8x) #%d", l, l, h, h, s)
		strs = append(strs, str)
	}
	for _, r32 := range rt.R32 {
		l, h, s := r32.Lo, r32.Hi, r32.Stride
		str := fmt.Sprintf("R32 %12q(%8x) -> %12q(%8x) #%d", l, l, h, h, s)
		strs = append(strs, str)
	}
	return strs
}

func runesCategory(r rune) string {
	for name, rt := range unicode.Categories {
		if unicode.Is(rt, r) {
			return name
		}
	}
	return "--na--"
}

func runesScript(r rune) string {
	for name, rt := range unicode.Scripts {
		if unicode.Is(rt, r) {
			return name
		}
	}
	return "--na--"
}

func runesProperties(r rune) string {
	for name, rt := range unicode.Properties {
		if unicode.Is(rt, r) {
			return name
		}
	}
	return "--na--"
}
