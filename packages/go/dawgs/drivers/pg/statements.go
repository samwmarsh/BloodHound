package pg

const (
	fetchNodeStatement                = `select kinds, properties from node where node.id = $1;`
	fetchNodeSliceStatement           = `select id, kinds, properties from node where node.id = any($1);`
	createNodeStatement               = `insert into node (graph_id, kind_ids, properties) values ($1, $2, $3) returning id;`
	createNodeWithoutIDBatchStatement = `insert into node (graph_id, kind_ids, properties) select $1, * from unnest($2::int2[][], $3::jsonb[])`
	createNodeWithIDBatchStatement    = `insert into node (graph_id, id, kind_ids, properties) select $1, unnest($2::int4[]), unnest($3::text[])::int2[], unnest($4::jsonb[])`
	deleteNodeStatement               = `delete from node where node.id = $1`
	deleteNodeSliceStatement          = `delete from node where node.id = any($1)`

	nodePropertySetOnlyStatement      = `update node set kind_ids = $1, properties = properties || $2::jsonb where node.id = $3`
	nodePropertyDeleteOnlyStatement   = `update node set kind_ids = $1, properties = properties - $2::text[] where node.id = $3`
	nodePropertySetAndDeleteStatement = `update node set kind_ids = $1, properties =  properties || $2::jsonb - $3::text[]) where node.id = $4`

	fetchEdgeStatement       = `select start_id, end_id, kind, properties from relationships where relationships.id = $1;`
	fetchEdgeSliceStatement  = `select id, start_id, end_id, kind, properties from node where relationships.id = any($1);`
	createEdgeStatement      = `insert into edge (graph_id, start_id, end_id, kind_id, properties) values ($1, $2, $3, $4, $5) returning id;`
	createEdgeBatchStatement = `insert into edge (graph_id, start_id, end_id, kind_id, properties) select $1, * from unnest($2::int4[], $3::int4[], $4::int2[], $5::jsonb[]);`
	deleteEdgeStatement      = `delete from edge where edge.id = $1`
	deleteEdgeSliceStatement = `delete from edge where edge.id = any($1)`

	edgePropertySetOnlyStatement      = `update edge set properties = properties || $1::jsonb where edge.id = $2`
	edgePropertyDeleteOnlyStatement   = `update edge set properties = properties - $1::text[] where edge.id = $2`
	edgePropertySetAndDeleteStatement = `update edge set properties = properties || $1::jsonb - $2::text[] where edge.id = $3`

	createNodesAndEdgeStatement = `with start_node as (insert into node (kinds, properties) values ($1, $2) returning id),
end_node as (insert into node (kinds, properties) values ($3, $4) returning id)

insert into relationships (start_id, end_id, kind, properties) values((select id from start_node), (select id from end_node), $5, $6);`

	createStartNodeAndEdgeStatement = `with start_node as (insert into node (kinds, properties) values ($1, $2) returning id)

insert into relationships (start_id, end_id, kind, properties) values((select id from start_node), $3, $4, $5);`

	createEndNodeAndEdgeStatement = `with end_node as (insert into node (kinds, properties) values ($1, $2) returning id)

insert into relationships (start_id, end_id, kind, properties) values($3, (select id from end_node), $4, $5);`
)
