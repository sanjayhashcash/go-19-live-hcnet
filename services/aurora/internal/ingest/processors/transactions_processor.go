package processors

import (
	"context"

	"github.com/hcnet/go/ingest"
	"github.com/hcnet/go/services/aurora/internal/db2/history"
	"github.com/hcnet/go/support/db"
	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/xdr"
)

type TransactionProcessor struct {
	batch history.TransactionBatchInsertBuilder
}

func NewTransactionFilteredTmpProcessor(batch history.TransactionBatchInsertBuilder) *TransactionProcessor {
	return &TransactionProcessor{
		batch: batch,
	}
}

func NewTransactionProcessor(batch history.TransactionBatchInsertBuilder) *TransactionProcessor {
	return &TransactionProcessor{
		batch: batch,
	}
}

func (p *TransactionProcessor) ProcessTransaction(lcm xdr.LedgerCloseMeta, transaction ingest.LedgerTransaction) error {
	if err := p.batch.Add(transaction, lcm.LedgerSequence()); err != nil {
		return errors.Wrap(err, "Error batch inserting transaction rows")
	}

	return nil
}

func (p *TransactionProcessor) Flush(ctx context.Context, session db.SessionInterface) error {
	return p.batch.Exec(ctx, session)
}
