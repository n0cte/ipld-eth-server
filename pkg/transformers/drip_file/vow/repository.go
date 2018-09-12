// Copyright 2018 Vulcanize
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vow

import (
	"github.com/vulcanize/vulcanizedb/pkg/core"
	"github.com/vulcanize/vulcanizedb/pkg/datastore/postgres"
)

type Repository interface {
	Create(headerID int64, model DripFileVowModel) error
	MissingHeaders(startingBlockNumber, endingBlockNumber int64) ([]core.Header, error)
}

type DripFileVowRepository struct {
	db *postgres.DB
}

func NewDripFileVowRepository(db *postgres.DB) DripFileVowRepository {
	return DripFileVowRepository{db: db}
}

func (repository DripFileVowRepository) Create(headerID int64, model DripFileVowModel) error {
	_, err := repository.db.Exec(
		`INSERT into maker.drip_file_vow (header_id, what, data, tx_idx, raw_log)
        VALUES($1, $2, $3, $4, $5)`,
		headerID, model.What, model.Data, model.TransactionIndex, model.Raw,
	)
	return err
}

func (repository DripFileVowRepository) MissingHeaders(startingBlockNumber, endingBlockNumber int64) ([]core.Header, error) {
	var result []core.Header
	err := repository.db.Select(
		&result,
		`SELECT headers.id, headers.block_number FROM headers
               LEFT JOIN maker.drip_file_vow on headers.id = header_id
               WHERE header_id ISNULL
               AND headers.block_number >= $1
               AND headers.block_number <= $2
               AND headers.eth_node_fingerprint = $3`,
		startingBlockNumber,
		endingBlockNumber,
		repository.db.Node.ID,
	)
	return result, err
}