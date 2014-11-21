# Change Log #

## v1.1.0 ##

* [FIX] #19 Any valid int64 number larger than 2^53 or smaller than -2^53 can be correctly decoded without precision lost.

This fix will result a possible breaking change in `Result#Get` and `Result#GetField`. If a JSON field is a number, these two functions will return json.Number instead of float64.

The fix also introduces a side effect in `Result#Decode` and `Result#DecodeField`. A number field (`int*` and `float*`) can be decoded to a string. It was not allowed in previous version.

## v1.0.0 ##

Initial tag. Library is stable enough for all features mentioned in README.md.