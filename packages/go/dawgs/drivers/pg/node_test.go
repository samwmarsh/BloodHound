package pg

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	pgxMocks "github.com/specterops/bloodhound/dawgs/vendormocks/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
)

type testKindMapper struct {
	known map[string]int16
}

func (s testKindMapper) MapKinds(kinds graph.Kinds) ([]int16, graph.Kinds) {
	var (
		kindIDs      = make([]int16, 0, len(kinds))
		missingKinds = make([]graph.Kind, 0, len(kinds))
	)

	for _, kind := range kinds {
		if kindID, hasKind := s.known[kind.String()]; hasKind {
			kindIDs = append(kindIDs, kindID)
		} else {
			missingKinds = append(missingKinds, kind)
		}
	}

	return kindIDs, missingKinds
}

func TestNodeQuery(t *testing.T) {
	var (
		mockCtrl = gomock.NewController(t)
		mockTx   = pgxMocks.NewMockTx(mockCtrl)

		kindMapper = testKindMapper{
			known: map[string]int16{
				"NodeKindA": 1,
				"NodeKindB": 2,
				"EdgeKindA": 3,
				"EdgeKindB": 4,
			},
		}

		nodeQueryInst = &nodeQuery{
			ctx: context.Background(),
			tx: &transaction{
				queryExecMode:   pgx.QueryExecModeCacheStatement,
				ctx:             context.Background(),
				tx:              mockTx,
				targetSchemaSet: false,
			},
			queryBuilder: newQueryBuilder(kindMapper),
		}
	)

	nodeQueryInst.Filter(
		query.Where(
			query.Equals(query.NodeProperty("prop"), "1234"),
		),
	)

	_, err := nodeQueryInst.First()
	require.Nil(t, err)
}
