package chain

import (
	abi "github.com/filecoin-project/specs-actors/actors/abi"
	builtin "github.com/filecoin-project/specs-actors/actors/builtin"
	node_base "github.com/filecoin-project/specs/systems/filecoin_nodes/node_base"
)

// Returns the tipset at or immediately prior to `epoch`.
// For negative epochs, it should return a tipset composed of the genesis block.
func (chain *Chain_I) TipsetAtEpoch(epoch abi.ChainEpoch) Tipset {
	current := chain.HeadTipset()

	for current.Epoch() > epoch && epoch >= genesisEpoch{
		// for epoch <= genesisEpoch, this should return a single-block tipset that includes the genesis block
		current = current.Parents()
	}

	return current
}

// Draws randomness from the tipset at or immediately prior to `epoch`.
func (chain *Chain_I) RandomnessAtEpoch(epoch abi.ChainEpoch) abi.RandomnessSeed {

	ts := chain.TipsetAtEpoch(epoch)
	buffer := []byte{}
	buffer = append(buffer, ts.MinTicket().Digest()...)
	buffer = append(buffer, BigEndianBytesFromInt(int64(epoch))...)
	return blake2b.Sum256(buffer)
}

func (chain *Chain_I) GetTicketProductionRandSeed(epoch abi.ChainEpoch) abi.RandomnessSeed {
	return chain.RandomnessAtEpoch(epoch - node_base.SPC_LOOKBACK_TICKET)
}

func (chain *Chain_I) GetSealRandSeed(epoch abi.ChainEpoch) abi.RandomnessSeed {
	return chain.RandomnessAtEpoch(epoch - builtin.SPC_LOOKBACK_SEAL)
}

func (chain *Chain_I) GetPoStChallengeRandSeed(epoch abi.ChainEpoch) abi.RandomnessSeed {
	return chain.RandomnessAtEpoch(epoch - builtin.SPC_LOOKBACK_POST)
}
