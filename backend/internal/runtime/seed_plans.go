package runtime

import "sort"

// SeedColumn is one column the seed statement fills, with the SQL expression
// producing its value for the generated row `g.i`.
type SeedColumn struct {
	Name string
	Expr string
}

// SeedPlan is one concrete table to fill with generated rows, and the columns to
// fill it through.
type SeedPlan struct {
	Table   string
	Columns []SeedColumn
}

// SeedPlans says which tables to fill and how, given the tables the migration
// itself names (from TouchedTables) and what the live catalog knows about them.
//
// It crosses from the tables of the migration to the tables *related* to them,
// because those are what actually hold the rows: a partitioned parent stores
// nothing itself, so it resolves to its leaf partitions, each pinned to a key
// value its own bound admits. A parent whose bounds admit no constant — hash
// partitioning — stays the plan, and tuple routing places the rows.
//
// A table the catalog does not hold is skipped, since the migration creates it.
func SeedPlans(catalog map[string]CatalogTable, migrationTables []string) []SeedPlan {
	var plans []SeedPlan
	seen := map[string]bool{}

	add := func(info CatalogTable) {
		if seen[info.Name] {
			return
		}
		seen[info.Name] = true
		plans = append(plans, SeedPlan{Table: info.Name, Columns: seedColumns(catalog, info)})
	}

	for _, table := range migrationTables {
		info, ok := catalog[table]
		if !ok {
			continue
		}
		if !info.Partitioned {
			add(info)
			continue
		}
		leaves := seedableLeaves(catalog, info)
		if len(leaves) == 0 {
			add(info)
			continue
		}
		for _, leaf := range leaves {
			add(leaf)
		}
	}
	return plans
}

// seedColumns pairs every fillable column with the expression that gives it a
// value, dropping the ones no generic filler covers.
func seedColumns(catalog map[string]CatalogTable, info CatalogTable) []SeedColumn {
	pinned := partitionKeyValues(catalog, info)

	var columns []SeedColumn
	for _, column := range info.Columns {
		if column.AutoFilled {
			continue
		}
		expr, ok := pinned[column.Name]
		if !ok {
			expr = SeedExpr(column.DataType, column.MaxLength)
		}
		if expr == "" {
			continue
		}
		columns = append(columns, SeedColumn{Name: column.Name, Expr: expr})
	}
	return columns
}

// partitionKeyValues pins the partition key columns of a leaf to values its own
// bound admits, so rows inserted straight into it satisfy the bound.
func partitionKeyValues(catalog map[string]CatalogTable, info CatalogTable) map[string]string {
	parent, ok := catalog[info.Parent]
	if !ok {
		return nil
	}

	keys := PartitionKeyColumns(parent.PartitionKeyDef)
	values := PartitionSeedValues(info.Bound)
	if keys == nil || values == nil {
		return nil
	}

	pinned := map[string]string{}
	for i := range keys {
		if i >= len(values) {
			break
		}
		pinned[keys[i]] = values[i]
	}
	return pinned
}

// seedableLeaves are the partitions under a partitioned table whose bounds admit
// a constant, descending through sub-partitioned levels.
func seedableLeaves(catalog map[string]CatalogTable, parent CatalogTable) []CatalogTable {
	var leaves []CatalogTable
	for _, child := range childrenOf(catalog, parent.Name) {
		if child.Partitioned {
			leaves = append(leaves, seedableLeaves(catalog, child)...)
			continue
		}
		if len(partitionKeyValues(catalog, child)) > 0 {
			leaves = append(leaves, child)
		}
	}
	return leaves
}

// childrenOf returns the partitions of a table, in name order so a seed run is
// reproducible whatever order the catalog was built in.
func childrenOf(catalog map[string]CatalogTable, parent string) []CatalogTable {
	var children []CatalogTable
	for _, info := range catalog {
		if info.Parent == parent {
			children = append(children, info)
		}
	}
	sort.Slice(children, func(i, j int) bool { return children[i].Name < children[j].Name })
	return children
}
