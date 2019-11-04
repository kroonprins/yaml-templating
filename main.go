package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

var (
	directiveMatcher = regexp.MustCompile("^(\\S+)\\[(.+)\\]")
	actionMatcher    = regexp.MustCompile("^#(\\S+)\\s+(.+)")
	supportedActions = make(map[string]func(result, string, interface{}, environment, []string) error)
)

const (
	root = "~"
)

type template = map[interface{}]interface{}
type environment = map[interface{}]interface{}
type result = map[interface{}]interface{}
type object = map[interface{}]interface{}

func main() {
	tpl, err := readYamlFile("template.yaml")
	check(err)
	// fmt.Printf("---\n%v\n", tpl)
	environment, err := readYamlFile("values.yaml")
	check(err)

	configureActions()
	res, err := process(tpl, environment)
	check(err)

	d, err := yaml.Marshal(&res)
	check(err)

	fmt.Printf("---\n%s\n", string(d))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func readYamlFile(path string) (map[interface{}]interface{}, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func configureActions() {
	supportedActions["merge"] = merge
	supportedActions["append"] = _append
	supportedActions["prepend"] = prepend
	supportedActions["insert"] = insert
	supportedActions["if"] = _if
	supportedActions["value"] = value
	supportedActions["repeat"] = repeat
	supportedActions["include"] = include
}

func process(tpl template, env environment) (result, error) {
	res := make(result)
	for key, value := range tpl {
		// fmt.Printf("--- part:\n%s: %s\n\n", key, value)
		if hasDirective, directives, nodeName := directive(key); hasDirective {
			// fmt.Printf("directive: %s for %s\n\n", directive, nodeName)
			noSupportedDirective := true
			for _, directive := range directives {
				if hasAction, action, arguments := action(directive); hasAction {
					// fmt.Printf("action: %s\n\n", action)
					if handler, supported := supportedActions[action]; supported {
						noSupportedDirective = false
						handler(res, nodeName, value, env, arguments)
					}
				}
			}
			if noSupportedDirective {
				val, err := processValue(value, env)
				if err != nil {
					return nil, err
				}
				res[key] = val
			}
		} else {
			val, err := processValue(value, env)
			if err != nil {
				return nil, err
			}
			res[key] = val
		}
	}
	return res, nil
}

func processValue(value interface{}, env environment) (interface{}, error) {
	if val, ok := value.(map[interface{}]interface{}); ok {
		return process(val, env)
	}
	return value, nil
}

func directive(node interface{}) (bool, []string, string) {
	matches := directiveMatcher.FindStringSubmatch(node.(string))
	if matches == nil || len(matches) <= 2 {
		return false, nil, ""
	}
	return true, strings.Split(matches[2], "]["), matches[1] // has directive, directive, node name
}

func action(directive string) (bool, string, []string) {
	matches := actionMatcher.FindStringSubmatch(directive)
	if matches == nil || len(matches) <= 2 {
		return false, "", nil
	}
	return true, matches[1], strings.Split(matches[2], " ") // has action, action, arguments
}

func merge(result result, nodeName string, value interface{}, env environment, arguments []string) error {
	for _, argument := range arguments {
		obj, err := processValue(eval(env, argument), env)
		if err != nil {
			return err
		}
		mergeObject(result, nodeName, obj)
	}
	obj, err := processValue(value, env)
	if err != nil {
		return err
	}
	mergeObject(result, nodeName, obj)
	return nil
}

func mergeObject(result result, nodeName string, obj interface{}) {
	if nodeName == root {
		for key, value := range obj.(object) {
			result[key] = value
		}
	} else {
		if _, exists := result[nodeName]; exists {
			for key, value := range obj.(object) {
				result[nodeName].(object)[key] = value
			}
		} else {
			result[nodeName] = obj
		}
	}
}

func _append(result result, nodeName string, values interface{}, env environment, arguments []string) error {
	return insertAt(-1, arguments[0], result, nodeName, values, env)
}

func prepend(result result, nodeName string, values interface{}, env environment, arguments []string) error {
	return insertAt(0, arguments[0], result, nodeName, values, env)
}

func insert(result result, nodeName string, values interface{}, env environment, arguments []string) error {
	obj := arguments[0]
	index, err := strconv.Atoi(arguments[1])
	if err != nil {
		return err
	}
	return insertAt(index, obj, result, nodeName, values, env)
}

func insertAt(index int, expression string, result result, nodeName string, values interface{}, env environment) error {
	var initialList []interface{}
	if result[nodeName] != nil {
		initialList = result[nodeName].([]interface{})
	} else {
		for _, value := range values.([]interface{}) {
			obj, err := processValue(value, env)
			if err != nil {
				return err
			}
			initialList = append(initialList, obj)
		}
	}

	// fmt.Printf("Requested: %d, %s\n", index, expression)
	// fmt.Printf("Initial list: %v\n", initialList)
	if index < 0 {
		index = len(initialList) - index - 1
	}
	// fmt.Printf("Index: %d\n", index)

	var list []interface{}

	list = append(list, initialList[0:index]...)
	startIdxRemaining := len(list)
	// fmt.Printf("Step 1: %v\n", list)
	// fmt.Printf("Remaining: %d\n", startIdxRemaining)

	obj, err := processValue(eval(env, expression), env)
	if err != nil {
		return err
	}
	if isList(obj) {
		list = append(list, obj.([]interface{})...)
	} else {
		list = append(list, obj)
	}

	// fmt.Printf("Step 2: %v\n", list)

	list = append(list, initialList[startIdxRemaining:]...)
	// TODO, possibly some mem clean up to do

	// fmt.Printf("Resulting list: %v\n\n", list)

	result[nodeName] = list
	return nil
}

func isList(obj interface{}) bool {
	_, isList := obj.([]interface{})
	return isList
}

func _if(result result, nodeName string, value interface{}, env environment, arguments []string) error {
	if val := eval(env, arguments[0]); val != nil {
		obj, err := processValue(value, env)
		if err != nil {
			return err
		}
		result[nodeName] = obj
	}
	return nil
}

func value(result result, nodeName string, value interface{}, env environment, arguments []string) error {
	res := value.(string)
	for idx, argument := range arguments {
		res = strings.ReplaceAll(res, fmt.Sprintf("\\%d", idx+1), eval(env, argument).(string))
	}
	result[nodeName] = res
	return nil
}

func repeat(result result, nodeName string, value interface{}, env environment, arguments []string) error {
	var list []interface{}
	for _, argument := range arguments {
		// value is content to repeat
		tpl := value.([]interface{})[0]

		// argument = "myobjectlist"
		// obj = myobjectlist
		obj := eval(env, argument)

		// loop over myobjectlist
		for _, item := range obj.([]interface{}) {
			// add item to environment as $item
			env["$item"] = item

			//template
			res, err := processValue(tpl, env)
			if err != nil {
				return err
			}

			//add to result list
			list = append(list, res)
		}
	}

	// remove $item again
	delete(env, "$item")

	result[nodeName] = list
	return nil
}

func include(result result, nodeName string, value interface{}, env environment, arguments []string) error {
	path := arguments[0]
	tpl, err := readYamlFile(path)
	if err != nil {
		return err
	}

	obj, err := processValue(tpl, env)
	if err != nil {
		return err
	}
	mergeObject(result, nodeName, obj)

	return nil
}

func eval(env environment, expression string) interface{} {
	parts := strings.Split(strings.TrimPrefix(expression, "."), ".")
	var res interface{}
	res = env
	for _, part := range parts {
		res = res.(environment)[part]
	}
	return res
}
