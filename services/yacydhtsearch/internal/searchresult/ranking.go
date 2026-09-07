package searchresult

import "github.com/nikitakarpei/yacy-rwi-node/yacymodel"

type Ranking struct {
	Items []Item `json:"items"`
}

func RankingFrom(answers [][]Item, ceiling int) Ranking {
	if ceiling <= 0 {
		return Ranking{}
	}

	items := make([]Item, 0, ceiling)
	taken := map[yacymodel.URLHash]struct{}{}
	for round := 0; round < longest(answers) && len(items) < ceiling; round++ {
		for _, answer := range answers {
			if round >= len(answer) || len(items) == ceiling {
				continue
			}
			item := answer[round]
			if _, seen := taken[item.Hash]; seen {
				continue
			}
			taken[item.Hash] = struct{}{}
			items = append(items, item)
		}
	}

	return Ranking{Items: items}
}

func (r Ranking) PageFrom(startRecord, records int) Page {
	if startRecord < 0 || startRecord >= len(r.Items) || records <= 0 {
		return Page{}
	}

	return Page{Items: r.Items[startRecord:min(startRecord+records, len(r.Items))]}
}

func longest(answers [][]Item) int {
	var rounds int
	for _, answer := range answers {
		rounds = max(rounds, len(answer))
	}

	return rounds
}
