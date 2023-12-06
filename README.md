# Ascii runes mapper
Ascii runes mapper extends `golang.org/x/text/transform` extension to support mapping from a single rune to a string with multiple runes. 
The idea behind implementing such a mapper is to replace ASCII 256 extended characters that are part of [`unicode.Ll`](https://www.compart.com/en/unicode/category/Ll) and [`unicode.Lu`](https://www.compart.com/en/unicode/category/Lu) to remap to basic ASCII 128 characters.

## What's the use case?
We had a common issue that Norwegian characters (could be others - Vietnamese, Greece, etc.) are part of unicode.Ll or Lu but using `norm.NFD` the decomposition of character does not work for characters such as ["Œ", "Æ", "Ø", "ß"] and their lowercase variants. So you are not able to modify using the `transform` module the name `groß`/`Æd`/`Snøfall 2` as it is not part of the `norm.NFD` decomposition.

When you would like to search for previously mentioned titles using simple ASCII characters, you come across a problem that it does not match at all (it has to be an exact match of Unicode character). That's why this library exists. It can transform uncommon characters into their string representatives that are part of ASCII 128. These new variants could be part of the search index as well as old variants.

When you would like to find `gross` or `groß` it does not matter, they are both the same. As I have not found such a library in Go that would fit the `transform` extension module, only custom libraries.

## Rune mapper
The rune Map transformation function has prescription of `Map(mapping func(rune) string) Transformer` which means that single `rune` cloud be converted into single/multiple `runes`. The tests were inspired by module `golang.org/x/text/runes` which was never adopted into `Go` modules.

## Examples
```go
import mapper "github.com/Matovidlo/ascii-runes-map"

func main() {
    t := mapper.Map(mapper.Ascii256Toascii128)
    text, _, _ = transform.String(t, "Hellø")
    fmt.Println(text) // Hello
}
```


