// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: chain.sql

package db

import (
	"context"
)

const getChains = `-- name: GetChains :many
SELECT id, name, chain_id, rpc_url, native_currency, created_at, updated_at, explorer_url, native_token_id, ws_url, last_scan_block_number, is_active FROM chains
`

func (q *Queries) GetChains(ctx context.Context) ([]Chain, error) {
	rows, err := q.db.Query(ctx, getChains)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Chain
	for rows.Next() {
		var i Chain
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.ChainID,
			&i.RpcUrl,
			&i.NativeCurrency,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.ExplorerUrl,
			&i.NativeTokenID,
			&i.WsUrl,
			&i.LastScanBlockNumber,
			&i.IsActive,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
