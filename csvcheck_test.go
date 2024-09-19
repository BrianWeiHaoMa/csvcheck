package csvcheck_test

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/BrianWeiHaoMa/csvcheck"

	"github.com/stretchr/testify/assert"
)

func Get2DArrayFromCsvString(csvString string) [][]csvcheck.StringHashable {
	reader := csv.NewReader(strings.NewReader(csvString))
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	res := make([][]csvcheck.StringHashable, len(records))
	for i, row := range records {
		res[i] = make([]csvcheck.StringHashable, len(row))
		for j, cell := range row {
			res[i][j] = csvcheck.BasicStringHashable(cell)
		}
	}
	return res
}

func getEmpty2DArray() [][]csvcheck.StringHashable {
	return [][]csvcheck.StringHashable{}
}

func getImproperCsvArrayDifferingRowLengths() [][]csvcheck.StringHashable {
	arr := [][]string{
		{"a", "b", "c"},
		{"1", "2", "3", "5"},
		{"4", "5", "6"},
		{"7", "8", "9", "2"},
	}
	return csvcheck.Get2DArrayFrom2DArray(arr)
}

func getImproperCsvArrayDifferingRepeatedColumnNames() [][]csvcheck.StringHashable {
	s := `
a,b,c,c
1,2,3,5
4,5,6,3
7,8,9,2
`
	return Get2DArrayFromCsvString(s)
}

func getCsvArray1() [][]csvcheck.StringHashable {
	s := `
a,b,c
1,2,3
4,5,6
7,8,9
`
	return Get2DArrayFromCsvString(s)
}

func getCsvArray2() [][]csvcheck.StringHashable {
	s := `
a,b,c
10,10,10
4,5,6
4,5,6
4,5,6
1,2,3
7,8,9
`
	return Get2DArrayFromCsvString(s)
}

func getCsvArray3() [][]csvcheck.StringHashable {
	s := `
a,b,c,d
1,2,3,a
4,5,6,b
7,8,9,c
`
	return Get2DArrayFromCsvString(s)
}

func generateRandom2DArray(cols []string, numCols, numRows, randomLimit int) [][]csvcheck.StringHashable {
	if cols != nil && numCols > 0 {
		panic("generateRandom2DArray: numCols and cols cannot both be specified")
	}

	if numCols > 0 {
		cols = make([]string, numCols)
		for i := range cols {
			cols[i] = fmt.Sprintf("%d", i)
		}
	}

	arr := make([][]csvcheck.StringHashable, numRows)
	arr[0] = make([]csvcheck.StringHashable, len(cols))
	copy(arr[0], csvcheck.GetRowFromRow(cols))
	for i := 1; i < numRows; i++ {
		row := make([]csvcheck.StringHashable, len(cols))
		for j := range cols {
			s := fmt.Sprintf("%d", rand.Intn(randomLimit))
			row[j] = csvcheck.BasicStringHashable(s)
		}
		arr[i] = row
	}
	return arr
}

func TestCheckForProperCsvArray(t *testing.T) {
	arr1 := getCsvArray1()
	err1 := csvcheck.CheckForProperCsvArray(arr1)

	arr2 := getImproperCsvArrayDifferingRowLengths()
	err2 := csvcheck.CheckForProperCsvArray(arr2)

	arr3 := getImproperCsvArrayDifferingRepeatedColumnNames()
	err3 := csvcheck.CheckForProperCsvArray(arr3)

	arr4 := getEmpty2DArray()
	err4 := csvcheck.CheckForProperCsvArray(arr4)

	assert.Nil(t, err1)
	assert.NotNil(t, err2)
	assert.NotNil(t, err3)
	assert.NotNil(t, err4)
}

func TestGetCommonIndicesMatchOneEmpty(t *testing.T) {
	empty := getEmpty2DArray()
	arr := getCsvArray1()
	commonIndices1, commonIndices2, err := csvcheck.GetCommonIndices(empty, arr, csvcheck.MethodMatch, false)

	assert.Nil(t, err)
	assert.Equal(t, []int{}, commonIndices1)
	assert.Equal(t, []int{}, commonIndices2)
}

func TestGetCommonIndicesMatchSomeDuplicateRows(t *testing.T) {
	arr1 := getCsvArray2()
	arr2 := Get2DArrayFromCsvString(`
a,b,c
10,10,10
4,5,6
4,5,6
1,2,3
7,8,9
7,8,9
`)
	commonIndices1, commonIndices2, err := csvcheck.GetCommonIndices(arr1, arr2, csvcheck.MethodMatch, true)

	assert.Nil(t, err)
	assert.Equal(t, []int{0, 1, 2, 3, 5, 6}, commonIndices1)
	assert.Equal(t, []int{0, 1, 2, 3, 4, 5}, commonIndices2)
}

func TestGetCommonIndicesDirectOneEmpty(t *testing.T) {
	arr1 := getEmpty2DArray()
	arr2 := getCsvArray1()
	commonIndices1, commonIndices2, err := csvcheck.GetCommonIndices(arr1, arr2, csvcheck.MethodDirect, true)

	assert.Nil(t, err)
	assert.Equal(t, []int{}, commonIndices1)
	assert.Equal(t, []int{}, commonIndices2)
}

func TestGetCommonIndicesDirectSomeRowsExactlyTheSame(t *testing.T) {
	arr1 := getCsvArray2()
	arr2 := Get2DArrayFromCsvString(`
a,b,c
10,10,10
0,0,0
4,5,6
1,2,3
1,2,3
0,0,0
`)
	commonIndices1, commonIndices2, err := csvcheck.GetCommonIndices(arr1, arr2, csvcheck.MethodDirect, true)

	assert.Nil(t, err)
	assert.Equal(t, []int{0, 1, 3, 5}, commonIndices1)
	assert.Equal(t, []int{0, 1, 3, 5}, commonIndices2)
}

func TestGetCommonIndicesSetOneEmpty(t *testing.T) {
	empty := getEmpty2DArray()
	arr := getCsvArray1()
	commonIndices1, commonIndices2, err := csvcheck.GetCommonIndices(empty, arr, csvcheck.MethodSet, true)

	assert.Nil(t, err)
	assert.Equal(t, []int{}, commonIndices1)
	assert.Equal(t, []int{}, commonIndices2)
}

func TestGetCommonIndicesSetSomeCommonRowsButNotNecessarilySameNumberOf(t *testing.T) {
	arr1 := getCsvArray2()
	arr2 := Get2DArrayFromCsvString(`
4,5,6
4,5,6
10,10,10
11,11,11
`)
	commonIndices1, commonIndices2, err := csvcheck.GetCommonIndices(arr1, arr2, csvcheck.MethodSet, true)

	assert.Nil(t, err)
	assert.Equal(t, []int{1, 2, 3, 4}, commonIndices1)
	assert.Equal(t, []int{0, 1, 2}, commonIndices2)
}

func TestGetDifferentIndicesMatchOneEmpty(t *testing.T) {
	empty := getEmpty2DArray()
	arr := getCsvArray1()
	differentIndices1, differentIndices2, err := csvcheck.GetDifferentIndices(empty, arr, csvcheck.MethodMatch, true)

	assert.Nil(t, err)
	assert.Equal(t, []int{}, differentIndices1)
	assert.Equal(t, []int{0, 1, 2, 3}, differentIndices2)
}

func TestGetDifferentIndicesMatchSomeDuplicateRows(t *testing.T) {
	arr1 := getCsvArray2()
	arr2 := Get2DArrayFromCsvString(`
a,b,c
10,10,10
4,5,6
4,5,6
1,2,3
7,8,9
7,8,9
`)
	differentIndices1, differentIndices2, err := csvcheck.GetDifferentIndices(arr1, arr2, csvcheck.MethodMatch, true)

	assert.Nil(t, err)
	assert.Equal(t, []int{4}, differentIndices1)
	assert.Equal(t, []int{6}, differentIndices2)
}

func TestGetDifferentIndicesDirectOneEmpty(t *testing.T) {
	empty := getEmpty2DArray()
	arr := getCsvArray1()
	differentIndices1, differentIndices2, err := csvcheck.GetDifferentIndices(empty, arr, csvcheck.MethodDirect, true)

	assert.Nil(t, err)
	assert.Equal(t, []int{}, differentIndices1)
	assert.Equal(t, []int{0, 1, 2, 3}, differentIndices2)
}

func TestGetDifferentIndicesDirectSomeRowsExactlyTheSame(t *testing.T) {
	arr1 := getCsvArray2()
	arr2 := Get2DArrayFromCsvString(`
a,b,c
10,10,10
0,0,0
4,5,6
1,2,3
1,2,3
0,0,0
`)
	differentIndices1, differentIndices2, err := csvcheck.GetDifferentIndices(arr1, arr2, csvcheck.MethodDirect, true)

	assert.Nil(t, err)
	assert.Equal(t, []int{2, 4, 6}, differentIndices1)
	assert.Equal(t, []int{2, 4, 6}, differentIndices2)
}

func TestGetDifferentIndicesSetOneEmpty(t *testing.T) {
	empty := getEmpty2DArray()
	arr := getCsvArray1()
	differentIndices1, differentIndices2, err := csvcheck.GetDifferentIndices(empty, arr, csvcheck.MethodSet, true)

	assert.Nil(t, err)
	assert.Equal(t, []int{}, differentIndices1)
	assert.Equal(t, []int{0, 1, 2, 3}, differentIndices2)
}

func TestGetDifferentIndicesSetSomeCommonRowsButNotNecessarilySameNumberOf(t *testing.T) {
	arr1 := getCsvArray2()
	arr2 := Get2DArrayFromCsvString(`
4,5,6
4,5,6
10,10,10
11,11,11
`)
	differentIndices1, differentIndices2, err := csvcheck.GetDifferentIndices(arr1, arr2, csvcheck.MethodSet, true)

	assert.Nil(t, err)
	assert.Equal(t, []int{0, 5, 6}, differentIndices1)
	assert.Equal(t, []int{3}, differentIndices2)
}

func TestKeepColumns(t *testing.T) {
	arr := getCsvArray1()

	res, err := csvcheck.KeepColumns(arr, csvcheck.GetRowFromRow([]string{"c"}))

	expected := Get2DArrayFromCsvString(`
c
3
6
9
`)
	assert.Nil(t, err)
	assert.Equal(t, expected, res)
}

func TestKeepColumnsManyMoreColumnInputValuesThanArray(t *testing.T) {
	arr := getCsvArray1()

	res, err := csvcheck.KeepColumns(arr, csvcheck.GetRowFromRow([]string{"c", "c", "s", "p", "q", "a"}))

	expected := Get2DArrayFromCsvString(`
a,c
1,3
4,6
7,9
`)
	assert.Nil(t, err)
	assert.Equal(t, expected, res)
}

func TestIgnoreColumns(t *testing.T) {
	arr := getCsvArray1()

	res, err := csvcheck.IgnoreColumns(arr, csvcheck.GetRowFromRow([]string{"c"}))

	expected := Get2DArrayFromCsvString(`
a,b
1,2
4,5
7,8
`)
	assert.Nil(t, err)
	assert.Equal(t, expected, res)
}

func TestIgnoreColumnsManyMoreColumnInputValuesThanArray(t *testing.T) {
	arr := getCsvArray1()

	res, err := csvcheck.IgnoreColumns(arr, csvcheck.GetRowFromRow([]string{"a", "a", "c", "s", "p", "q", "c", "a", "a"}))

	expected := Get2DArrayFromCsvString(`
b
2
5
8
`)
	assert.Nil(t, err)
	assert.Equal(t, expected, res)
}

func TestKeepRows(t *testing.T) {
	arr := getCsvArray2()

	res, err := csvcheck.KeepRows(arr, []int{0, 1, 3, 5})

	expected := Get2DArrayFromCsvString(`
a,b,c
10,10,10
4,5,6
1,2,3
`)
	assert.Nil(t, err)
	assert.Equal(t, expected, res)
}

func TestKeepRowsKeepNoRows(t *testing.T) {
	arr := getCsvArray2()

	res, err := csvcheck.KeepRows(arr, []int{})

	assert.Nil(t, err)
	assert.Equal(t, [][]csvcheck.StringHashable{}, res)
}

func TestKeepRowsDifferentIndiceOrdersAndRepeatedValues(t *testing.T) {
	arr := getCsvArray2()

	res, err := csvcheck.KeepRows(arr, []int{3, 5, 0, 0, 0, 1, 3, 5})

	expected := Get2DArrayFromCsvString(`
a,b,c
10,10,10
4,5,6
1,2,3
`)
	assert.Nil(t, err)
	assert.Equal(t, expected, res)
}

func TestIgnoreRows(t *testing.T) {
	arr := getCsvArray2()

	res, err := csvcheck.IgnoreRows(arr, []int{0, 1, 3, 5})

	expected := Get2DArrayFromCsvString(`
4,5,6
4,5,6
7,8,9
`)
	assert.Nil(t, err)
	assert.Equal(t, expected, res)
}

func TestIgnoreRowsIgnoreNoRows(t *testing.T) {
	arr := getCsvArray2()

	res, err := csvcheck.IgnoreRows(arr, []int{})

	assert.Nil(t, err)
	assert.Equal(t, getCsvArray2(), res)
}

func TestIgnoreRowsDifferentIndiceOrdersAndRepeatedValues(t *testing.T) {
	arr := getCsvArray2()

	res, err := csvcheck.IgnoreRows(arr, []int{3, 5, 0, 0, 0, 1, 3, 5})

	expected := Get2DArrayFromCsvString(`
4,5,6
4,5,6
7,8,9
`)
	assert.Nil(t, err)
	assert.Equal(t, expected, res)
}

func TestGetCommonRowsErrorsOnImproperCsvArray(t *testing.T) {
	arrs := [][][]csvcheck.StringHashable{
		getEmpty2DArray(),
		getImproperCsvArrayDifferingRepeatedColumnNames(),
		getImproperCsvArrayDifferingRowLengths(),
	}
	methods := []int{csvcheck.MethodMatch, csvcheck.MethodSet, csvcheck.MethodDirect}

	for _, arr1 := range arrs {
		for _, method := range methods {
			options := csvcheck.Options{
				Method:        method,
				UseColumns:    nil,
				IgnoreColumns: nil,
			}
			_, _, _, _, err := csvcheck.GetCommonRows(arr1, arr1, options)
			assert.NotNil(t, err)
		}
	}
}

func TestGetCommonRowsMatchUseAllColumns(t *testing.T) {
	arr1 := getCsvArray1()
	arr2 := getCsvArray2()

	options := csvcheck.Options{
		Method:        csvcheck.MethodMatch,
		UseColumns:    nil,
		IgnoreColumns: nil,
		SortIndices:   true,
	}

	res1, res2, indices1, indices2, err := csvcheck.GetCommonRows(arr1, arr2, options)

	expected1 := Get2DArrayFromCsvString(`
a,b,c
1,2,3
4,5,6
7,8,9
`)
	expectedIndices1 := []int{0, 1, 2, 3}

	expected2 := Get2DArrayFromCsvString(`
a,b,c
4,5,6
1,2,3
7,8,9
`)
	expectedIndices2 := []int{0, 2, 5, 6}

	assert.Nil(t, err)
	assert.Equal(t, expected1, res1)
	assert.Equal(t, expectedIndices1, indices1)
	assert.Equal(t, expected2, res2)
	assert.Equal(t, expectedIndices2, indices2)
}

func TestGetCommonRowsMatchUseNoColumns(t *testing.T) {
	arr1 := getCsvArray1()
	arr2 := getCsvArray2()

	options := csvcheck.Options{
		Method:        csvcheck.MethodMatch,
		UseColumns:    csvcheck.GetRowFromRow([]string{}),
		IgnoreColumns: nil,
	}

	_, _, _, _, err := csvcheck.GetCommonRows(arr1, arr2, options)

	assert.NotNil(t, err)
}

func TestGetCommonRowsMatchUseColumnsThatExistInOneButNotTheOther(t *testing.T) {
	arr1 := getCsvArray2()
	arr2 := getCsvArray3()

	options := csvcheck.Options{
		Method:        csvcheck.MethodMatch,
		UseColumns:    csvcheck.GetRowFromRow([]string{"d"}),
		IgnoreColumns: nil,
	}

	_, _, _, _, err := csvcheck.GetCommonRows(arr1, arr2, options)

	assert.NotNil(t, err)
}

func TestGetCommonRowsMatchUseColumnsWithSomeThatExistInOneButNotTheOther(t *testing.T) {
	arr1 := getCsvArray2()
	arr2 := getCsvArray3()

	options := csvcheck.Options{
		Method:        csvcheck.MethodMatch,
		UseColumns:    csvcheck.GetRowFromRow([]string{"a", "d"}),
		IgnoreColumns: nil,
	}

	_, _, _, _, err := csvcheck.GetCommonRows(arr1, arr2, options)

	assert.NotNil(t, err)
}

func TestGetCommonRowsMatchUseColumnsThatAreASubsetOfBothWithColumnsUnaligned(t *testing.T) {
	arr1 := getCsvArray3()
	arr2 := Get2DArrayFromCsvString(`
d,b,q,c
d,b,q,c
a,b,q,3
c,b,q,9
`)

	options := csvcheck.Options{
		Method:        csvcheck.MethodMatch,
		UseColumns:    csvcheck.GetRowFromRow([]string{"c", "d"}),
		IgnoreColumns: nil,
		SortIndices:   true,
	}

	res1, res2, indices1, indices2, err := csvcheck.GetCommonRows(arr1, arr2, options)

	expected1 := Get2DArrayFromCsvString(`
a,b,c,d
1,2,3,a
7,8,9,c
`)
	expectedIndices1 := []int{0, 1, 3}

	expected2 := Get2DArrayFromCsvString(`
d,b,q,c
a,b,q,3
c,b,q,9
`)
	expectedIndices2 := []int{0, 2, 3}

	assert.Nil(t, err)
	assert.Equal(t, expected1, res1)
	assert.Equal(t, expectedIndices1, indices1)
	assert.Equal(t, expected2, res2)
	assert.Equal(t, expectedIndices2, indices2)
}

func TestGetCommonRowsSetUseAllColumns(t *testing.T) {
	arr1 := getCsvArray1()
	arr2 := getCsvArray2()

	options := csvcheck.Options{
		Method:        csvcheck.MethodSet,
		UseColumns:    csvcheck.GetRowFromRow([]string{"a", "b", "c"}),
		IgnoreColumns: nil,
		SortIndices:   true,
	}

	res1, res2, indices1, indices2, err := csvcheck.GetCommonRows(arr1, arr2, options)

	expected1 := Get2DArrayFromCsvString(`
a,b,c
1,2,3
4,5,6
7,8,9
`)
	expectedIndices1 := []int{0, 1, 2, 3}

	expected2 := Get2DArrayFromCsvString(`
a,b,c
4,5,6
4,5,6
4,5,6
1,2,3
7,8,9
`)
	expectedIndices2 := []int{0, 2, 3, 4, 5, 6}

	assert.Nil(t, err)
	assert.Equal(t, expected1, res1)
	assert.Equal(t, expectedIndices1, indices1)
	assert.Equal(t, expected2, res2)
	assert.Equal(t, expectedIndices2, indices2)
}

func TestGetCommonRowsSetUseNoColumns(t *testing.T) {
	arr1 := getCsvArray1()
	arr2 := getCsvArray2()

	options := csvcheck.Options{
		Method:        csvcheck.MethodSet,
		UseColumns:    csvcheck.GetRowFromRow([]string{}),
		IgnoreColumns: nil,
	}

	_, _, _, _, err := csvcheck.GetCommonRows(arr1, arr2, options)

	assert.NotNil(t, err)
}

func TestGetCommonRowsSetUseColumnsThatExistInOneButNotTheOther(t *testing.T) {
	arr1 := getCsvArray2()
	arr2 := getCsvArray3()

	options := csvcheck.Options{
		Method:        csvcheck.MethodSet,
		UseColumns:    csvcheck.GetRowFromRow([]string{"d"}),
		IgnoreColumns: nil,
	}

	_, _, _, _, err := csvcheck.GetCommonRows(arr1, arr2, options)

	assert.NotNil(t, err)
}

func TestGetCommonRowsSetUseColumnsWithSomeThatExistInOneButNotTheOther(t *testing.T) {
	arr1 := getCsvArray2()
	arr2 := getCsvArray3()

	options := csvcheck.Options{
		Method:        csvcheck.MethodSet,
		UseColumns:    csvcheck.GetRowFromRow([]string{"a", "d"}),
		IgnoreColumns: nil,
	}

	_, _, _, _, err := csvcheck.GetCommonRows(arr1, arr2, options)

	assert.NotNil(t, err)
}

func TestGetCommonRowsSetUseColumnsThatAreASubsetOfBothWithColumnsUnaligned(t *testing.T) {
	arr1 := getCsvArray3()
	arr2 := Get2DArrayFromCsvString(`
d,b,q,c
d,b,q,c
a,b,q,3
c,b,q,9
c,b,q,9
`)

	options := csvcheck.Options{
		Method:        csvcheck.MethodSet,
		UseColumns:    csvcheck.GetRowFromRow([]string{"c", "d"}),
		IgnoreColumns: nil,
		SortIndices:   true,
	}

	res1, res2, indices1, indices2, err := csvcheck.GetCommonRows(arr1, arr2, options)

	expected1 := Get2DArrayFromCsvString(`
a,b,c,d
1,2,3,a
7,8,9,c
`)
	expectedIndices1 := []int{0, 1, 3}

	expected2 := Get2DArrayFromCsvString(`
d,b,q,c
a,b,q,3
c,b,q,9
c,b,q,9
`)
	expectedIndices2 := []int{0, 2, 3, 4}

	assert.Nil(t, err)
	assert.Equal(t, expected1, res1)
	assert.Equal(t, expectedIndices1, indices1)
	assert.Equal(t, expected2, res2)
	assert.Equal(t, expectedIndices2, indices2)
}

func TestGetCommonRowsDirectUseAllColumns(t *testing.T) {
	arr1 := getCsvArray1()
	arr2 := getCsvArray2()

	options := csvcheck.Options{
		Method:        csvcheck.MethodDirect,
		UseColumns:    csvcheck.GetRowFromRow([]string{"a", "b", "c"}),
		IgnoreColumns: nil,
		SortIndices:   true,
	}

	res1, res2, indices1, indices2, err := csvcheck.GetCommonRows(arr1, arr2, options)

	expected1 := Get2DArrayFromCsvString(`
a,b,c
4,5,6
`)
	expectedIndices1 := []int{0, 2}

	expected2 := Get2DArrayFromCsvString(`
a,b,c
4,5,6
`)
	expectedIndices2 := []int{0, 2}

	assert.Nil(t, err)
	assert.Equal(t, expected1, res1)
	assert.Equal(t, expectedIndices1, indices1)
	assert.Equal(t, expected2, res2)
	assert.Equal(t, expectedIndices2, indices2)
}

func TestGetCommonRowsDirectUseNoColumns(t *testing.T) {
	arr1 := getCsvArray1()
	arr2 := getCsvArray2()

	options := csvcheck.Options{
		Method:        csvcheck.MethodDirect,
		UseColumns:    csvcheck.GetRowFromRow([]string{}),
		IgnoreColumns: nil,
	}

	_, _, _, _, err := csvcheck.GetCommonRows(arr1, arr2, options)

	assert.NotNil(t, err)
}

func TestGetCommonRowsDirectUseColumnsThatExistInOneButNotTheOther(t *testing.T) {
	arr1 := getCsvArray2()
	arr2 := getCsvArray3()

	options := csvcheck.Options{
		Method:        csvcheck.MethodDirect,
		UseColumns:    csvcheck.GetRowFromRow([]string{"d"}),
		IgnoreColumns: nil,
	}

	_, _, _, _, err := csvcheck.GetCommonRows(arr1, arr2, options)

	assert.NotNil(t, err)
}

func TestGetCommonRowsDirectUseColumnsWithSomeThatExistInOneButNotTheOther(t *testing.T) {
	arr1 := getCsvArray2()
	arr2 := getCsvArray3()

	options := csvcheck.Options{
		Method:        csvcheck.MethodDirect,
		UseColumns:    csvcheck.GetRowFromRow([]string{"a", "d"}),
		IgnoreColumns: nil,
	}

	_, _, _, _, err := csvcheck.GetCommonRows(arr1, arr2, options)

	assert.NotNil(t, err)
}

func TestGetCommonRowsDirectUseColumnsThatAreASubsetOfBothWithColumnsUnaligned(t *testing.T) {
	arr1 := getCsvArray3()
	arr2 := Get2DArrayFromCsvString(`
d,b,q,c
d,b,q,c
a,b,q,3
c,b,q,9
`)

	options := csvcheck.Options{
		Method:        csvcheck.MethodDirect,
		UseColumns:    csvcheck.GetRowFromRow([]string{"c", "d"}),
		IgnoreColumns: nil,
		SortIndices:   true,
	}

	res1, res2, indices1, indices2, err := csvcheck.GetCommonRows(arr1, arr2, options)

	expected1 := Get2DArrayFromCsvString(`
a,b,c,d
7,8,9,c
`)
	expectedIndices1 := []int{0, 3}

	expected2 := Get2DArrayFromCsvString(`
d,b,q,c
c,b,q,9
`)
	expectedIndices2 := []int{0, 3}

	assert.Nil(t, err)
	assert.Equal(t, expected1, res1)
	assert.Equal(t, expectedIndices1, indices1)
	assert.Equal(t, expected2, res2)
	assert.Equal(t, expectedIndices2, indices2)
}

func TestGetDifferentRowsErrorsOnImproperCsvArray(t *testing.T) {
	arrs := [][][]csvcheck.StringHashable{
		getEmpty2DArray(),
		getImproperCsvArrayDifferingRepeatedColumnNames(),
		getImproperCsvArrayDifferingRowLengths(),
	}
	methods := []int{csvcheck.MethodMatch, csvcheck.MethodSet, csvcheck.MethodDirect}

	for _, arr1 := range arrs {
		for _, method := range methods {
			options := csvcheck.Options{
				Method:        method,
				UseColumns:    nil,
				IgnoreColumns: nil,
			}
			_, _, _, _, err := csvcheck.GetDifferentRows(arr1, arr1, options)
			assert.NotNil(t, err)
		}
	}
}

func TestGetDifferentRowsMatchUseAllColumns(t *testing.T) {
	arr1 := getCsvArray1()
	arr2 := getCsvArray2()

	options := csvcheck.Options{
		Method:        csvcheck.MethodMatch,
		UseColumns:    nil,
		IgnoreColumns: nil,
		SortIndices:   true,
	}

	res1, res2, indices1, indices2, err := csvcheck.GetDifferentRows(arr1, arr2, options)

	expected1 := Get2DArrayFromCsvString(`
a,b,c
`)
	expectedIndices1 := []int{0}

	expected2 := Get2DArrayFromCsvString(`
a,b,c
10,10,10
4,5,6
4,5,6
`)
	expectedIndices2 := []int{0, 1, 3, 4}

	assert.Nil(t, err)
	assert.Equal(t, expected1, res1)
	assert.Equal(t, expectedIndices1, indices1)
	assert.Equal(t, expected2, res2)
	assert.Equal(t, expectedIndices2, indices2)
}

func TestGetDifferentRowsMatchUseNoColumns(t *testing.T) {
	arr1 := getCsvArray1()
	arr2 := getCsvArray2()

	options := csvcheck.Options{
		Method:        csvcheck.MethodMatch,
		UseColumns:    csvcheck.GetRowFromRow([]string{}),
		IgnoreColumns: nil,
	}

	_, _, _, _, err := csvcheck.GetDifferentRows(arr1, arr2, options)

	assert.NotNil(t, err)
}

func TestGetDifferentRowsMatchUseColumnsThatExistInOneButNotTheOther(t *testing.T) {
	arr1 := getCsvArray2()
	arr2 := getCsvArray3()

	options := csvcheck.Options{
		Method:        csvcheck.MethodMatch,
		UseColumns:    csvcheck.GetRowFromRow([]string{"d"}),
		IgnoreColumns: nil,
	}

	_, _, _, _, err := csvcheck.GetDifferentRows(arr1, arr2, options)

	assert.NotNil(t, err)
}

func TestGetDifferentRowsMatchUseColumnsWithSomeThatExistInOneButNotTheOther(t *testing.T) {
	arr1 := getCsvArray2()
	arr2 := getCsvArray3()

	options := csvcheck.Options{
		Method:        csvcheck.MethodMatch,
		UseColumns:    csvcheck.GetRowFromRow([]string{"a", "d"}),
		IgnoreColumns: nil,
	}

	_, _, _, _, err := csvcheck.GetDifferentRows(arr1, arr2, options)

	assert.NotNil(t, err)
}

func TestGetDifferentRowsMatchUseColumnsThatAreASubsetOfBothWithColumnsUnaligned(t *testing.T) {
	arr1 := getCsvArray3()
	arr2 := Get2DArrayFromCsvString(`
d,b,q,c
d,b,q,c
a,b,q,3
c,b,q,9
c,b,q,9
`)

	options := csvcheck.Options{
		Method:        csvcheck.MethodMatch,
		UseColumns:    csvcheck.GetRowFromRow([]string{"c", "d"}),
		IgnoreColumns: nil,
		SortIndices:   true,
	}

	res1, res2, indices1, indices2, err := csvcheck.GetDifferentRows(arr1, arr2, options)

	expected1 := Get2DArrayFromCsvString(`
a,b,c,d
4,5,6,b
`)
	expectedIndices1 := []int{0, 2}

	expected2 := Get2DArrayFromCsvString(`
d,b,q,c
d,b,q,c
c,b,q,9
`)
	expectedIndices2 := []int{0, 1, 4}

	assert.Nil(t, err)
	assert.Equal(t, expected1, res1)
	assert.Equal(t, expectedIndices1, indices1)
	assert.Equal(t, expected2, res2)
	assert.Equal(t, expectedIndices2, indices2)
}

func TestGetDifferentRowsSetUseAllColumns(t *testing.T) {
	arr1 := getCsvArray1()
	arr2 := getCsvArray2()

	options := csvcheck.Options{
		Method:        csvcheck.MethodSet,
		UseColumns:    csvcheck.GetRowFromRow([]string{"a", "b", "c"}),
		IgnoreColumns: nil,
		SortIndices:   true,
	}

	res1, res2, indices1, indices2, err := csvcheck.GetDifferentRows(arr1, arr2, options)

	expected1 := Get2DArrayFromCsvString(`
a,b,c
`)
	expectedIndices1 := []int{0}

	expected2 := Get2DArrayFromCsvString(`
a,b,c
10,10,10
`)
	expectedIndices2 := []int{0, 1}

	assert.Nil(t, err)
	assert.Equal(t, expected1, res1)
	assert.Equal(t, expectedIndices1, indices1)
	assert.Equal(t, expected2, res2)
	assert.Equal(t, expectedIndices2, indices2)
}

func TestGetDifferentRowsSetUseNoColumns(t *testing.T) {
	arr1 := getCsvArray1()
	arr2 := getCsvArray2()

	options := csvcheck.Options{
		Method:        csvcheck.MethodSet,
		UseColumns:    csvcheck.GetRowFromRow([]string{}),
		IgnoreColumns: nil,
	}

	_, _, _, _, err := csvcheck.GetDifferentRows(arr1, arr2, options)

	assert.NotNil(t, err)
}

func TestGetDifferentRowsSetUseColumnsThatExistInOneButNotTheOther(t *testing.T) {
	arr1 := getCsvArray2()
	arr2 := getCsvArray3()

	options := csvcheck.Options{
		Method:        csvcheck.MethodSet,
		UseColumns:    csvcheck.GetRowFromRow([]string{"d"}),
		IgnoreColumns: nil,
	}

	_, _, _, _, err := csvcheck.GetDifferentRows(arr1, arr2, options)

	assert.NotNil(t, err)
}

func TestGetDifferentRowsSetUseColumnsWithSomeThatExistInOneButNotTheOther(t *testing.T) {
	arr1 := getCsvArray2()
	arr2 := getCsvArray3()

	options := csvcheck.Options{
		Method:        csvcheck.MethodSet,
		UseColumns:    csvcheck.GetRowFromRow([]string{"a", "d"}),
		IgnoreColumns: nil,
	}

	_, _, _, _, err := csvcheck.GetDifferentRows(arr1, arr2, options)

	assert.NotNil(t, err)
}

func TestGetDifferentRowsSetUseColumnsThatAreASubsetOfBothWithColumnsUnaligned(t *testing.T) {
	arr1 := getCsvArray3()
	arr2 := Get2DArrayFromCsvString(`
d,b,q,c
d,b,q,c
a,b,q,3
c,b,q,9
c,b,q,9
`)

	options := csvcheck.Options{
		Method:        csvcheck.MethodSet,
		UseColumns:    csvcheck.GetRowFromRow([]string{"c", "d"}),
		IgnoreColumns: nil,
		SortIndices:   true,
	}

	res1, res2, indices1, indices2, err := csvcheck.GetDifferentRows(arr1, arr2, options)

	expected1 := Get2DArrayFromCsvString(`
a,b,c,d
4,5,6,b
`)
	expectedIndices1 := []int{0, 2}

	expected2 := Get2DArrayFromCsvString(`
d,b,q,c
d,b,q,c
`)
	expectedIndices2 := []int{0, 1}

	assert.Nil(t, err)
	assert.Equal(t, expected1, res1)
	assert.Equal(t, expectedIndices1, indices1)
	assert.Equal(t, expected2, res2)
	assert.Equal(t, expectedIndices2, indices2)
}

func TestGetDifferentRowsDirectUseAllColumns(t *testing.T) {
	arr1 := getCsvArray1()
	arr2 := getCsvArray2()

	options := csvcheck.Options{
		Method:        csvcheck.MethodDirect,
		UseColumns:    csvcheck.GetRowFromRow([]string{"a", "b", "c"}),
		IgnoreColumns: nil,
		SortIndices:   true,
	}

	res1, res2, indices1, indices2, err := csvcheck.GetDifferentRows(arr1, arr2, options)

	expected1 := Get2DArrayFromCsvString(`
a,b,c
1,2,3
7,8,9
`)
	expectedIndices1 := []int{0, 1, 3}

	expected2 := Get2DArrayFromCsvString(`
a,b,c
10,10,10
4,5,6
4,5,6
1,2,3
7,8,9
`)
	expectedIndices2 := []int{0, 1, 3, 4, 5, 6}

	assert.Nil(t, err)
	assert.Equal(t, expected1, res1)
	assert.Equal(t, expectedIndices1, indices1)
	assert.Equal(t, expected2, res2)
	assert.Equal(t, expectedIndices2, indices2)
}

func TestGetDifferentRowsDirectUseNoColumns(t *testing.T) {
	arr1 := getCsvArray1()
	arr2 := getCsvArray2()

	options := csvcheck.Options{
		Method:        csvcheck.MethodDirect,
		UseColumns:    csvcheck.GetRowFromRow([]string{}),
		IgnoreColumns: nil,
	}

	_, _, _, _, err := csvcheck.GetDifferentRows(arr1, arr2, options)

	assert.NotNil(t, err)
}

func TestGetDifferentRowsDirectUseColumnsThatExistInOneButNotTheOther(t *testing.T) {
	arr1 := getCsvArray2()
	arr2 := getCsvArray3()

	options := csvcheck.Options{
		Method:        csvcheck.MethodDirect,
		UseColumns:    csvcheck.GetRowFromRow([]string{"d"}),
		IgnoreColumns: nil,
	}

	_, _, _, _, err := csvcheck.GetDifferentRows(arr1, arr2, options)

	assert.NotNil(t, err)
}

func TestGetDifferentRowsDirectUseColumnsWithSomeThatExistInOneButNotTheOther(t *testing.T) {
	arr1 := getCsvArray2()
	arr2 := getCsvArray3()

	options := csvcheck.Options{
		Method:        csvcheck.MethodSet,
		UseColumns:    csvcheck.GetRowFromRow([]string{"a", "d"}),
		IgnoreColumns: nil,
	}

	_, _, _, _, err := csvcheck.GetDifferentRows(arr1, arr2, options)

	assert.NotNil(t, err)
}

func TestRearrangeColumns(t *testing.T) {
	arr := Get2DArrayFromCsvString(`
a,b,c,d
1,2,3,4
5,6,7,8
9,10,11,12
`)
	columns := csvcheck.GetRowFromRow([]string{"c", "a", "d", "b"})

	res, err := csvcheck.RearrangeColumns(arr, columns)

	expected := Get2DArrayFromCsvString(`
c,a,d,b
3,1,4,2
7,5,8,6
11,9,12,10
`)

	assert.Nil(t, err)
	assert.Equal(t, expected, res)
}

func TestRearrangeColumnsNonExistantColumn(t *testing.T) {
	arr := Get2DArrayFromCsvString(`
a,b,c,d
1,2,3,4
5,6,7,8
9,10,11,12
`)
	columns := csvcheck.GetRowFromRow([]string{"l", "a", "d", "b"})

	_, err := csvcheck.RearrangeColumns(arr, columns)

	assert.NotNil(t, err)
}

func TestRearrangeColumnsRepeatedExtraColumns(t *testing.T) {
	arr := Get2DArrayFromCsvString(`
a,b,c,d
1,2,3,4
5,6,7,8
9,10,11,12
`)
	columns := csvcheck.GetRowFromRow([]string{"c", "a", "d", "b", "a"})

	_, err := csvcheck.RearrangeColumns(arr, columns)

	assert.NotNil(t, err)
}

func TestRearrangeColumnsRepeatedColumnsSameLength(t *testing.T) {
	arr := Get2DArrayFromCsvString(`
a,b,c,d
1,2,3,4
5,6,7,8
9,10,11,12
`)
	columns := csvcheck.GetRowFromRow([]string{"c", "a", "b", "a"})

	_, err := csvcheck.RearrangeColumns(arr, columns)

	assert.NotNil(t, err)
}

func TestRearrangeColumnsMissingColumn(t *testing.T) {
	arr := Get2DArrayFromCsvString(`
a,b,c,d
1,2,3,4
5,6,7,8
9,10,11,12
`)
	columns := csvcheck.GetRowFromRow([]string{"a", "b", "a"})

	_, err := csvcheck.RearrangeColumns(arr, columns)

	assert.NotNil(t, err)
}

func TestAutoAlignCsvArraysSomeCommonColumns(t *testing.T) {
	arr1 := Get2DArrayFromCsvString(`
a,b,c,d,e
1,2,3,4,5
6,7,8,9,10
`)
	arr2 := Get2DArrayFromCsvString(`
d,f,e,a
1,2,3,4
9,8,7,6
`)
	res1, res2, err := csvcheck.AutoAlignCsvArrays(arr1, arr2)

	expected1 := Get2DArrayFromCsvString(`
a,d,e,b,c
1,4,5,2,3
6,9,10,7,8
`)
	expected2 := Get2DArrayFromCsvString(`
a,d,e,f
4,1,3,2
6,9,7,8
`)
	assert.Nil(t, err)
	assert.Equal(t, expected1, res1)
	assert.Equal(t, expected2, res2)
}

func TestAutoAlignCsvArraysNoCommonColumns(t *testing.T) {
	arr1 := Get2DArrayFromCsvString(`
a,b,c,d,e
1,2,3,4,5
6,7,8,9,10
`)
	arr2 := Get2DArrayFromCsvString(`
z,x,t,y
1,2,3,4
9,8,7,6
`)
	res1, res2, err := csvcheck.AutoAlignCsvArrays(arr1, arr2)

	expected1 := Get2DArrayFromCsvString(`
a,b,c,d,e
1,2,3,4,5
6,7,8,9,10
`)
	expected2 := Get2DArrayFromCsvString(`
z,x,t,y
1,2,3,4
9,8,7,6
`)
	assert.Nil(t, err)
	assert.Equal(t, expected1, res1)
	assert.Equal(t, expected2, res2)
}

func TestGetCommonColumnsErrorsOnImproperCsvArray(t *testing.T) {
	arrs := [][][]csvcheck.StringHashable{
		getEmpty2DArray(),
		getImproperCsvArrayDifferingRepeatedColumnNames(),
		getImproperCsvArrayDifferingRowLengths(),
	}

	for _, arr := range arrs {
		_, err := csvcheck.GetCommonColumns(arr, arr)
		assert.NotNil(t, err)
	}
}

func TestGetCommonColumnsSomeCommonColumns(t *testing.T) {
	arr1 := Get2DArrayFromCsvString(`
a,b,c,d,e
1,2,3,4,5
6,7,8,9,10
`)
	arr2 := Get2DArrayFromCsvString(`
d,f,e,a
1,2,3,4
9,8,7,6
`)
	res, err := csvcheck.GetCommonColumns(arr1, arr2)

	expected := csvcheck.GetRowFromRow([]string{"a", "d", "e"})

	assert.Nil(t, err)
	assert.Equal(t, expected, res)
}

func TestGetCommonColumnsNoCommonColumns(t *testing.T) {
	arr1 := Get2DArrayFromCsvString(`
a,b,c,d,e
1,2,3,4,5
6,7,8,9,10
`)
	arr2 := Get2DArrayFromCsvString(`
z,x,t,y
1,2,3,4
9,8,7,6
`)
	res, err := csvcheck.GetCommonColumns(arr1, arr2)

	expected := csvcheck.GetRowFromRow([]string{})

	assert.Nil(t, err)
	assert.Equal(t, expected, res)
}

func TestPrettyFormatCsvArrayErrorsOnImproperArray(t *testing.T) {
	arrs := [][][]csvcheck.StringHashable{
		getEmpty2DArray(),
		getImproperCsvArrayDifferingRepeatedColumnNames(),
		getImproperCsvArrayDifferingRowLengths(),
	}

	for _, arr := range arrs {
		_, err := csvcheck.PrettyFormatCsvArray(arr, 3, -1)
		assert.NotNil(t, err)
	}
}

func TestPrettyFormatCsvArrayProperCsvArray(t *testing.T) {
	arr := Get2DArrayFromCsvString(`
aaaa,b,ccc
1,2,3
5,88888,7
`)

	res, err := csvcheck.PrettyFormatCsvArray(arr, 3, -1)
	expected := `
aaaa   b       ccc
1      2       3
5      88888   7
`[1:]
	assert.Nil(t, err)
	assert.Equal(t, expected, res)
}

func TestPrettyFormatCsvArrayProperCsvArrayMaxColLength(t *testing.T) {
	arr := Get2DArrayFromCsvString(`
aaaa,b,ccc
1,2,3
5,88888,7
`)

	res, err := csvcheck.PrettyFormatCsvArray(arr, 3, 2)
	expected := fmt.Sprintf(`
aa%s   b      cc%s
1      2      3
5      88%s   7
`[1:], csvcheck.TruncatedMark, csvcheck.TruncatedMark, csvcheck.TruncatedMark)
	assert.Nil(t, err)
	assert.Equal(t, expected, res)
}

func TestFormatCsvArrayErrorsOnImproperCsvArray(t *testing.T) {
	arrs := [][][]csvcheck.StringHashable{
		getEmpty2DArray(),
		getImproperCsvArrayDifferingRepeatedColumnNames(),
		getImproperCsvArrayDifferingRowLengths(),
	}

	for _, arr := range arrs {
		_, err := csvcheck.StringFormatCsvArray(arr)
		assert.NotNil(t, err)
	}
}

func TestFormatCsvArray(t *testing.T) {
	arr := Get2DArrayFromCsvString(`
aaaa,b,ccc
1,2,3
5,88888,7
`)

	res, err := csvcheck.StringFormatCsvArray(arr)
	expected := `
aaaa,b,ccc
1,2,3
5,88888,7
`[1:]
	assert.Nil(t, err)
	assert.Equal(t, expected, res)
}

func BenchmarkGetCommonRowsMatchVeryLittleCommonRows_5700x5700_5700x5700(b *testing.B) {
	arr1 := generateRandom2DArray(nil, 5700, 5700, 4000000)
	arr2 := generateRandom2DArray(nil, 5700, 5700, 4000000)

	options := csvcheck.Options{
		Method:        csvcheck.MethodMatch,
		UseColumns:    nil,
		IgnoreColumns: nil,
	}

	b.ResetTimer()

	csvcheck.GetCommonRows(arr1, arr2, options)
}

func BenchmarkGetCommonRowsMatchVeryLittleCommonRows_4000000x8_4000000x8(b *testing.B) {
	columns := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	arr1 := generateRandom2DArray(columns, -1, 4000000, 4000000)
	arr2 := generateRandom2DArray(columns, -1, 4000000, 4000000)

	options := csvcheck.Options{
		Method:        csvcheck.MethodMatch,
		UseColumns:    nil,
		IgnoreColumns: nil,
	}

	b.ResetTimer()

	csvcheck.GetCommonRows(arr1, arr2, options)
}

func BenchmarkGetCommonRowsMatchVeryLittleCommonRows_8x4000000_8x4000000(b *testing.B) {
	arr1 := generateRandom2DArray(nil, 4000000, 8, 4000000)
	arr2 := generateRandom2DArray(nil, 4000000, 8, 4000000)

	options := csvcheck.Options{
		Method:        csvcheck.MethodMatch,
		UseColumns:    nil,
		IgnoreColumns: nil,
	}

	b.ResetTimer()

	csvcheck.GetCommonRows(arr1, arr2, options)
}

func BenchmarkGetDifferentRowsMatchVeryLittleCommonRows_5700x5700_5700x5700(b *testing.B) {
	arr1 := generateRandom2DArray(nil, 5700, 5700, 4000000)
	arr2 := generateRandom2DArray(nil, 5700, 5700, 4000000)

	options := csvcheck.Options{
		Method:        csvcheck.MethodMatch,
		UseColumns:    nil,
		IgnoreColumns: nil,
	}

	b.ResetTimer()

	csvcheck.GetDifferentRows(arr1, arr2, options)
}

func BenchmarkGetDifferentRowsMatchVeryLittleCommonRows_4000000x8_4000000x8(b *testing.B) {
	columns := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	arr1 := generateRandom2DArray(columns, -1, 4000000, 4000000)
	arr2 := generateRandom2DArray(columns, -1, 4000000, 4000000)

	options := csvcheck.Options{
		Method:        csvcheck.MethodMatch,
		UseColumns:    nil,
		IgnoreColumns: nil,
	}

	b.ResetTimer()

	csvcheck.GetDifferentRows(arr1, arr2, options)
}

func BenchmarkGetDifferentRowsMatchVeryLittleCommonRows_8x4000000_8x4000000(b *testing.B) {
	arr1 := generateRandom2DArray(nil, 4000000, 8, 4000000)
	arr2 := generateRandom2DArray(nil, 4000000, 8, 4000000)

	options := csvcheck.Options{
		Method:        csvcheck.MethodMatch,
		UseColumns:    nil,
		IgnoreColumns: nil,
	}

	b.ResetTimer()

	csvcheck.GetDifferentRows(arr1, arr2, options)
}
