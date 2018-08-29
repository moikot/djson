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
DJSON parsing is similar to Helm's `--set` and `--set-string` parameters parsing ([see here](https://github.com/helm/helm/blob/master/docs/using_helm.md)).

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
  m := map[string]interface{}{}
  if err := djson.MergeValue(m, "foo=bar"); err == nil {
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

### Maps and arrays

The root is always a map key in a DJSON expression. Like in the example above `key` is a root key in the `key=val` expression. However, after a root key you can define another, nested one using a dot as the key separator, so that `key1.key2=val` will be converted to:
```go
map[string]interface{}{
  "key1": map[string]interface{}{
    "key2": "val",
  },
},
```       

Apart of the map keys you can also specify indexes of array elements where you want to set or override a value. Assuming that the root should always be a map key, you can provide `key[0]=val` and it will be converted into a map containing an array of values (one value in this particular case):
```go
map[string]interface{}{
  "key": []interface{}{
    "val",
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

`MergeValue` always attempts to convert strings you provide into different types. For example, value `"true"` will be automatically converted to a Boolean value `true`, and `key=true` string will be deserialized into:
```go
map[string]interface{}{
  "key": true,
},
```      

In order to suppress the automatic conversion, you can use `AppendString` instead and the result will be:
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

## Merging

If you call sequentially call `MergeValue` and `MergeString` in any order, the result of an individual call will be merged into the map provided using some simple rules. For example, merging the following strings `key1=val1` and `key2=val2` you get the following result:
```go
map[string]interface{}{
  "key1": "val1",
  "key2": "val2",
},
```   

You can also merge a map into another one. In case of `key1=val1` and `key2.key3=val2` merged, result will be:
```go
map[string]interface{}{
  "key1": "val1",
  "key2": map[string]interface{}{
    "key3": "val2",
  },
},
```

If the case of using the same key in several key-value expressions, the last call takes precedence, for example, when `key1=val1` merged with `key1=val2` one after another, value `val2` overrides `val1` and the result will be:
```go
map[string]interface{}{
  "key1": "val2",
},
```

If you need to compose an array from several values you can merge the values sequentially e.g. merging `key[0]=val1` and `key[1]=val2` you will get:
```go
map[string]interface{}{
  "key": []interface{}{
    "val1",
    "val2",
  },
},
```

If you miss some items in the array when you merge a value, those values will default to nil if they are not previously defined. In this example merging `key[0]=val1` and `key[2]=val2` index `1` was missed and the result will be:
```go
map[string]interface{}{
  "key": []interface{}{
    "val1",
    nil,
    "val2",
  },
},
```   

## Escaping

Some characters have special meaning in the keys definition. For example, character `'.'`  separates map keys and if you define `part1.part2=val`, it will be deserialized to:
```go
map[string]interface{}{
  "part1": map[string]interface{}{
    "part2": "val",
  },
},
```   

But if you escape `'.'` using a backslash character `'\'`, input string `part1\.part2=val` will be deserialized to:
```go
map[string]interface{}{
  "part1.part2": "val",
},
```   

The following characters can be escaped in the map keys: `'.'`, `'['` and `']'`. If you try to escape any other character, the parsers will fail. In order to avoid the failure you can escape a backslash using another backslash in front of it `'\\'`.
