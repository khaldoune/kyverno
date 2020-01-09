package variables

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/nirmata/kyverno/pkg/engine/context"
	"github.com/nirmata/kyverno/pkg/engine/operator"
)

//SubstituteVariables substitutes the JMESPATH with variable substitution
// supported substitutions
// - no operator + variable(string,object)
// unsupported substitutions
// - operator + variable(object) -> as we dont support operators with object types
func SubstituteVariables(ctx context.EvalInterface, pattern interface{}, path string) interface{} {
	// var err error
	switch typedPattern := pattern.(type) {
	case map[string]interface{}:
		return substituteMap(ctx, typedPattern, path)
	case []interface{}:
		return substituteArray(ctx, typedPattern, path)
	case string:
		// variable substitution
		return substituteValue(ctx, typedPattern, path)
	default:
		return pattern
	}
}
func escape(key string) string {
	return fmt.Sprintf("\"%s\"", key)
}

func substituteMap(ctx context.EvalInterface, patternMap map[string]interface{}, path string) map[string]interface{} {
	for key, patternElement := range patternMap {
		currentPath := path + "." + escape(key)
		value := SubstituteVariables(ctx, patternElement, currentPath)
		patternMap[key] = value
	}
	return patternMap
}

func substituteArray(ctx context.EvalInterface, patternList []interface{}, path string) []interface{} {
	for idx, patternElement := range patternList {
		currentPath := escape(path+strconv.Itoa(idx)) + "."
		value := SubstituteVariables(ctx, patternElement, currentPath)
		patternList[idx] = value
	}
	return patternList
}

func substituteValue(ctx context.EvalInterface, valuePattern string, path string) interface{} {
	// patterns supported
	// - operator + string
	// operator + variable
	operatorVariable := getOperator(valuePattern)
	variable := valuePattern[len(operatorVariable):]
	// substitute variable with value
	value := getValueQuery(ctx, variable, path)
	if operatorVariable == "" {
		// default or operator.Equal
		// equal + string variable
		// object variable
		return value
	}
	// operator + string variable
	switch value.(type) {
	case string:
		return string(operatorVariable) + value.(string)
	default:
		glog.Infof("cannot use operator with object variables. operator used %s in pattern %v", string(operatorVariable), valuePattern)
		var emptyInterface interface{}
		return emptyInterface
	}
}

func getValueQuery(ctx context.EvalInterface, valuePattern string, path string) interface{} {
	glog.V(4).Infof("path:= %s", path)
	var emptyInterface interface{}
	// extract variable {{<variable>}}
	validRegex := regexp.MustCompile(`\{\{([^{}]*)\}\}`)
	groups := validRegex.FindAllStringSubmatch(valuePattern, -1)
	// can have multiple variables in a single value pattern
	// var Map <variable,value>
	varMap := getValues(ctx, groups, path)
	if len(varMap) == 0 {
		// there are no varaiables
		// return the original value
		return valuePattern
	}
	// only substitute values if all the variable values are of type string
	if isAllVarStrings(varMap) {
		newVal := valuePattern
		for key, value := range varMap {
			if val, ok := value.(string); ok {
				newVal = strings.Replace(newVal, key, val, -1)
			}
		}
		return newVal
	}

	// we do not support mutliple substitution per statement for non-string types
	for _, value := range varMap {
		return value
	}
	return emptyInterface
}

// returns map of variables as keys and variable values as values
func getValues(ctx context.EvalInterface, groups [][]string, path string) map[string]interface{} {
	var emptyInterface interface{}
	subs := map[string]interface{}{}
	for _, group := range groups {
		if len(group) == 2 {
			// 0th is string
			// 1st is the capture group
			if group[1] == "@this" {
				glog.V(4).Infof("@this found at path %s", path)
			}
			variable, err := ctx.Query(group[1])
			if err != nil {
				glog.V(4).Infof("variable substitution failed for query %s: %v", group[0], err)
				subs[group[0]] = emptyInterface
				continue
			}
			if variable == nil {
				subs[group[0]] = emptyInterface
			} else {
				subs[group[0]] = variable
			}
		}
	}
	return subs
}

func isAllVarStrings(subVar map[string]interface{}) bool {
	for _, value := range subVar {
		if _, ok := value.(string); !ok {
			return false
		}
	}
	return true
}

func getOperator(pattern string) string {
	operatorVariable := operator.GetOperatorFromStringPattern(pattern)
	if operatorVariable == operator.Equal {
		return ""
	}
	return string(operatorVariable)
}
