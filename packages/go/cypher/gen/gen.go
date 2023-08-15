package gen

import (
	"bytes"
	"github.com/specterops/bloodhound/cypher/frontend"
	"github.com/specterops/bloodhound/cypher/gen/cypher"
	"github.com/specterops/bloodhound/cypher/model"
	"io"
)

type Emitter interface {
	Write(query *model.RegularQuery, writer io.Writer) error
	WriteExpression(output io.Writer, expression model.Expression) error
}

func CypherToCypher(ctx *frontend.Context, input string) (string, error) {
	if query, err := frontend.ParseCypher(ctx, input); err != nil {
		return "", err
	} else {
		var (
			output  = &bytes.Buffer{}
			emitter = cypher.Emitter{
				StripLiterals: false,
			}
		)

		if err := emitter.Write(query, output); err != nil {
			return "", err
		}

		return output.String(), nil
	}
}
