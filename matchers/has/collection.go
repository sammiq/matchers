package has

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/sammiq/charmset"
	"github.com/sammiq/charmset/internal"
	"github.com/sammiq/charmset/matchers/is"
)

func everyInSliceMatch(matcher charmset.Matcher, actual reflect.Value) (err error) {
	if actual.Len() == 0 {
		return errors.New("was empty")
	}
	for i := 0; i < actual.Len(); i++ {
		value := actual.Index(i)
		err = matcher.Match(value.Interface())
		if err != nil {
			return fmt.Errorf("contained an item where %s", err)
		}
	}
	return nil
}

func anyInSliceMatch(matcher charmset.Matcher, actual reflect.Value) (err error) {
	if actual.Len() == 0 {
		return errors.New("was empty")
	}
	errs := make([]string, 0, actual.Len())
	for i := 0; i < actual.Len(); i++ {
		value := actual.Index(i)
		err = matcher.Match(value.Interface())
		if err == nil {
			return nil
		}
		errs = append(errs, err.Error())
	}
	if len(errs) == 1 {
		return fmt.Errorf("contained an item where %s", errs[0])
	}
	return fmt.Errorf("no item matched where [\n          %s\n          ]", strings.Join(errs, ",\n          "))
}

func matchSliceSequence(expected []interface{}, actual reflect.Value) (err error) {
	matchIndex := 0
	matchLen := len(expected)
	for i := 0; i < actual.Len() && matchIndex < matchLen; {
		err = internal.Equal(expected[matchIndex], actual.Index(i).Interface())
		if err == nil {
			i++
			matchIndex++
			if matchIndex == matchLen {
				return nil
			}
		} else {
			if matchIndex == 0 {
				//if you always move to the next item you will skip rechecking
				//current index on the first element in the other slice
				i++
			}
			matchIndex = 0
		}
	}

	if matchIndex < matchLen {
		if matchLen-matchIndex > 1 {
			return fmt.Errorf("did not contain <%v>", expected[matchIndex:])
		}
		return fmt.Errorf("did not contain <%v>", expected[matchIndex])
	}
	return nil
}

// EveryItemMatching returns a matcher that checks whether each element of an array
// or slice matches a given matcher. Returns early if an item does not match.
func EveryItemMatching(matcher charmset.Matcher) *charmset.MatcherType {
	return charmset.NewMatcher(
		fmt.Sprintf("every item to have %s", matcher.Description()),
		func(actual interface{}) error {
			actualValue := reflect.ValueOf(actual)
			switch actualValue.Kind() {
			case reflect.Array, reflect.Slice:
				return everyInSliceMatch(matcher, actualValue)
			default:
				return errors.New("was not a slice or array")
			}
		},
	)
}

// AnyItemMatching returns a matcher that checks whether any element of an array
// or slice matches a given matcher. Returns early if an item matches.
func AnyItemMatching(matcher charmset.Matcher) *charmset.MatcherType {
	return charmset.NewMatcher(
		fmt.Sprintf("any item to have %s", matcher.Description()),
		func(actual interface{}) error {
			actualValue := reflect.ValueOf(actual)
			switch actualValue.Kind() {
			case reflect.Array, reflect.Slice:
				return anyInSliceMatch(matcher, actualValue)
			default:
				return errors.New("was not a slice or array")
			}
		},
	)
}

// Item returns a matcher that checks whether any element of an array
// or slice matches a given value. Returns early if an item matches.
func Item(expected interface{}) *charmset.MatcherType {
	return AnyItemMatching(is.EqualTo(expected))
}

// ItemIn returns a matcher that checks whether any element of an array
// or slice matches any of a set of given values. Returns early if a match is found.
func ItemIn(expected ...interface{}) *charmset.MatcherType {
	return AnyItemMatching(is.OneOf(expected...))
}

// Items returns a matcher that checks whether all elements of a given set of values
// are contained in an array or slice. Returns early if a match is not found.
func Items(expected ...interface{}) *charmset.MatcherType {
	if len(expected) == 0 {
		//panic as there is no reason to continue the test if expected is invalid at construction
		panic("will never match an empty set of items")
	}
	return charmset.NewMatcher(
		fmt.Sprintf("values equal to <%v> in any order", expected),
		func(actual interface{}) error {
			actualValue := reflect.ValueOf(actual)
			switch actualValue.Kind() {
			case reflect.Array, reflect.Slice:
				for _, ex := range expected {
					if err := anyInSliceMatch(is.EqualTo(ex), actualValue); err != nil {
						return err
					}
				}
				return nil
			default:
				return errors.New("was not a slice or array")
			}
		},
	)
}

// Sequence returns a matcher that checks whether all elements of a given set of values
// are contained in an array or slice in the order specified. Returns early if a match is found.
func Sequence(expected ...interface{}) *charmset.MatcherType {
	if len(expected) == 0 {
		//panic as there is no reason to continue the test if expected is invalid at construction
		panic("will never match an empty sequence of items")
	}
	return charmset.NewMatcher(
		fmt.Sprintf("values equal to <%v> in order", expected),
		func(actual interface{}) error {
			actualValue := reflect.ValueOf(actual)
			switch actualValue.Kind() {
			case reflect.Array, reflect.Slice:
				return matchSliceSequence(expected, actualValue)
			default:
				return errors.New("was not a slice or array")
			}
		},
	)
}
