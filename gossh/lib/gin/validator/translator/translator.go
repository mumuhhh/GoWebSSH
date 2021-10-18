// Code generated by golang.org/x/tools/cmd/bundle. DO NOT EDIT.

package translator

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	// ErrUnknowTranslation indicates the translation could not be found
	ut_ErrUnknowTranslation = errors.New("Unknown Translation")
)

var _ error = new(ut_ErrConflictingTranslation)

var _ error = new(ut_ErrRangeTranslation)

var _ error = new(ut_ErrOrdinalTranslation)

var _ error = new(ut_ErrCardinalTranslation)

var _ error = new(ut_ErrMissingPluralTranslation)

var _ error = new(ut_ErrExistingTranslator)

// ErrExistingTranslator is the error representing a conflicting translator
type ut_ErrExistingTranslator struct {
	locale string
}

// Error returns ErrExistingTranslator's internal error text
func (e *ut_ErrExistingTranslator) Error() string {
	return fmt.Sprintf("error: conflicting translator for locale '%s'", e.locale)
}

// ErrConflictingTranslation is the error representing a conflicting translation
type ut_ErrConflictingTranslation struct {
	locale string
	key  interface{}
	rule PluralRule
	text string
}

// Error returns ErrConflictingTranslation's internal error text
func (e *ut_ErrConflictingTranslation) Error() string {

	if _, ok := e.key.(string); !ok {
		return fmt.Sprintf("error: conflicting key '%#v' rule '%s' with text '%s' for locale '%s', value being ignored", e.key, e.rule, e.text, e.locale)
	}

	return fmt.Sprintf("error: conflicting key '%s' rule '%s' with text '%s' for locale '%s', value being ignored", e.key, e.rule, e.text, e.locale)
}

// ErrRangeTranslation is the error representing a range translation error
type ut_ErrRangeTranslation struct {
	text string
}

// Error returns ErrRangeTranslation's internal error text
func (e *ut_ErrRangeTranslation) Error() string {
	return e.text
}

// ErrOrdinalTranslation is the error representing an ordinal translation error
type ut_ErrOrdinalTranslation struct {
	text string
}

// Error returns ErrOrdinalTranslation's internal error text
func (e *ut_ErrOrdinalTranslation) Error() string {
	return e.text
}

// ErrCardinalTranslation is the error representing a cardinal translation error
type ut_ErrCardinalTranslation struct {
	text string
}

// Error returns ErrCardinalTranslation's internal error text
func (e *ut_ErrCardinalTranslation) Error() string {
	return e.text
}

// ErrMissingPluralTranslation is the error signifying a missing translation given
// the locales plural rules.
type ut_ErrMissingPluralTranslation struct {
	locale          string
	key             interface{}
	rule            PluralRule
	translationType string
}

// Error returns ErrMissingPluralTranslation's internal error text
func (e *ut_ErrMissingPluralTranslation) Error() string {

	if _, ok := e.key.(string); !ok {
		return fmt.Sprintf("error: missing '%s' plural rule '%s' for translation with key '%#v' and locale '%s'", e.translationType, e.rule, e.key, e.locale)
	}

	return fmt.Sprintf("error: missing '%s' plural rule '%s' for translation with key '%s' and locale '%s'", e.translationType, e.rule, e.key, e.locale)
}

// ErrMissingBracket is the error representing a missing bracket in a translation
// eg. This is a {0 <-- missing ending '}'
type ut_ErrMissingBracket struct {
	locale string
	key    interface{}
	text   string
}

// Error returns ErrMissingBracket error message
func (e *ut_ErrMissingBracket) Error() string {
	return fmt.Sprintf("error: missing bracket '{}', in translation. locale: '%s' key: '%v' text: '%s'", e.locale, e.key, e.text)
}

// ErrBadParamSyntax is the error representing a bad parameter definition in a translation
// eg. This is a {must-be-int}
type ut_ErrBadParamSyntax struct {
	locale string
	param  string
	key    interface{}
	text   string
}

// Error returns ErrBadParamSyntax error message
func (e *ut_ErrBadParamSyntax) Error() string {
	return fmt.Sprintf("error: bad parameter syntax, missing parameter '%s' in translation. locale: '%s' key: '%v' text: '%s'", e.param, e.locale, e.key, e.text)
}

// import/export errors

// ErrMissingLocale is the error representing an expected locale that could
// not be found aka locale not registered with the UniversalTranslator Instance
type ut_ErrMissingLocale struct {
	locale string
}

// Error returns ErrMissingLocale's internal error text
func (e *ut_ErrMissingLocale) Error() string {
	return fmt.Sprintf("error: locale '%s' not registered.", e.locale)
}

// ErrBadPluralDefinition is the error representing an incorrect plural definition
// usually found within translations defined within files during the import process.
type ut_ErrBadPluralDefinition struct {
	tl ut_translation
}

// Error returns ErrBadPluralDefinition's internal error text
func (e *ut_ErrBadPluralDefinition) Error() string {
	return fmt.Sprintf("error: bad plural definition '%#v'", e.tl)
}

type ut_translation struct {
	Locale           string      `json:"locale"`
	Key              interface{} `json:"key"` // either string or integer
	Translation      string      `json:"trans"`
	PluralType       string      `json:"type,omitempty"`
	PluralRule       string      `json:"rule,omitempty"`
	OverrideExisting bool        `json:"override,omitempty"`
}

const (
	ut_cardinalType = "Cardinal"
	ut_ordinalType  = "Ordinal"
	ut_rangeType    = "Range"
)

// ImportExportFormat is the format of the file import or export
type ut_ImportExportFormat uint8

// supported Export Formats
const (
	ut_FormatJSON ut_ImportExportFormat = iota
)

// Export writes the translations out to a file on disk.
//
// NOTE: this currently only works with string or int translations keys.
func (t *ut_UniversalTranslator) Export(format ut_ImportExportFormat, dirname string) error {

	_, err := os.Stat(dirname)
	fmt.Println(dirname, err, os.IsNotExist(err))
	if err != nil {

		if !os.IsNotExist(err) {
			return err
		}

		if err = os.MkdirAll(dirname, 0744); err != nil {
			return err
		}
	}

	// build up translations
	var trans []ut_translation
	var b []byte
	var ext string

	for _, locale := range t.translators {

		for k, v := range locale.(*ut_translator).translations {
			trans = append(trans, ut_translation{
				Locale:      locale.Locale(),
				Key:         k,
				Translation: v.text,
			})
		}

		for k, pluralTrans := range locale.(*ut_translator).cardinalTanslations {

			for i, plural := range pluralTrans {

				// leave enough for all plural rules
				// but not all are set for all languages.
				if plural == nil {
					continue
				}

				trans = append(trans, ut_translation{
					Locale:      locale.Locale(),
					Key:         k.(string),
					Translation: plural.text,
					PluralType:  ut_cardinalType,
					PluralRule:  PluralRule(i).String(),
				})
			}
		}

		for k, pluralTrans := range locale.(*ut_translator).ordinalTanslations {

			for i, plural := range pluralTrans {

				// leave enough for all plural rules
				// but not all are set for all languages.
				if plural == nil {
					continue
				}

				trans = append(trans, ut_translation{
					Locale:      locale.Locale(),
					Key:         k.(string),
					Translation: plural.text,
					PluralType:  ut_ordinalType,
					PluralRule:  PluralRule(i).String(),
				})
			}
		}

		for k, pluralTrans := range locale.(*ut_translator).rangeTanslations {

			for i, plural := range pluralTrans {

				// leave enough for all plural rules
				// but not all are set for all languages.
				if plural == nil {
					continue
				}

				trans = append(trans, ut_translation{
					Locale:      locale.Locale(),
					Key:         k.(string),
					Translation: plural.text,
					PluralType:  ut_rangeType,
					PluralRule:  PluralRule(i).String(),
				})
			}
		}

		switch format {
		case ut_FormatJSON:
			b, err = json.MarshalIndent(trans, "", "    ")
			ext = ".json"
		}

		if err != nil {
			return err
		}

		err = ioutil.WriteFile(filepath.Join(dirname, fmt.Sprintf("%s%s", locale.Locale(), ext)), b, 0644)
		if err != nil {
			return err
		}

		trans = trans[0:0]
	}

	return nil
}

// Import reads the translations out of a file or directory on disk.
//
// NOTE: this currently only works with string or int translations keys.
func (t *ut_UniversalTranslator) Import(format ut_ImportExportFormat, dirnameOrFilename string) error {

	fi, err := os.Stat(dirnameOrFilename)
	if err != nil {
		return err
	}

	processFn := func(filename string) error {

		f, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer f.Close()

		return t.ImportByReader(format, f)
	}

	if !fi.IsDir() {
		return processFn(dirnameOrFilename)
	}

	// recursively go through directory
	walker := func(path string, info os.FileInfo, err error) error {

		if info.IsDir() {
			return nil
		}

		switch format {
		case ut_FormatJSON:
			// skip non JSON files
			if filepath.Ext(info.Name()) != ".json" {
				return nil
			}
		}

		return processFn(path)
	}

	return filepath.Walk(dirnameOrFilename, walker)
}

// ImportByReader imports the the translations found within the contents read from the supplied reader.
//
// NOTE: generally used when assets have been embedded into the binary and are already in memory.
func (t *ut_UniversalTranslator) ImportByReader(format ut_ImportExportFormat, reader io.Reader) error {

	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	var trans []ut_translation

	switch format {
	case ut_FormatJSON:
		err = json.Unmarshal(b, &trans)
	}

	if err != nil {
		return err
	}

	for _, tl := range trans {

		locale, found := t.FindTranslator(tl.Locale)
		if !found {
			return &ut_ErrMissingLocale{locale: tl.Locale}
		}

		pr := ut_stringToPR(tl.PluralRule)

		if pr == PluralRuleUnknown {

			err = locale.Add(tl.Key, tl.Translation, tl.OverrideExisting)
			if err != nil {
				return err
			}

			continue
		}

		switch tl.PluralType {
		case ut_cardinalType:
			err = locale.AddCardinal(tl.Key, tl.Translation, pr, tl.OverrideExisting)
		case ut_ordinalType:
			err = locale.AddOrdinal(tl.Key, tl.Translation, pr, tl.OverrideExisting)
		case ut_rangeType:
			err = locale.AddRange(tl.Key, tl.Translation, pr, tl.OverrideExisting)
		default:
			return &ut_ErrBadPluralDefinition{tl: tl}
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func ut_stringToPR(s string) PluralRule {

	switch s {
	case "Zero":
		return PluralRuleZero
	case "One":
		return PluralRuleOne
	case "Two":
		return PluralRuleTwo
	case "Few":
		return PluralRuleFew
	case "Many":
		return PluralRuleMany
	case "Other":
		return PluralRuleOther
	default:
		return PluralRuleUnknown
	}

}

const (
	ut_paramZero          = "{0}"
	ut_paramOne           = "{1}"
	ut_unknownTranslation = ""
)

// Translator is universal translators
// translator instance which is a thin wrapper
// around Translator instance providing
// some extra functionality
type ut_Translator interface {
	Translator

	// adds a normal translation for a particular language/locale
	// {#} is the only replacement type accepted and are ad infinitum
	// eg. one: '{0} day left' other: '{0} days left'
	Add(key interface{}, text string, override bool) error

	// adds a cardinal plural translation for a particular language/locale
	// {0} is the only replacement type accepted and only one variable is accepted as
	// multiple cannot be used for a plural rule determination, unless it is a range;
	// see AddRange below.
	// eg. in locale 'en' one: '{0} day left' other: '{0} days left'
	AddCardinal(key interface{}, text string, rule PluralRule, override bool) error

	// adds an ordinal plural translation for a particular language/locale
	// {0} is the only replacement type accepted and only one variable is accepted as
	// multiple cannot be used for a plural rule determination, unless it is a range;
	// see AddRange below.
	// eg. in locale 'en' one: '{0}st day of spring' other: '{0}nd day of spring'
	// - 1st, 2nd, 3rd...
	AddOrdinal(key interface{}, text string, rule PluralRule, override bool) error

	// adds a range plural translation for a particular language/locale
	// {0} and {1} are the only replacement types accepted and only these are accepted.
	// eg. in locale 'nl' one: '{0}-{1} day left' other: '{0}-{1} days left'
	AddRange(key interface{}, text string, rule PluralRule, override bool) error

	// creates the translation for the locale given the 'key' and params passed in
	T(key interface{}, params ...string) (string, error)

	// creates the cardinal translation for the locale given the 'key', 'num' and 'digit' arguments
	//  and param passed in
	C(key interface{}, num float64, digits uint64, param string) (string, error)

	// creates the ordinal translation for the locale given the 'key', 'num' and 'digit' arguments
	// and param passed in
	O(key interface{}, num float64, digits uint64, param string) (string, error)

	//  creates the range translation for the locale given the 'key', 'num1', 'digit1', 'num2' and
	//  'digit2' arguments and 'param1' and 'param2' passed in
	R(key interface{}, num1 float64, digits1 uint64, num2 float64, digits2 uint64, param1, param2 string) (string, error)

	// VerifyTranslations checks to ensures that no plural rules have been
	// missed within the translations.
	VerifyTranslations() error
}

var _ ut_Translator = new(ut_translator)

var _ Translator = new(ut_translator)

type ut_translator struct {
	Translator
	translations        map[interface{}]*ut_transText
	cardinalTanslations map[interface{}][]*ut_transText // array index is mapped to PluralRule index + the PluralRuleUnknown
	ordinalTanslations  map[interface{}][]*ut_transText
	rangeTanslations    map[interface{}][]*ut_transText
}

type ut_transText struct {
	text    string
	indexes []int
}

func ut_newTranslator(trans Translator) ut_Translator {
	return &ut_translator{
		Translator:          trans,
		translations:        make(map[interface{}]*ut_transText), // translation text broken up by byte index
		cardinalTanslations: make(map[interface{}][]*ut_transText),
		ordinalTanslations:  make(map[interface{}][]*ut_transText),
		rangeTanslations:    make(map[interface{}][]*ut_transText),
	}
}

// Add adds a normal translation for a particular language/locale
// {#} is the only replacement type accepted and are ad infinitum
// eg. one: '{0} day left' other: '{0} days left'
func (t *ut_translator) Add(key interface{}, text string, override bool) error {

	if _, ok := t.translations[key]; ok && !override {
		return &ut_ErrConflictingTranslation{locale: t.Locale(), key: key, text: text}
	}

	lb := strings.Count(text, "{")
	rb := strings.Count(text, "}")

	if lb != rb {
		return &ut_ErrMissingBracket{locale: t.Locale(), key: key, text: text}
	}

	trans := &ut_transText{
		text: text,
	}

	var idx int

	for i := 0; i < lb; i++ {
		s := "{" + strconv.Itoa(i) + "}"
		idx = strings.Index(text, s)
		if idx == -1 {
			return &ut_ErrBadParamSyntax{locale: t.Locale(), param: s, key: key, text: text}
		}

		trans.indexes = append(trans.indexes, idx)
		trans.indexes = append(trans.indexes, idx+len(s))
	}

	t.translations[key] = trans

	return nil
}

// AddCardinal adds a cardinal plural translation for a particular language/locale
// {0} is the only replacement type accepted and only one variable is accepted as
// multiple cannot be used for a plural rule determination, unless it is a range;
// see AddRange below.
// eg. in locale 'en' one: '{0} day left' other: '{0} days left'
func (t *ut_translator) AddCardinal(key interface{}, text string, rule PluralRule, override bool) error {

	var verified bool

	// verify plural rule exists for locale
	for _, pr := range t.PluralsCardinal() {
		if pr == rule {
			verified = true
			break
		}
	}

	if !verified {
		return &ut_ErrCardinalTranslation{text: fmt.Sprintf("error: cardinal plural rule '%s' does not exist for locale '%s' key: '%v' text: '%s'", rule, t.Locale(), key, text)}
	}

	tarr, ok := t.cardinalTanslations[key]
	if ok {
		// verify not adding a conflicting record
		if len(tarr) > 0 && tarr[rule] != nil && !override {
			return &ut_ErrConflictingTranslation{locale: t.Locale(), key: key, rule: rule, text: text}
		}

	} else {
		tarr = make([]*ut_transText, 7)
		t.cardinalTanslations[key] = tarr
	}

	trans := &ut_transText{
		text:    text,
		indexes: make([]int, 2),
	}

	tarr[rule] = trans

	idx := strings.Index(text, ut_paramZero)
	if idx == -1 {
		tarr[rule] = nil
		return &ut_ErrCardinalTranslation{text: fmt.Sprintf("error: parameter '%s' not found, may want to use 'Add' instead of 'AddCardinal'. locale: '%s' key: '%v' text: '%s'", ut_paramZero, t.Locale(), key, text)}
	}

	trans.indexes[0] = idx
	trans.indexes[1] = idx + len(ut_paramZero)

	return nil
}

// AddOrdinal adds an ordinal plural translation for a particular language/locale
// {0} is the only replacement type accepted and only one variable is accepted as
// multiple cannot be used for a plural rule determination, unless it is a range;
// see AddRange below.
// eg. in locale 'en' one: '{0}st day of spring' other: '{0}nd day of spring' - 1st, 2nd, 3rd...
func (t *ut_translator) AddOrdinal(key interface{}, text string, rule PluralRule, override bool) error {

	var verified bool

	// verify plural rule exists for locale
	for _, pr := range t.PluralsOrdinal() {
		if pr == rule {
			verified = true
			break
		}
	}

	if !verified {
		return &ut_ErrOrdinalTranslation{text: fmt.Sprintf("error: ordinal plural rule '%s' does not exist for locale '%s' key: '%v' text: '%s'", rule, t.Locale(), key, text)}
	}

	tarr, ok := t.ordinalTanslations[key]
	if ok {
		// verify not adding a conflicting record
		if len(tarr) > 0 && tarr[rule] != nil && !override {
			return &ut_ErrConflictingTranslation{locale: t.Locale(), key: key, rule: rule, text: text}
		}

	} else {
		tarr = make([]*ut_transText, 7)
		t.ordinalTanslations[key] = tarr
	}

	trans := &ut_transText{
		text:    text,
		indexes: make([]int, 2),
	}

	tarr[rule] = trans

	idx := strings.Index(text, ut_paramZero)
	if idx == -1 {
		tarr[rule] = nil
		return &ut_ErrOrdinalTranslation{text: fmt.Sprintf("error: parameter '%s' not found, may want to use 'Add' instead of 'AddOrdinal'. locale: '%s' key: '%v' text: '%s'", ut_paramZero, t.Locale(), key, text)}
	}

	trans.indexes[0] = idx
	trans.indexes[1] = idx + len(ut_paramZero)

	return nil
}

// AddRange adds a range plural translation for a particular language/locale
// {0} and {1} are the only replacement types accepted and only these are accepted.
// eg. in locale 'nl' one: '{0}-{1} day left' other: '{0}-{1} days left'
func (t *ut_translator) AddRange(key interface{}, text string, rule PluralRule, override bool) error {

	var verified bool

	// verify plural rule exists for locale
	for _, pr := range t.PluralsRange() {
		if pr == rule {
			verified = true
			break
		}
	}

	if !verified {
		return &ut_ErrRangeTranslation{text: fmt.Sprintf("error: range plural rule '%s' does not exist for locale '%s' key: '%v' text: '%s'", rule, t.Locale(), key, text)}
	}

	tarr, ok := t.rangeTanslations[key]
	if ok {
		// verify not adding a conflicting record
		if len(tarr) > 0 && tarr[rule] != nil && !override {
			return &ut_ErrConflictingTranslation{locale: t.Locale(), key: key, rule: rule, text: text}
		}

	} else {
		tarr = make([]*ut_transText, 7)
		t.rangeTanslations[key] = tarr
	}

	trans := &ut_transText{
		text:    text,
		indexes: make([]int, 4),
	}

	tarr[rule] = trans

	idx := strings.Index(text, ut_paramZero)
	if idx == -1 {
		tarr[rule] = nil
		return &ut_ErrRangeTranslation{text: fmt.Sprintf("error: parameter '%s' not found, are you sure you're adding a Range Translation? locale: '%s' key: '%v' text: '%s'", ut_paramZero, t.Locale(), key, text)}
	}

	trans.indexes[0] = idx
	trans.indexes[1] = idx + len(ut_paramZero)

	idx = strings.Index(text, ut_paramOne)
	if idx == -1 {
		tarr[rule] = nil
		return &ut_ErrRangeTranslation{text: fmt.Sprintf("error: parameter '%s' not found, a Range Translation requires two parameters. locale: '%s' key: '%v' text: '%s'", ut_paramOne, t.Locale(), key, text)}
	}

	trans.indexes[2] = idx
	trans.indexes[3] = idx + len(ut_paramOne)

	return nil
}

// T creates the translation for the locale given the 'key' and params passed in
func (t *ut_translator) T(key interface{}, params ...string) (string, error) {

	trans, ok := t.translations[key]
	if !ok {
		return ut_unknownTranslation, ut_ErrUnknowTranslation
	}

	b := make([]byte, 0, 64)

	var start, end, count int

	for i := 0; i < len(trans.indexes); i++ {
		end = trans.indexes[i]
		b = append(b, trans.text[start:end]...)
		b = append(b, params[count]...)
		i++
		start = trans.indexes[i]
		count++
	}

	b = append(b, trans.text[start:]...)

	return string(b), nil
}

// C creates the cardinal translation for the locale given the 'key', 'num' and 'digit' arguments and param passed in
func (t *ut_translator) C(key interface{}, num float64, digits uint64, param string) (string, error) {

	tarr, ok := t.cardinalTanslations[key]
	if !ok {
		return ut_unknownTranslation, ut_ErrUnknowTranslation
	}

	rule := t.CardinalPluralRule(num, digits)

	trans := tarr[rule]

	b := make([]byte, 0, 64)
	b = append(b, trans.text[:trans.indexes[0]]...)
	b = append(b, param...)
	b = append(b, trans.text[trans.indexes[1]:]...)

	return string(b), nil
}

// O creates the ordinal translation for the locale given the 'key', 'num' and 'digit' arguments and param passed in
func (t *ut_translator) O(key interface{}, num float64, digits uint64, param string) (string, error) {

	tarr, ok := t.ordinalTanslations[key]
	if !ok {
		return ut_unknownTranslation, ut_ErrUnknowTranslation
	}

	rule := t.OrdinalPluralRule(num, digits)

	trans := tarr[rule]

	b := make([]byte, 0, 64)
	b = append(b, trans.text[:trans.indexes[0]]...)
	b = append(b, param...)
	b = append(b, trans.text[trans.indexes[1]:]...)

	return string(b), nil
}

// R creates the range translation for the locale given the 'key', 'num1', 'digit1', 'num2' and 'digit2' arguments
// and 'param1' and 'param2' passed in
func (t *ut_translator) R(key interface{}, num1 float64, digits1 uint64, num2 float64, digits2 uint64, param1, param2 string) (string, error) {

	tarr, ok := t.rangeTanslations[key]
	if !ok {
		return ut_unknownTranslation, ut_ErrUnknowTranslation
	}

	rule := t.RangePluralRule(num1, digits1, num2, digits2)

	trans := tarr[rule]

	b := make([]byte, 0, 64)
	b = append(b, trans.text[:trans.indexes[0]]...)
	b = append(b, param1...)
	b = append(b, trans.text[trans.indexes[1]:trans.indexes[2]]...)
	b = append(b, param2...)
	b = append(b, trans.text[trans.indexes[3]:]...)

	return string(b), nil
}

// VerifyTranslations checks to ensures that no plural rules have been
// missed within the translations.
func (t *ut_translator) VerifyTranslations() error {

	for k, v := range t.cardinalTanslations {

		for _, rule := range t.PluralsCardinal() {

			if v[rule] == nil {
				return &ut_ErrMissingPluralTranslation{locale: t.Locale(), translationType: "plural", rule: rule, key: k}
			}
		}
	}

	for k, v := range t.ordinalTanslations {

		for _, rule := range t.PluralsOrdinal() {

			if v[rule] == nil {
				return &ut_ErrMissingPluralTranslation{locale: t.Locale(), translationType: "ordinal", rule: rule, key: k}
			}
		}
	}

	for k, v := range t.rangeTanslations {

		for _, rule := range t.PluralsRange() {

			if v[rule] == nil {
				return &ut_ErrMissingPluralTranslation{locale: t.Locale(), translationType: "range", rule: rule, key: k}
			}
		}
	}

	return nil
}

// UniversalTranslator holds all locale & translation data
type ut_UniversalTranslator struct {
	translators map[string]ut_Translator
	fallback    ut_Translator
}

// New returns a new UniversalTranslator instance set with
// the fallback locale and locales it should support
func ut_New(fallback Translator, supportedLocales ...Translator) *ut_UniversalTranslator {

	t := &ut_UniversalTranslator{
		translators: make(map[string]ut_Translator),
	}

	for _, v := range supportedLocales {

		trans := ut_newTranslator(v)
		t.translators[strings.ToLower(trans.Locale())] = trans

		if fallback.Locale() == v.Locale() {
			t.fallback = trans
		}
	}

	if t.fallback == nil && fallback != nil {
		t.fallback = ut_newTranslator(fallback)
	}

	return t
}

// FindTranslator trys to find a Translator based on an array of locales
// and returns the first one it can find, otherwise returns the
// fallback translator.
func (t *ut_UniversalTranslator) FindTranslator(locales ...string) (trans ut_Translator, found bool) {

	for _, locale := range locales {

		if trans, found = t.translators[strings.ToLower(locale)]; found {
			return
		}
	}

	return t.fallback, false
}

// GetTranslator returns the specified translator for the given locale,
// or fallback if not found
func (t *ut_UniversalTranslator) GetTranslator(locale string) (trans ut_Translator, found bool) {

	if trans, found = t.translators[strings.ToLower(locale)]; found {
		return
	}

	return t.fallback, false
}

// GetFallback returns the fallback locale
func (t *ut_UniversalTranslator) GetFallback() ut_Translator {
	return t.fallback
}

// AddTranslator adds the supplied translator, if it already exists the override param
// will be checked and if false an error will be returned, otherwise the translator will be
// overridden; if the fallback matches the supplied translator it will be overridden as well
// NOTE: this is normally only used when translator is embedded within a library
func (t *ut_UniversalTranslator) AddTranslator(translator Translator, override bool) error {

	lc := strings.ToLower(translator.Locale())
	_, ok := t.translators[lc]
	if ok && !override {
		return &ut_ErrExistingTranslator{locale: translator.Locale()}
	}

	trans := ut_newTranslator(translator)

	if t.fallback.Locale() == translator.Locale() {

		// because it's optional to have a fallback, I don't impose that limitation
		// don't know why you wouldn't but...
		if !override {
			return &ut_ErrExistingTranslator{locale: translator.Locale()}
		}

		t.fallback = trans
	}

	t.translators[lc] = trans

	return nil
}

// VerifyTranslations runs through all locales and identifies any issues
// eg. missing plural rules for a locale
func (t *ut_UniversalTranslator) VerifyTranslations() (err error) {

	for _, trans := range t.translators {
		err = trans.VerifyTranslations()
		if err != nil {
			return
		}
	}

	return
}