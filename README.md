# DJSON

[![Build Status](https://travis-ci.com/moikot/djson.svg?branch=master)](https://travis-ci.com/moikot/djson)
[![Go Report Card](https://goreportcard.com/badge/github.com/moikot/djson)](https://goreportcard.com/report/github.com/moikot/djson)
[![Coverage Status](https://coveralls.io/repos/github/moikot/djson/badge.svg?branch=master)](https://coveralls.io/github/moikot/djson?branch=master)
[![GoDoc](https://godoc.org/github.com/moikot/djson?status.svg)](https://godoc.org/github.com/moikot/djson)

DJSON provides you with a simple way to convert a sting to an object.

For example, `key=val` will be converted to:
```go
map[string]interface{}{
  "key": "val",
},
```       

## Why?

You might find DJSON useful in a case when you need to pass a parameter into a program for changing some part of a JSON defined elsewhere. 
DJSON parsing is similar to Helm's `--set` and `--set-string` parameters parsing combined ([see here](https://github.com/helm/helm/blob/master/docs/using_helm.md)).

## Installation

```bash
go get -u github.com/moikot/djson
```

## Usage

```go
package main

import (
	"log"
	
	"github.com/moikot/djson"
)

func main() {
	if m, err := djson.Parse("foo=bar"); err == nil {
		log.Printf("%v", m)
	} else {
		log.Printf("unable to parse: %v", err)
	}
}
```

The expected output is:
```bash
map[foo:bar]
```

## Syntax  

### Maps

The root is always a map key in a DJSON expression. Like in the example above `key` is a root key in the `key=val` expression. However, after a root key you can define another, nested one using a dot as the key separator, so that `key1.key2=val` will be converted to:
```go
map[string]interface{}{
  "key1": map[string]interface{}{
    "key2": "val",
  },
},
```       

You can define several key-value pairs by separating them using commas. In this case, the expressions you define can be merged into one map object. For example: `key1=val1,key2=val2` will be converted and merged to:
```go
map[string]interface{}{
  "key1": "val1",
  "key2": "val2",
},
```   

Or you can merge a map into another one `key1=val1,key2.key3=val2` and the result will be:
```go
map[string]interface{}{
  "key1": "val1",
  "key2": map[string]interface{}{
    "key3": "val2",
  },
},
```

If the case of using the same key in several key-value expressions, the rightmost one takes precedence, for example in `key1=val1,key2.key1=val2` value `val2` overrides `val` and the result will be:
```go
map[string]interface{}{
  "key1": "val2",
},
```

### Arrays

Apart of the map keys you can also specify indexes of array elements where you want to set or override a value. Assuming that the root should always be a map key, you can provide `key[0]=val` and it will be converted into a map containing an array of values (one value in this particular case):
```go
map[string]interface{}{
  "key": []interface{}{
    "val",
  },
},
```

If you need to compose an array from several values you can separate them one after another using a comma `','`, and `key[0]=val1,key[1]=val2` will be deserialized into:
```go
map[string]interface{}{
  "key": []interface{}{
    "val1",
    "val2",
  },
},
```

If you miss some items in the array when you set your value, those values will default to nil if they are not previously defined. In this example, `key[0]=val1,key[2]=val2` index `1` is missing and the string will be converted to:
```go
map[string]interface{}{
  "key": []interface{}{
    "val1",
    nil,
    "val2",
  },
},
``` 

In addition, instead of setting array values one by one, you may choose to specify an array as the value you set. For example `key={val1,val2}` will be converted to:
```go
map[string]interface{}{
  "key": []interface{}{
    "val1",
    "val2",
  },
},
```

There is a difference between providing an array value and specifying values in an array one by one using indexes. In the former case, you completely override the destination array. 

If you skip some values in the array you provide as a value, those elements will be replaced by empty strings, e.g. `key={val1,,val2}` will be converted to:
```go
map[string]interface{}{
  "key": []interface{}{
    "val1",
    "",
    "val2",
  },
},
```  

You can chain array indexes and map keys in different order as long as the first item in the chain is a map key. String `key1[0].key2=val`, for example, will be deserialized to:
```go
map[string]interface{}{
  "key": []interface{}{
    map[string]interface{}{
      "key2": "val",
    },
  },
},
```  

### Values conversion

DJSON parser attempts to convert strings you provide into different types. For example, string `"true"` will be automatically converted to a Boolean value `true`, and `key=true` string will be deserialized into:
```go
map[string]interface{}{
  "key": true,
},
```      

In order to supress the automatic conversion, you can enclose a value you provide into single quotes like `key='true'`, so that the value can escape the conversion.
```go
map[string]interface{}{
  "key": "true",
},
```      

All the recognized types are listed in the following table:

| Type | Example |
|:------|:------|
|boolean | `true`, `false`|
|64-bit integer | `1000`, `-1000`|
|64-bit float | `10.01`, `4e-04` |
|null value| null|

The `null` value is needed for representing an uninitialized value so that such strings as `key=null` can be deserialized into:
```go
map[string]interface{}{
  "key": nil,
},
```      

### Escaping 

Some characters have special meaning in the map keys and value definitions. For example, characters `'.'`, `'['`, `']'` and `'='` cannot be directly used in map keys because they have special meaning. E.g. character `'.'` separates map keys and if you define `part1.part2=val`, it will be converted to:
```go
map[string]interface{}{
  "part1": map[string]interface{}{
    "part2": "val",
  },
},
```   

But if you escape `'.'` using `'\'` character, input string `part1\.part2=val` will be converted to:
```go
map[string]interface{}{
  "part1.part2": "val",
},
```   

The same rule applies to the values, where symbols `','`, `'{'`, `'}'` and `'''` (a single quote) can be escaped using `'\'`. For example `key=\{val\}` will be deserialized to:
```go
map[string]interface{}{
  "key": "{val}",
},
```
