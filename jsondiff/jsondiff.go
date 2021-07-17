package jsondiff

import (
	"fmt"
	"reflect"
	"sort"
)

func JSONDiff(j1 interface{}, j2 interface{}, ignoreZeroValue bool, path string) []string {
	if map1, isMap1 := j1.(map[string]interface{}); isMap1 {
		if map2, isMap2 := j2.(map[string]interface{}); isMap2 {
			return jmapEqual(map1, map2, ignoreZeroValue, path)
		}
	} else if arr1, isArr1 := j1.([]interface{}); isArr1 {
		if arr2, isArr2 := j2.([]interface{}); isArr2 {
			return jarrayEqual(arr1, arr2, ignoreZeroValue, path)
		}
	} else if prim1, isPrim1 := j1.(string); isPrim1 {
		if prim2, isPrim2 := j2.(string); isPrim2 {
			if prim1 != prim2 {
				return []string{fmt.Sprintf("Strings unequal at '%s'\n 1:%s\n 2:%s", path, prim1, prim2)}
			}
			return []string{}
		}
	} else if prim1, isPrim1 := j1.(float64); isPrim1 {
		if prim2, isPrim2 := j2.(float64); isPrim2 {
			if prim1 != prim2 {
				return []string{fmt.Sprintf("Numbers unequal at '%s'\n 1:%f\n 2:%f", path, prim1, prim2)}
			}
			return []string{}
		}
	} else if prim1, isPrim1 := j1.(bool); isPrim1 {
		if prim2, isPrim2 := j2.(bool); isPrim2 {
			if prim1 != prim2 {
				return []string{fmt.Sprintf("Bools unequal at '%s'\n 1:%t\n 2:%t", path, prim1, prim2)}
			}
			return []string{}
		}
	} else if j1 == nil && j2 == nil {
		return []string{}
	}

	return []string{fmt.Sprintf("Uncomparable types at '%s'\n 1:%v\n 2:%v",
		path,
		reflect.TypeOf(j1),
		reflect.TypeOf(j2))}
}

func jmapEqual(j1 map[string]interface{}, j2 map[string]interface{}, ignoreZeroValue bool, path string) []string {
	diff := make([]string, 0)

	var sortedKeys1 []string
	for key := range j1 {
		sortedKeys1 = append(sortedKeys1, key)
	}
	sort.Strings(sortedKeys1)

	var sortedKeys2 []string
	for key := range j2 {
		sortedKeys2 = append(sortedKeys2, key)
	}
	sort.Strings(sortedKeys2)

	for _, key := range sortedKeys1 {
		val1 := j1[key]
		if val2, exists := j2[key]; exists {
			subDiff := JSONDiff(val1, val2, ignoreZeroValue, path+"."+key)
			diff = append(diff, subDiff...)
		} else {
			if !ignoreZeroValue || !isZeroValue(val1) {
				diff = append(diff, "Map key '"+path+"."+key+"' is present in J1 but not J2")
			}
		}
	}

	for _, key := range sortedKeys2 {
		val2 := j2[key]
		if _, exists := j1[key]; !exists {
			if !ignoreZeroValue || !isZeroValue(val2) {
				diff = append(diff, "Map key '"+path+"."+key+"' is present in J2 but not J1")
			}
		}
	}

	return diff
}

func jarrayEqual(j1 []interface{}, j2 []interface{}, ignoreZeroValue bool, path string) []string {
	if len(j1) != len(j2) {
		return []string{fmt.Sprintf("Arrays at '%s' differ in length\n 1:%d\n 2:%d", path, len(j1), len(j2))}
	}

	diff := make([]string, 0)
	for i := range j1 {
		subDiff := JSONDiff(j1[i], j2[i], ignoreZeroValue, fmt.Sprintf("%s[%d]", path, i))
		diff = append(diff, subDiff...)
	}

	return diff
}

func isZeroValue(jVal interface{}) bool {
	if jVal == nil {
		return true
	} else if jList, isList := jVal.([]interface{}); isList {
		return len(jList) == 0
	} else if jMap, isMap := jVal.(map[string]interface{}); isMap {
		return len(jMap) == 0
	} else if jNum, isNum := jVal.(float64); isNum {
		return jNum == 0
	} else if jBool, isBool := jVal.(bool); isBool {
		return jBool == false
	} else if jStr, isStr := jVal.(string); isStr {
		return jStr == ""
	}

	return false
}
