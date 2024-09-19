package csvcheck

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cespare/xxhash"
)

// Supported comparison methods.
const (
	MethodMatch = iota
	MethodDirect
	MethodSet
)

// For truncated pretty formatted strings.
const TruncatedMark = ".."

// For allowing different types to be used in the
// csv arrays.
type StringHashable interface {
	StringHash() string // Returns true iff the two objects are equal.
}

// A basic implementation of the StringHashable interface
// for convenience.
type BasicStringHashable string

func (s BasicStringHashable) StringHash() string {
	return string(s)
}

func getStringKey(s StringHashable) uint64 {
	return xxhash.Sum64String(s.StringHash())
}

// For holding supported options.
type Options struct {
	Method        int
	UseColumns    []StringHashable
	IgnoreColumns []StringHashable
	SortIndices   bool
}

// Checks if the options are valid.
func (o *Options) CheckAttributes() error {
	if o.Method != MethodMatch && o.Method != MethodDirect && o.Method != MethodSet {
		return fmt.Errorf("unsupported method: %d", o.Method)
	}

	if o.UseColumns != nil && o.IgnoreColumns != nil {
		return fmt.Errorf("cannot use both UseColumns and IgnoreColumns together")
	}

	return nil
}

// Returns a string array from a StringHashable array.
func getStringsRow(row []StringHashable) []string {
	arr := make([]string, len(row))
	for i, v := range row {
		arr[i] = v.StringHash()
	}
	return arr
}

// Returns true iff the two rows are permutations of each other.
func rowsArePermutationsOfEachOther(row1, row2 []StringHashable) bool {
	if len(row1) != len(row2) {
		return false
	}

	cnt1 := make(map[uint64]int)
	for _, v := range row1 {
		cnt1[getStringKey(v)]++
	}

	for _, v := range row2 {
		stringKey := getStringKey(v)
		if _, exists := cnt1[stringKey]; !exists {
			return false
		}
		cnt1[stringKey]--
	}

	for _, v := range cnt1 {
		if v != 0 {
			return false
		}
	}

	return true
}

// A hash key for a row.
type rowKey struct {
	row     uint64
	lengths uint64
}

// Returns a hash key for a row.
func getRowKey(row []StringHashable) rowKey {
	rowHolder := make([]string, len(row))
	lengthsHolder := make([]string, len(row))
	for i, v := range row {
		s := v.StringHash()
		rowHolder[i] = s
		lengthsHolder[i] = fmt.Sprintf("%d", len(s))
	}
	rowS := strings.Join(rowHolder, "")
	lengthsS := strings.Join(lengthsHolder, ",")
	res := rowKey{
		row:     xxhash.Sum64String(rowS),
		lengths: xxhash.Sum64String(lengthsS),
	}
	return res
}

// Returns a mapping of rows to sorted lists of their indices in the input array.
func getRowsMapping(arr [][]StringHashable) map[rowKey][]int {
	mapping := make(map[rowKey][]int)
	for i, row := range arr {
		key := getRowKey(row)
		mapping[key] = append(mapping[key], i)
	}
	return mapping
}

// Returns the indices of rows common to both arrays
// using the match method.
func getCommonIndicesMatch(arr1, arr2 [][]StringHashable) ([]int, []int) {
	rowsMapping1 := getRowsMapping(arr1)
	rowsMapping2 := getRowsMapping(arr2)

	commonIndices1 := []int{}
	commonIndices2 := []int{}
	for key, indices1 := range rowsMapping1 {
		if indices2, exists := rowsMapping2[key]; exists {
			minLength := min(len(indices1), len(indices2))
			commonIndices1 = append(commonIndices1, indices1[:minLength]...)
			commonIndices2 = append(commonIndices2, indices2[:minLength]...)
		}
	}
	return commonIndices1, commonIndices2
}

// Returns the indices of rows common to both arrays
// using the direct method.
func getCommonIndicesDirect(arr1, arr2 [][]StringHashable) ([]int, []int) {
	commonIndices1 := []int{}
	commonIndices2 := []int{}
	for i := 0; i < len(arr1) && i < len(arr2); i++ {
		if getRowKey(arr1[i]) == getRowKey(arr2[i]) {
			commonIndices1 = append(commonIndices1, i)
			commonIndices2 = append(commonIndices2, i)
		}
	}
	return commonIndices1, commonIndices2
}

// Returns the indices of rows common to both arrays
// using the set method.
func getCommonIndicesSet(arr1, arr2 [][]StringHashable) ([]int, []int) {
	rowsMapping1 := getRowsMapping(arr1)
	rowsMapping2 := getRowsMapping(arr2)

	commonIndices1 := []int{}
	commonIndices2 := []int{}
	for key, indices1 := range rowsMapping1 {
		if indices2, exists := rowsMapping2[key]; exists {
			commonIndices1 = append(commonIndices1, indices1...)
			commonIndices2 = append(commonIndices2, indices2...)
		}
	}
	return commonIndices1, commonIndices2
}

// Returns the indices of rows common to both arrays
// based on the method given.
func GetCommonIndices(arr1, arr2 [][]StringHashable, method int, sortIndices bool) ([]int, []int, error) {
	var indices1 []int
	var indices2 []int

	switch method {
	case MethodMatch:
		indices1, indices2 = getCommonIndicesMatch(arr1, arr2)
	case MethodDirect:
		indices1, indices2 = getCommonIndicesDirect(arr1, arr2)
	case MethodSet:
		indices1, indices2 = getCommonIndicesSet(arr1, arr2)
	default:
		return nil, nil, fmt.Errorf("unsupported method: %d", method)
	}

	if sortIndices {
		sort.Ints(indices1)
		sort.Ints(indices2)
	}

	return indices1, indices2, nil
}

// Returns the indices of rows that are different between the two arrays
// using the match method.
func getDifferentIndicesMatch(arr1, arr2 [][]StringHashable) ([]int, []int) {
	rowsMapping1 := getRowsMapping(arr1)
	rowsMapping2 := getRowsMapping(arr2)

	differentIndices1 := []int{}
	differentIndices2 := []int{}
	for key, indices1 := range rowsMapping1 {
		if indices2, exists := rowsMapping2[key]; !exists {
			differentIndices1 = append(differentIndices1, indices1...)
		} else {
			if len(indices1) > len(indices2) {
				differentIndices1 = append(differentIndices1, indices1[len(indices2):]...)
			} else if len(indices2) > len(indices1) {
				differentIndices2 = append(differentIndices2, indices2[len(indices1):]...)
			}
		}
	}
	for key, indices2 := range rowsMapping2 {
		if _, exists := rowsMapping1[key]; !exists {
			differentIndices2 = append(differentIndices2, indices2...)
		}
	}
	return differentIndices1, differentIndices2
}

// Returns the indices of rows that are different between the two arrays
// using the direct method.
func getDifferentIndicesDirect(arr1, arr2 [][]StringHashable) ([]int, []int) {
	differentIndices1 := []int{}
	differentIndices2 := []int{}

	i := 0
	for ; i < len(arr1) && i < len(arr2); i++ {
		if getRowKey(arr1[i]) != getRowKey(arr2[i]) {
			differentIndices1 = append(differentIndices1, i)
			differentIndices2 = append(differentIndices2, i)
		}
	}

	for ; i < len(arr1); i++ {
		differentIndices1 = append(differentIndices1, i)
	}
	for ; i < len(arr2); i++ {
		differentIndices2 = append(differentIndices2, i)
	}

	return differentIndices1, differentIndices2
}

// Returns the indices of rows that are different between the two arrays
// using the set method.
func getDifferentIndicesSet(arr1, arr2 [][]StringHashable) ([]int, []int) {
	rowsMapping1 := getRowsMapping(arr1)
	rowsMapping2 := getRowsMapping(arr2)

	differentIndices1 := []int{}
	differentIndices2 := []int{}
	for key, indices1 := range rowsMapping1 {
		if _, exists := rowsMapping2[key]; !exists {
			differentIndices1 = append(differentIndices1, indices1...)
		}
	}
	for key, indices2 := range rowsMapping2 {
		if _, exists := rowsMapping1[key]; !exists {
			differentIndices2 = append(differentIndices2, indices2...)
		}
	}

	return differentIndices1, differentIndices2
}

// Returns the indices of rows that are different between the two arrays
// based on the method given.
func GetDifferentIndices(arr1, arr2 [][]StringHashable, method int, sortIndices bool) ([]int, []int, error) {
	var indices1 []int
	var indices2 []int

	switch method {
	case MethodMatch:
		indices1, indices2 = getDifferentIndicesMatch(arr1, arr2)
	case MethodDirect:
		indices1, indices2 = getDifferentIndicesDirect(arr1, arr2)
	case MethodSet:
		indices1, indices2 = getDifferentIndicesSet(arr1, arr2)
	default:
		return nil, nil, fmt.Errorf("unsupported method: %d", method)
	}

	if sortIndices {
		sort.Ints(indices1)
		sort.Ints(indices2)
	}

	return indices1, indices2, nil
}

// Returns nil iff the array is not empty, has no duplicate columns,
// and all rows have the same number of columns.
func CheckForProperCsvArray(arr [][]StringHashable) error {
	if len(arr) == 0 {
		return fmt.Errorf("empty array")
	}

	marker := make(map[uint64]bool)
	for _, column := range arr[0] {
		s := column.StringHash()
		key := xxhash.Sum64String(s)
		if _, exists := marker[key]; exists {
			return fmt.Errorf("duplicate column: %s", s)
		}
		marker[key] = true
	}

	length := len(arr[0])
	for i, row := range arr {
		if len(row) != length {
			return fmt.Errorf("row %d has %d columns, expected %d", i, len(row), length)
		}
	}
	return nil
}

// Returns a sorted list of indices in arr whose corresponding value exists in values.
func getIndicesInRow(arr, values []StringHashable) []int {
	res := []int{}

	marker := make(map[uint64]bool)
	for _, column := range values {
		marker[getStringKey(column)] = true
	}

	for i, column := range arr {
		if _, exists := marker[getStringKey(column)]; exists {
			res = append(res, i)
		}
	}

	return res
}

// Returns a new csv array with only the columns specified.
func KeepColumns(arr [][]StringHashable, columns []StringHashable) ([][]StringHashable, error) {
	err := CheckForProperCsvArray(arr)
	if err != nil {
		return nil, err
	}

	var indicesToKeep []int
	if columns == nil {
		indicesToKeep = make([]int, len(arr[0]))
		for i := range arr[0] {
			indicesToKeep[i] = i
		}
	} else {
		indicesToKeep = getIndicesInRow(arr[0], columns)
	}

	res := make([][]StringHashable, len(arr))
	length := len(indicesToKeep)
	for i, row := range arr {
		res[i] = make([]StringHashable, length)
		for j, index := range indicesToKeep {
			res[i][j] = row[index]
		}
	}
	return res, nil
}

// Returns a new csv array with the columns specified removed.
func IgnoreColumns(arr [][]StringHashable, columns []StringHashable) ([][]StringHashable, error) {
	err := CheckForProperCsvArray(arr)
	if err != nil {
		return nil, err
	}

	var indicesToIgnore []int
	if columns == nil {
		indicesToIgnore = []int{}
	} else {
		indicesToIgnore = getIndicesInRow(arr[0], columns)
	}

	indicesToKeep := []int{}
	i := 0
	j := 0
	for ; i < len(arr[0]); i++ {
		if j >= len(indicesToIgnore) || i != indicesToIgnore[j] {
			indicesToKeep = append(indicesToKeep, i)
		} else {
			j++
		}
	}

	res := make([][]StringHashable, len(arr))
	length := len(indicesToKeep)
	for i, row := range arr {
		res[i] = make([]StringHashable, length)
		for j, index := range indicesToKeep {
			res[i][j] = row[index]
		}
	}
	return res, nil
}

// Returns a new 2D array with only the rows specified.
func KeepRows(arr [][]StringHashable, rows []int) ([][]StringHashable, error) {
	err := CheckForProperCsvArray(arr)
	if err != nil {
		return nil, err
	}

	rowsCopy := make([]int, len(rows))
	copy(rowsCopy, rows)
	sort.Ints(rowsCopy)

	res := [][]StringHashable{}
	j := 0
	for i, row := range arr {
		if j < len(rowsCopy) && i == rowsCopy[j] {
			res = append(res, row)
			for j < len(rowsCopy) && i == rowsCopy[j] {
				j++
			}
		}
	}
	return res, nil
}

// Returns a new 2D array with the rows specified removed.
func IgnoreRows(arr [][]StringHashable, rows []int) ([][]StringHashable, error) {
	err := CheckForProperCsvArray(arr)
	if err != nil {
		return nil, err
	}

	ignore := make(map[int]bool)
	for _, row := range rows {
		ignore[row] = true
	}

	res := [][]StringHashable{}
	for i, row := range arr {
		if _, exists := ignore[i]; !exists {
			res = append(res, row)
		}
	}
	return res, nil
}

// Helper function for getting all the rows below the columns row for comparison purposes.
func getBelowComparisonArrays(arr1, arr2 [][]StringHashable, options Options) ([][]StringHashable, [][]StringHashable, error) {
	var comparisonArray1 [][]StringHashable
	var comparisonArray2 [][]StringHashable
	if options.UseColumns != nil {
		comparisonArray1, _ = KeepColumns(arr1, options.UseColumns)
		comparisonArray2, _ = KeepColumns(arr2, options.UseColumns)
	} else if options.IgnoreColumns != nil {
		comparisonArray1, _ = IgnoreColumns(arr1, options.IgnoreColumns)
		comparisonArray2, _ = IgnoreColumns(arr2, options.IgnoreColumns)
	} else {
		comparisonArray1 = arr1
		comparisonArray2 = arr2
	}

	columns1 := comparisonArray1[0]
	columns2 := comparisonArray2[0]
	if len(columns1) == 0 || len(columns2) == 0 {
		return nil, nil, fmt.Errorf("no columns to compare")
	} else if !rowsArePermutationsOfEachOther(columns1, columns2) {
		return nil, nil, fmt.Errorf("check the columns being compared")
	}

	comparisonArray2, _ = RearrangeColumns(comparisonArray2, columns1)

	comparisonArray1, _ = IgnoreRows(comparisonArray1, []int{0})
	comparisonArray2, _ = IgnoreRows(comparisonArray2, []int{0})

	return comparisonArray1, comparisonArray2, nil
}

// Helper function that adds 1 to all elements of the array.
func addOneToIntArray(arr []int) []int {
	res := make([]int, len(arr))
	copy(res, arr)
	for i := range res {
		res[i]++
	}
	return res
}

// Returns the common rows between the two arrays based on the
// given options and the indices of the rows in the results in the
// original arrays.
func GetCommonRows(csvArray1, csvArray2 [][]StringHashable, options Options) ([][]StringHashable, [][]StringHashable, []int, []int, error) {
	err := CheckForProperCsvArray(csvArray1)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	err = CheckForProperCsvArray(csvArray2)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	err = options.CheckAttributes()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	belowArray1, belowArray2, err := getBelowComparisonArrays(csvArray1, csvArray2, options)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	belowIndices1, belowIndices2, _ := GetCommonIndices(belowArray1, belowArray2, options.Method, options.SortIndices)

	indices1 := append([]int{0}, addOneToIntArray(belowIndices1)...)
	indices2 := append([]int{0}, addOneToIntArray(belowIndices2)...)

	res1, _ := KeepRows(csvArray1, indices1)
	res2, _ := KeepRows(csvArray2, indices2)

	return res1, res2, indices1, indices2, nil
}

// Returns the different rows between the two arrays based on the
// given options and the indices of the rows in the results in the
// original arrays.
func GetDifferentRows(csvArray1, csvArray2 [][]StringHashable, options Options) ([][]StringHashable, [][]StringHashable, []int, []int, error) {
	err := CheckForProperCsvArray(csvArray1)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	err = CheckForProperCsvArray(csvArray2)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	err = options.CheckAttributes()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	belowArray1, belowArray2, err := getBelowComparisonArrays(csvArray1, csvArray2, options)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	belowIndices1, belowIndices2, _ := GetDifferentIndices(belowArray1, belowArray2, options.Method, options.SortIndices)

	indices1 := append([]int{0}, addOneToIntArray(belowIndices1)...)
	indices2 := append([]int{0}, addOneToIntArray(belowIndices2)...)

	res1, _ := KeepRows(csvArray1, indices1)
	res2, _ := KeepRows(csvArray2, indices2)

	return res1, res2, indices1, indices2, nil
}

// Returns a csv array with the columns rearranged accordingly.
func RearrangeColumns(arr [][]StringHashable, columns []StringHashable) ([][]StringHashable, error) {
	err := CheckForProperCsvArray(arr)
	if err != nil {
		return nil, err
	}

	marker := make(map[uint64]bool)
	for _, column := range columns {
		marker[getStringKey(column)] = true
	}

	mapping := make(map[uint64]int)
	for i, column := range arr[0] {
		s := column.StringHash()
		stringHash := xxhash.Sum64String(s)
		if _, exists := marker[stringHash]; !exists {
			return nil, fmt.Errorf("column %s not found", s)
		}
		mapping[stringHash] = i
	}

	if len(mapping) != len(columns) {
		return nil, fmt.Errorf("columns must use the same names")
	}

	columnsStringHashes := make([]uint64, len(columns))
	for i, column := range columns {
		columnsStringHashes[i] = getStringKey(column)
	}

	res := make([][]StringHashable, len(arr))
	for i, row := range arr {
		newRow := []StringHashable{}
		for j := range columns {
			newRow = append(newRow, row[mapping[columnsStringHashes[j]]])
		}
		res[i] = newRow
	}
	return res, nil
}

// Automatically aligns the common columns on the left side. The relative positions of
// the common columns are the same as that of csvArray1 and uncommon columns are placed at the end
// with the same relative positions as that of the original arrays.
func AutoAlignCsvArrays(csvArray1, csvArray2 [][]StringHashable) ([][]StringHashable, [][]StringHashable, error) {
	err := CheckForProperCsvArray(csvArray1)
	if err != nil {
		return nil, nil, err
	}
	err = CheckForProperCsvArray(csvArray2)
	if err != nil {
		return nil, nil, err
	}

	marker2 := make(map[uint64]int)
	for i, v := range csvArray2[0] {
		marker2[getStringKey(v)] = i
	}

	common := make(map[uint64]bool)
	newColumnIndices1 := []int{}
	newColumnIndices2 := []int{}
	tail1 := []int{}
	for i, v := range csvArray1[0] {
		key := getStringKey(v)
		if _, exists := marker2[key]; exists {
			newColumnIndices1 = append(newColumnIndices1, i)
			newColumnIndices2 = append(newColumnIndices2, marker2[key])
			common[key] = true
		} else {
			tail1 = append(tail1, i)
		}
	}
	newColumnIndices1 = append(newColumnIndices1, tail1...)

	tail2 := []int{}
	for i, v := range csvArray2[0] {
		if _, exists := common[getStringKey(v)]; !exists {
			tail2 = append(tail2, i)
		}
	}
	newColumnIndices2 = append(newColumnIndices2, tail2...)

	newCsvArray1 := make([][]StringHashable, len(csvArray1))
	for i, row := range csvArray1 {
		newRow := []StringHashable{}
		for _, j := range newColumnIndices1 {
			newRow = append(newRow, row[j])
		}
		newCsvArray1[i] = newRow
	}

	newCsvArray2 := make([][]StringHashable, len(csvArray2))
	for i, row := range csvArray2 {
		newRow := []StringHashable{}
		for _, j := range newColumnIndices2 {
			newRow = append(newRow, row[j])
		}
		newCsvArray2[i] = newRow
	}

	return newCsvArray1, newCsvArray2, nil
}

// Returns the common columns between the two csv arrays keeping
// the order of the relative positions of the columns in csvArray1.
func GetCommonColumns(csvArray1, csvArray2 [][]StringHashable) ([]StringHashable, error) {
	err := CheckForProperCsvArray(csvArray1)
	if err != nil {
		return nil, err
	}
	err = CheckForProperCsvArray(csvArray2)
	if err != nil {
		return nil, err
	}

	marker2 := make(map[uint64]bool)
	for _, column := range csvArray2[0] {
		marker2[getStringKey(column)] = true
	}

	common := []StringHashable{}
	for _, column := range csvArray1[0] {
		if _, exists := marker2[getStringKey(column)]; exists {
			common = append(common, column)
		}
	}
	return common, nil
}

// Takes a csv array and returns a columns-aligned formatted string
// according to spaces. Strings with lengths exceeding maxColLength are truncated.
// Use a negative value for maxColLength to keep strings of all lengths.
func PrettyFormatCsvArray(csvArray [][]StringHashable, spaces int, maxColLength int) (string, error) {
	err := CheckForProperCsvArray(csvArray)
	if err != nil {
		return "", err
	}

	if spaces < 0 {
		return "", fmt.Errorf("spaces must be non-negative")
	}

	passedMaxColLengthLength := maxColLength + len(TruncatedMark)
	rowLength := len(csvArray[0])
	maxLengths := make([]int, rowLength)
	for _, row := range csvArray {
		for i, cell := range row {
			maxLengths[i] = max(maxLengths[i], len(cell.StringHash()))
		}
	}
	if maxColLength >= 0 {
		for i, v := range maxLengths {
			if v > maxColLength {
				maxLengths[i] = passedMaxColLengthLength
			}
		}
	}

	holder := make([]string, rowLength*len(csvArray))
	i := 0
	for _, row := range csvArray {
		for j, cell := range row {
			s := cell.StringHash()
			if maxColLength >= 0 && len(s) > maxColLength {
				s = s[:maxColLength] + TruncatedMark
			}

			if j < rowLength-1 {
				holder[i] = fmt.Sprintf("%-*s", maxLengths[j]+spaces, s)
			} else {
				holder[i] = fmt.Sprintf("%s\n", s)
			}
			i++
		}
	}
	res := strings.Join(holder, "")

	return res, nil
}

// Takes a csv array and returns a csv formatted string.
func StringFormatCsvArray(csvArray [][]StringHashable) (string, error) {
	err := CheckForProperCsvArray(csvArray)
	if err != nil {
		return "", err
	}

	holder := make([]string, len(csvArray))
	for i, row := range csvArray {
		holder[i] += strings.Join(getStringsRow(row), ",") + "\n"
	}
	res := strings.Join(holder, "")

	return res, nil
}

// Returns a StringHashable row from a string row.
func GetRowFromRow(arr []string) []StringHashable {
	if arr == nil {
		return nil
	}

	res := make([]StringHashable, len(arr))
	for i, cell := range arr {
		res[i] = BasicStringHashable(cell)
	}
	return res
}

// Returns a 2D array of StringHashable from a 2D array of strings.
func Get2DArrayFrom2DArray(arr [][]string) [][]StringHashable {
	if arr == nil {
		return nil
	}

	res := make([][]StringHashable, len(arr))
	for i, row := range arr {
		res[i] = make([]StringHashable, len(row))
		for j, cell := range row {
			res[i][j] = BasicStringHashable(cell)
		}
	}
	return res
}
