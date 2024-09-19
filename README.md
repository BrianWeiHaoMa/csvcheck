# csvcheck
A Go module for comparing the rows of different csv arrays
with a bundle of convenient functions to help.

## Installation
Install the latest version of csvcheck using
```
go get -u github.com/BrianWeiHaoMa/csvcheck@latest
```
To include csvcheck in your program
```
import "github.com/BrianWeiHaoMa/csvcheck"
```

## Example 1:
```
// First csv array.
arr1 := csvcheck.Get2DArrayFrom2DArray([][]string{
    {"a", "b", "c"},
    {"1", "2", "3"},
    {"4", "5", "6"},
    {"7", "8", "9"},
    {"1", "2", "3"},
    {"7", "8", "9"},
})

// Second csv array.
arr2 := csvcheck.Get2DArrayFrom2DArray([][]string{
    {"a", "b", "c"},
    {"x", "y", "z"},
    {"1", "2", "3"},
    {"4", "5", "6"},
    {"7", "8", "9"},
    {"10", "11", "12"},
})

// Returns the different rows, and their indices in the
// original arrays, respectively.
res1, res2, indices1, indices2, _ := csvcheck.GetCommonRows(
    arr1,
    arr2,
    csvcheck.Options{
        SortIndices: true,
        Method:      csvcheck.MethodSet,
        // Can choose one of MethodDirect, MethodSet, and MethodMatch.
    },
)

res1String, _ := csvcheck.StringFormatCsvArray(res1)
res2String, _ := csvcheck.PrettyFormatCsvArray(res2, 3, -1)

fmt.Println(indices1)
fmt.Println(res1String)

fmt.Println(indices2)
fmt.Println(res2String)
```
### Output 1:
```
[0 1 2 3]
a,b,c
1,2,3
4,5,6
7,8,9

[0 2 3 4]
a   b   c
1   2   3
4   5   6
7   8   9

```

## Example 2:
```
arr1 := csvcheck.Get2DArrayFrom2DArray([][]string{
    {"a", "b", "c"},
    {"1", "2", "3"},
    {"4", "5", "6"},
    {"7", "8", "9"},
    {"1", "2", "3"},
    {"7", "8", "9"},
})
arr2 := csvcheck.Get2DArrayFrom2DArray([][]string{
    {"a", "x", "y"},
    {"x", "y", "z"},
    {"1", "2", "3"},
    {"4", "5", "6"},
    {"7", "8", "9"},
    {"10", "11", "12"},
})

res1, res2, indices1, indices2, _ := csvcheck.GetDifferentRows(
    arr1,
    arr2,
    csvcheck.Options{
        SortIndices: true,
        Method:      csvcheck.MethodMatch,
        UseColumns: csvcheck.GetRowFromRow([]string{"a"}),
    },
)

res1String, _ := csvcheck.StringFormatCsvArray(res1)
res2String, _ := csvcheck.PrettyFormatCsvArray(res2, 3, -1)

fmt.Println(indices1)
fmt.Println(res1String)

fmt.Println(indices2)
fmt.Println(res2String)
```
### Output 2:
```
[0 4 5]
a,b,c
1,2,3
7,8,9

[0 1 5]
a    x    y
x    y    z
10   11   12

```

## Notes
### GetCommonRows
- MethodDirect: Compares each row of arr1 with the row in arr2 at the same index.
- MethodSet: Iff the row exists in arr1 and arr2 (ignoring indices), the row is kept.
- MethodMatch: Matches rows from arr1 to arr2 from the top down. Only rows that can be matched are kept.
### GetDifferentRows
- MethodDirect: Compares each row of arr1 with the row in arr2 at the same index.
- MethodSet: Iff the row doesn't exist in the other array keep it in the result for the current array.
- MethodMatch: All the rows not returned by GetCommonRows using MethodMatch, respectively.
