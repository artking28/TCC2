package models

import (
	"log"

	"github.com/tcc2-davi-arthur/models/interfaces"
)

func AssignIGram(size int, docID, wdId0, wdId1, wdId2 uint16, jump0, jump1 int8) (ngram interfaces.IGram) {
	switch size {
	case 1:
		ngram = &InverseUnigram{
			Wd0Id: wdId0,
			DocId: docID,
		}
		break
	case 2:
		ngram = &InverseBigram{
			Wd0Id: wdId0,
			Wd1Id: wdId1,
			Jump0: jump0,
			DocId: docID,
		}
		break
	case 3:
		ngram = &InverseTrigram{
			Wd0Id: wdId0,
			Wd1Id: wdId1,
			Wd2Id: wdId2,
			Jump0: jump0,
			Jump1: jump1,
			DocId: docID,
		}
		break
	default:
		log.Fatalf("invalid gramSize: %d", size)
	}
	return ngram
}
