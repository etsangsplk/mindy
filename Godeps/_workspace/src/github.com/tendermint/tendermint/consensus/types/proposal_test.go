package consensus

import (
	"testing"

	acm "github.com/eris-ltd/mindy/Godeps/_workspace/src/github.com/tendermint/tendermint/account"
	. "github.com/eris-ltd/mindy/Godeps/_workspace/src/github.com/tendermint/tendermint/common"
	_ "github.com/eris-ltd/mindy/Godeps/_workspace/src/github.com/tendermint/tendermint/config/tendermint_test"
	"github.com/eris-ltd/mindy/Godeps/_workspace/src/github.com/tendermint/tendermint/types"
)

func TestProposalSignable(t *testing.T) {
	proposal := &Proposal{
		Height:           12345,
		Round:            23456,
		BlockPartsHeader: types.PartSetHeader{111, []byte("blockparts")},
		POLRound:         -1,
	}
	signBytes := acm.SignBytes(config.GetString("chain_id"), proposal)
	signStr := string(signBytes)

	expected := Fmt(`{"chain_id":"%s","proposal":{"block_parts_header":{"hash":"626C6F636B7061727473","total":111},"height":12345,"pol_round":-1,"round":23456}}`,
		config.GetString("chain_id"))
	if signStr != expected {
		t.Errorf("Got unexpected sign string for SendTx. Expected:\n%v\nGot:\n%v", expected, signStr)
	}
}
