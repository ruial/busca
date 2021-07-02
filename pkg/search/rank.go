package search

import (
	"math"
	"sort"

	"github.com/ruial/busca/pkg/core"
)

type Ranker func(terms []string, docsData map[core.DocumentID]core.DocumentData) (documentScores []core.DocumentScore)

type TfWeightScheme int
type IdfWeightScheme int

const (
	TfWeightDefault TfWeightScheme = iota
	TfWeightLog
)

const (
	IdfWeightDefault IdfWeightScheme = iota
	IdfWeightSmooth
)

func CalcIdf(term string, docsData map[core.DocumentID]core.DocumentData, idfScheme IdfWeightScheme) (idf float64) {
	for _, data := range docsData {
		if data.Frequencies[term] > 0 {
			idf++
		}
	}
	if idfScheme == IdfWeightSmooth {
		idf = 1 + math.Log10((1+float64(len(docsData)))/(1+idf))
	} else if idfScheme == IdfWeightDefault {
		if idf > 0 {
			idf = math.Log10(float64(len(docsData)) / idf)
		}
	}
	return
}

func GetTermsIdf(terms []string, docsData map[core.DocumentID]core.DocumentData, idfScheme IdfWeightScheme) core.TermFrequency {
	termsIdf := make(core.TermFrequency)
	for _, term := range terms {
		termsIdf[term] = CalcIdf(term, docsData, idfScheme)
	}
	return termsIdf
}

func CalcDocumentScore(terms []string, data core.DocumentData, termsIdf core.TermFrequency, tfScheme TfWeightScheme, idfScheme IdfWeightScheme) (score float64) {
	for _, term := range terms {
		var tf float64
		if tfScheme == TfWeightLog {
			tf = math.Log10(1 + data.Frequencies[term])
		} else if tfScheme == TfWeightDefault {
			tf = data.Frequencies[term] / data.TermsCount
		}
		score += tf * termsIdf[term]
	}
	return
}

func TfIdfRanker(tfWeightScheme TfWeightScheme, idfWeightScheme IdfWeightScheme) Ranker {
	return func(terms []string, docsData map[core.DocumentID]core.DocumentData) []core.DocumentScore {
		documentScores := make([]core.DocumentScore, 0, len(docsData))
		termsIdf := GetTermsIdf(terms, docsData, idfWeightScheme)
		for _, data := range docsData {
			score := CalcDocumentScore(terms, data, termsIdf, tfWeightScheme, idfWeightScheme)
			documentScore := core.DocumentScore{Doc: data.Doc, Score: score}
			documentScores = append(documentScores, documentScore)
		}
		// could use heap instead, but same algorithmic complexity
		sort.Slice(documentScores, func(i, j int) bool {
			return documentScores[i].Score > documentScores[j].Score
		})
		return documentScores
	}
}
