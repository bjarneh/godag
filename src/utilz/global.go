// © Knug Industries 2009 all rights reserved
// GNU GENERAL PUBLIC LICENSE VERSION 3.0
// Author bjarneh@ifi.uio.no

package global

// be a man and use some globals.

var intMap map[string]int
var stringMap map[string]string
var floatMap map[string]float
var boolMap map[string]bool
var interfaceMap map[string]interface{}


func init() {
    intMap = make(map[string]int)
    stringMap = make(map[string]string)
    floatMap = make(map[string]float)
    boolMap = make(map[string]bool)
    interfaceMap = make(map[string]interface{})
}

// setters

func SetInt(key string, value int) {
    intMap[key] = value
}

func SetString(key, value string) {
    stringMap[key] = value
}

func SetFloat(key string, value float) {
    floatMap[key] = value
}

func SetBool(key string, value bool) {
    boolMap[key] = value
}

func SetInterface(key string, value interface{}) {
    interfaceMap[key] = value
}

// getters

func GetIntSafe(key string) (value int, ok bool) {
    value, ok = intMap[key]
    return value, ok
}

func GetInt(key string) int {
    return intMap[key]
}

func GetStringSafe(key string) (value string, ok bool) {
    value, ok = stringMap[key]
    return value, ok
}

func GetString(key string) string {
    return stringMap[key]
}

func GetFloatSafe(key string) (value float, ok bool) {
    value, ok = floatMap[key]
    return value, ok
}

func GetFloat(key string) float {
    return floatMap[key]
}

func GetBoolSafe(key string) (value, ok bool) {
    value, ok = boolMap[key]
    return value, ok
}

func GetBool(key string) bool {
    return boolMap[key]
}
