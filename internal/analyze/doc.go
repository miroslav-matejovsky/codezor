// Package analyze counts a source tree with scc and estimates what producing
// that code with an LLM would cost.
//
// scc is driven through package level globals, so Run sets them on every call
// and must not be called concurrently.
package analyze
