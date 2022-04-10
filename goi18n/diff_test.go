package main

import (
	"testing"
)

func TestDiff1(t *testing.T) {
	t.Log("情景 1")
	runTest(t, M{
		"A": {ID: "A", One: "123"},
	}, M{}, M{})

	t.Log("情景 2-1")
	runTest(t, M{
		"B": {ID: "B", One: "123"},
	}, M{
		"B": {ID: "B", One: "123"},
	}, M{})

	t.Log("情景 2-2")
	runTest(t, M{
		"C": {ID: "C", One: "123"},
	}, M{
		"C": {ID: "C", One: "456"},
	}, M{})

	t.Log("情景 3-1")
	runTest(t, M{
		"D": {ID: "D", One: "123"},
	}, M{}, M{
		"D": {ID: "D", One: "123"},
	})

	t.Log("情景 3-2")
	runTest(t, M{
		"E": {ID: "E", One: "123"},
	}, M{}, M{
		"E": {ID: "E", One: "456"},
	})

	t.Log("情景 4-1")
	runTest(t, M{
		"F": {ID: "F", One: "123"},
	}, M{
		"F": {ID: "F", One: "123"},
	}, M{
		"F": {ID: "F", One: "456"},
	})

	t.Log("情景 4-2")
	runTest(t, M{
		"G": {ID: "G", One: "123"},
	}, M{
		"G": {ID: "G", One: "456"},
	}, M{
		"G": {ID: "G", One: "456"},
	})
}

func runTest(t *testing.T, A, B, C M) {
	sourceMapResult, middleMapResult, targetMapResult := diff(A, B, C)

	printResult(t, "source", sourceMapResult)
	printResult(t, "middle", middleMapResult)
	printResult(t, "target", targetMapResult)
}

func printResult(t *testing.T, n string, m M) {
	t.Logf("\t%s", n)
	for k, _ := range m {
		t.Logf("\t\t%s", k)
	}
}
