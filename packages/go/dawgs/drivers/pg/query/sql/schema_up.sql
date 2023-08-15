-- Copyright 2023 Specter Ops, Inc.
--
-- Licensed under the Apache License, Version 2.0
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--     http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.
--
-- SPDX-License-Identifier: Apache-2.0

-- DAWGS Property Graph Partitioned Layout for PostgreSQL

-- Notes on TOAST:
--
-- Graph entity properties are stored in a JSONB column at the end of the row. There is a soft-limit of 2KiB for rows in
-- a PostgreSQL database page. The database will compress this value in an attempt not to exceed this limit. Once a
-- compressed value reaches the absolute limit of what the database can do to either compact it or give it more of the
-- 8 KiB page size limit, the database evicts the value to an associated TOAST (The Oversized-Attribute Storage Technique)
-- table and creates a reference to the entry to be joined upon fetch of the row.
--
-- TOAST comes with certain performance caveats that can affect access time anywhere from a factor 3 to 6 times. It is
-- in the best interest of the database user that the properties of a graph entity never exceed this limit in large
-- graphs.

-- We need the tri-gram extension to create a GIN text-search index. The goal here isn't full-text search, in which
-- case ts_vector and its ilk would be more suited. This particular selection was made to support accelerated lookups
-- for "contains", "starts with" and, "ends with" comparison operations.
create extension if not exists pg_trgm;

-- We need the intarray extension for extended integer array operations like unions. This is useful for managing kind
-- arrays for nodes.
create extension if not exists intarray;

-- Table definitions

-- The graph table contains name to ID mappings for graphs contained within the database. Each graph ID should have
-- corresponding table partitions for the node and edge tables.
create table if not exists graph
(
  id   serial,
  name varchar(256) not null,

  primary key (id),
  unique (name)
);

-- The kind table contains name to ID mappings for graph kinds. Storage of these types is necessary to maintain search
-- capability of a database without the origin application that generated it.
create table if not exists kind
(
  id   smallserial,
  name varchar(256) not null,

  primary key (id),
  unique (name)
);

-- Node composite type
do
$$
  begin
    create type nodeComposite as
    (
      id         integer,
      kind_ids   smallint[8],
      properties jsonb
    );
  exception
    when duplicate_object then null;
  end
$$;

-- The node table is a partitioned table view that partitions over the graph ID that each node belongs to. Nodes may
-- contain a disjunction of up to 8 kinds for creating clique subsets without requiring edges.
create table if not exists node
(
  id         serial      not null,
  graph_id   integer     not null,
  kind_ids   smallint[8] not null,
  properties jsonb       not null,

  primary key (id, graph_id),

  foreign key (graph_id) references graph (id) on delete cascade
) partition by list (graph_id);

-- The storage strategy chosen for the properties JSONB column informs the database of the user's preference to resort
-- to creating a TOAST table entry only after there is no other possible way to inline the row attribute in the current
-- page.
alter table node
  alter column properties set storage main;

-- Index on the graph ID of each node.
create index if not exists node_graph_id_index on node using btree (graph_id);

-- Index node kind IDs so that lookups by kind is accelerated.
create index if not exists node_kind_ids_index on node using gin (kind_ids);

-- Edge composite type
do
$$
  begin
    create type edgeComposite as
    (
      id         integer,
      start_id   integer,
      end_id     integer,
      kind_id    smallint,
      properties jsonb
    );
  exception
    when duplicate_object then null;
  end
$$;

-- The edge table is a partitioned table view that partitions over the graph ID that each edge belongs to.
create table if not exists edge
(
  id         serial   not null,
  graph_id   integer  not null,
  start_id   integer  not null,
  end_id     integer  not null,
  kind_id    smallint not null,
  properties jsonb    not null,

  primary key (id, graph_id),

  foreign key (graph_id) references graph (id) on delete cascade
) partition by list (graph_id);

-- We have to use a trigger and associated plpgsql function to cascade delete edges when owning nodes are deleted. While
-- this could be done with a foreign key relationship, it would scope the cascade delete to individual node partitions
-- and therefore requires the graph_id of each node.
create or replace function delete_node_edges() returns trigger as
$$
begin
  delete from edge where start_id = OLD.id or end_id = OLD.id;
  return null;
end
$$
  language plpgsql;

-- Drop and create the delete_node_edges trigger for the delete_node_edges() plpgsql function. See the function comment
-- for more information.
drop trigger if exists delete_node_edges on node;
create trigger delete_node_edges
  after delete
  on node
  for each row
execute procedure delete_node_edges();


-- The storage strategy chosen for the properties JSONB column informs the database of the user's preference to resort
-- to creating a TOAST table entry only after there is no other possible way to inline the row attribute in the current
-- page.
alter table edge
  alter column properties set storage main;


-- Index on the graph ID of each edge.
create index if not exists edge_graph_id_index on edge using btree (graph_id);

-- Index on the start vertex of each edge.
create index if not exists edge_start_id_index on edge using btree (start_id);

-- Index on the start vertex of each edge.
create index if not exists edge_end_id_index on edge using btree (end_id);

-- Index on the kind of each edge.
create index if not exists edge_kind_index on edge using btree (kind_id);

-- Because arrays are hard
create or replace function public.reduce_dim(a_in anyarray, out a_out anyarray)
  returns setof anyarray as
$$
begin
  foreach a_out slice 1 in array a_in loop
      return next;
    end loop;
end;
$$
  language plpgsql immutable parallel safe strict;
