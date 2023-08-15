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

-- Drop triggers
drop trigger if exists delete_node_edges on node;
drop function if exists delete_node_edges();

-- Drop all tables in order of dependency.
drop table if exists node;
drop table if exists edge;
drop table if exists kind;
drop table if exists graph;

-- Remove custom types
do
$$
  begin
    drop type nodeComposite;
  exception
    when undefined_object then null;
  end
$$;

do
$$
  begin
    drop type edgeComposite;
  exception
    when undefined_object then null;
  end
$$;

-- Pull the tri-gram and intarray extensions.
drop extension if exists pg_trgm;
drop extension if exists intarray;
