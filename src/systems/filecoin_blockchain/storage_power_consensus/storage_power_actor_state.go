package storage_power_consensus

import (
	"math/big"

	filproofs "github.com/filecoin-project/specs/libraries/filcrypto/filproofs"
	block "github.com/filecoin-project/specs/systems/filecoin_blockchain/struct/block"
	actor "github.com/filecoin-project/specs/systems/filecoin_vm/actor"
	addr "github.com/filecoin-project/specs/systems/filecoin_vm/actor/address"
	util "github.com/filecoin-project/specs/util"
)

func (st *StoragePowerActorState_I) _slashPledgeCollateral(rt Runtime, minerID addr.Address, amount actor.TokenAmount) {
	if amount < 0 {
		rt.Abort("negative amount.")
	}

	currEntry := st._safeGetPowerEntry(rt, minerID)

	amountToSlash := amount

	if currEntry.Impl().LockedPledgeCollateral() < amount {
		amountToSlash = currEntry.Impl().LockedPledgeCollateral_
		currEntry.Impl().LockedPledgeCollateral_ = 0
		// TODO: extra handling of not having enough pledge collateral to be slashed
	} else {
		currEntry.Impl().LockedPledgeCollateral_ = currEntry.LockedPledgeCollateral() - amountToSlash
	}

	st.Impl().PledgeCollateralSlashed_ += amountToSlash
	st.Impl().PowerTable_[minerID] = currEntry

}

// TODO: batch process this if possible
func (st *StoragePowerActorState_I) _lockPledgeCollateral(rt Runtime, address addr.Address, amount actor.TokenAmount) {
	// AvailableBalance -> LockedPledgeCollateral
	if amount < 0 {
		rt.Abort("negative amount.")
	}

	minerID := rt.ToplevelSender()
	currEntry := st._safeGetPowerEntry(rt, minerID)

	if currEntry.Impl().AvailableBalance() < amount {
		rt.Abort("insufficient available balance.")
	}

	currEntry.Impl().AvailableBalance_ = currEntry.AvailableBalance() - amount
	currEntry.Impl().LockedPledgeCollateral_ = currEntry.LockedPledgeCollateral() + amount
	st.Impl().PowerTable_[minerID] = currEntry
}

func (st *StoragePowerActorState_I) _unlockPledgeCollateral(rt Runtime, address addr.Address, amount actor.TokenAmount) {
	// lockedPledgeCollateral -> AvailableBalance
	if amount < 0 {
		rt.Abort("negative amount.")
	}

	minerID := rt.ToplevelSender()

	currEntry := st._safeGetPowerEntry(rt, minerID)
	if currEntry.Impl().LockedPledgeCollateral() < amount {
		rt.Abort("insufficient locked balance.")
	}

	currEntry.Impl().LockedPledgeCollateral_ = currEntry.LockedPledgeCollateral() - amount
	currEntry.Impl().AvailableBalance_ = currEntry.AvailableBalance() + amount
	st.Impl().PowerTable_[minerID] = currEntry

}

func (st *StoragePowerActorState_I) _getPledgeCollateralReq(rt Runtime, power block.StoragePower) actor.TokenAmount {

	// TODO: Implement
	pcRequired := actor.TokenAmount(0)

	return pcRequired
}

// _sampleMinersToSurprise implements the PoSt-Surprise sampling algorithm
func (st *StoragePowerActorState_I) _sampleMinersToSurprise(rt Runtime, challengeCount int, randomness util.Randomness) []addr.Address {
	// this wont quite work -- a.PowerTable() is a HAMT by actor address, doesn't
	// support enumerating by int index. maybe we need that as an interface too,
	// or something similar to an iterator (or iterator over the keys)
	// or even a seeded random call directly in the HAMT: myhamt.GetRandomElement(seed []byte, idx int) using the ticket as a seed

	ptSize := big.NewInt(int64(len(st.PowerTable())))
	allMiners := make([]addr.Address, len(st.PowerTable()))
	index := 0

	for address, _ := range st.PowerTable() {
		allMiners[index] = address
		index++
	}

	sampledMiners := make([]addr.Address, 0)

	for chall := 0; chall < challengeCount; chall++ {
		minerIndex := filproofs.RandomInt(randomness, chall, ptSize)
		panic(minerIndex)
		// hack to turn bigint into int
		minerIndexInt := 0
		potentialChallengee := allMiners[minerIndexInt]
		// call to storage miner actor:
		// if should_challenge(lookupMinerActorStateByAddr(potentialChallengee).ShouldChallenge(rt, SURPRISE_NO_CHALLENGE_PERIOD)){
		// hack below TODO fix
		if true {
			sampledMiners = append(sampledMiners, potentialChallengee)
		}
	}

	return sampledMiners
}

func (st *StoragePowerActorState_I) _safeGetPowerEntry(rt Runtime, minerID addr.Address) PowerTableEntry {
	powerEntry, found := st.PowerTable()[minerID]

	if !found {
		rt.Abort("sm._safeGetPowerEntry: miner not found in power table.")
	}

	return powerEntry
}

func (st *StoragePowerActorState_I) _ensurePledgeCollateralSatisfied(rt Runtime) bool {

	minerID := rt.ToplevelSender()

	powerEntry := st._safeGetPowerEntry(rt, minerID)
	pledgeCollateralRequired := st._getPledgeCollateralReq(rt, powerEntry.ActivePower()+powerEntry.InactivePower())

	if pledgeCollateralRequired < powerEntry.LockedPledgeCollateral() {
		extraLockedFund := powerEntry.LockedPledgeCollateral() - pledgeCollateralRequired
		st._unlockPledgeCollateral(rt, minerID, extraLockedFund)
		return true
	} else if pledgeCollateralRequired < (powerEntry.LockedPledgeCollateral() + powerEntry.AvailableBalance()) {
		fundToLock := pledgeCollateralRequired - powerEntry.LockedPledgeCollateral()
		st._lockPledgeCollateral(rt, minerID, fundToLock)
		return true
	}

	return false
}