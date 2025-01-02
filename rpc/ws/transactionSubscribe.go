package ws

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type TransactionResult struct {
	Transaction struct {
		Transaction []string             `json:"transaction,omitempty"`
		Meta        *rpc.TransactionMeta `json:"meta,omitempty"`
	} `json:"transaction"`
	Signature string `json:"signature,omitempty"`
	Slot      uint64 `json:"slot,omitempty"`
}

func (result *TransactionResult) parseTransaction() (*solana.Transaction, error) {
	for _, s := range result.Transaction.Transaction {
		if s == "base64" {
			continue
		}
		return solana.TransactionFromBase64(s)
	}
	return nil, nil
}

type TransactionSubscribeFilter struct {
	Vote            bool     `json:"vote"`
	Failed          bool     `json:"failed"`
	Signature       string   `json:"signature,omitempty"`
	AccountInclude  []string `json:"accountInclude,omitempty"`
	AccountExclude  []string `json:"accountExclude,omitempty"`
	AccountRequired []string `json:"accountRequired,omitempty"`
}

type TransactionSubscribeOptions struct {
	Commitment                     string `json:"commitment"`
	Encoding                       string `json:"encoding"`
	TransactionDetails             string `json:"transactionDetails"`
	ShowRewards                    bool   `json:"showRewards"`
	MaxSupportedTransactionVersion int    `json:"maxSupportedTransactionVersion"`
}

type TransactionSubscription struct {
	sub *Subscription
}

func (cl *Client) TransactionSubscribe(
	filter *TransactionSubscribeFilter,
	opts *TransactionSubscribeOptions,
) (*TransactionSubscription, error) {
	var params []interface{}
	if filter != nil {
		params = append(params, filter)
	}
	if opts != nil {
		params = append(params, opts)
	}
	genSub, err := cl.subscribe(
		params,
		nil,
		"transactionSubscribe",
		"transactionUnSubscribe",
		func(msg []byte) (interface{}, error) {
			var res TransactionResult
			err := decodeResponseFromMessage(msg, &res)
			return &res, err
		},
	)
	if err != nil {
		return nil, err
	}
	return &TransactionSubscription{
		sub: genSub,
	}, nil
}

func (sw *TransactionSubscription) Recv(ctx context.Context) (*TransactionResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case d, ok := <-sw.sub.stream:
		if !ok {
			return nil, ErrSubscriptionClosed
		}
		return d.(*TransactionResult), nil
	case err := <-sw.sub.err:
		return nil, err
	}
}

func (sw *TransactionSubscription) Err() <-chan error {
	return sw.sub.err
}

func (sw *TransactionSubscription) Response() <-chan *TransactionResult {
	typedChan := make(chan *TransactionResult, 1)
	go func(ch chan *TransactionResult) {
		// TODO: will this subscription yield more than one result?
		d, ok := <-sw.sub.stream
		if !ok {
			return
		}
		ch <- d.(*TransactionResult)
	}(typedChan)
	return typedChan
}

func (sw *TransactionSubscription) Unsubscribe() {
	sw.sub.Unsubscribe()
}
